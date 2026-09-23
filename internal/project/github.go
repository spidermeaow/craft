package project

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type GitHubDependency struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Version    string `json:"version"`
}
type RemotePin struct {
	GitHubDependency
	Commit   string `json:"commit"`
	Checksum string `json:"checksum"`
}

var repositoryPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,38}/[a-z0-9_][a-z0-9_.-]{0,99}$`)
var releaseTagPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var checksumPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func (d GitHubDependency) validate() error {
	if !repositoryPattern.MatchString(d.Repository) || strings.HasSuffix(d.Repository, ".git") || len(d.Version) > 80 || !releaseTagPattern.MatchString(d.Tag) || d.Version != strings.TrimPrefix(d.Tag, "v") {
		return fmt.Errorf("expected owner/repository and exact vMAJOR.MINOR.PATCH tag matching version")
	}
	return nil
}
func (d GitHubDependency) key() string { return d.Repository + "@" + d.Tag }
func parseGitHubSpec(spec string) (GitHubDependency, error) {
	spec = strings.TrimPrefix(spec, "https://")
	if !strings.HasPrefix(spec, "github.com/") {
		return GitHubDependency{}, fmt.Errorf("use github.com/owner/repository@vMAJOR.MINOR.PATCH")
	}
	repo, tag, ok := strings.Cut(strings.TrimPrefix(spec, "github.com/"), "@")
	d := GitHubDependency{strings.ToLower(repo), tag, strings.TrimPrefix(tag, "v")}
	if !ok {
		return d, fmt.Errorf("an exact version tag is required")
	}
	return d, d.validate()
}

type githubTransport struct{ client *http.Client }

func newGitHubTransport() *githubTransport {
	return &githubTransport{client: &http.Client{Timeout: 60 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || req.URL.Scheme != "https" || (req.URL.Host != "api.github.com" && req.URL.Host != "codeload.github.com") || req.URL.User != nil {
			return fmt.Errorf("untrusted package redirect")
		}
		return nil
	}}}
}
func (g *githubTransport) get(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Craft-package-client/0.1.11")
	res, err := g.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("GitHub request failed (network, redirect or timeout)")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		if res.StatusCode == 403 || res.StatusCode == 429 {
			return nil, fmt.Errorf("GitHub access/rate limit: retry later; public repositories only")
		}
		return nil, fmt.Errorf("GitHub HTTP %d; verify public repository and tag", res.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("incomplete GitHub response")
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("GitHub download exceeds limit")
	}
	return b, nil
}
func (g *githubTransport) resolve(ctx context.Context, d GitHubDependency) (string, error) {
	// Resolve the tag namespace explicitly (a branch may have the same name).
	b, err := g.get(ctx, "https://api.github.com/repos/"+d.Repository+"/git/ref/tags/"+d.Tag, 2<<20)
	if err != nil {
		return "", err
	}
	var result struct {
		Ref    string `json:"ref"`
		Object struct {
			SHA  string `json:"sha"`
			Type string `json:"type"`
		} `json:"object"`
	}
	if json.Unmarshal(b, &result) != nil || result.Ref != "refs/tags/"+d.Tag || !commitPattern.MatchString(result.Object.SHA) {
		return "", fmt.Errorf("invalid GitHub commit response")
	}
	for depth := 0; depth < 8; depth++ {
		if result.Object.Type == "commit" {
			return result.Object.SHA, nil
		}
		if result.Object.Type != "tag" {
			return "", fmt.Errorf("tag does not reference a commit")
		}
		b, err = g.get(ctx, "https://api.github.com/repos/"+d.Repository+"/git/tags/"+result.Object.SHA, 2<<20)
		if err != nil {
			return "", err
		}
		result.Object.SHA = ""
		result.Object.Type = ""
		if json.Unmarshal(b, &result) != nil || !commitPattern.MatchString(result.Object.SHA) {
			return "", fmt.Errorf("invalid annotated tag")
		}
	}
	return "", fmt.Errorf("annotated tag chain exceeds limit")
}
func (g *githubTransport) archive(ctx context.Context, d GitHubDependency, commit string) ([]byte, error) {
	return g.get(ctx, "https://codeload.github.com/"+d.Repository+"/tar.gz/"+commit, 32<<20)
}

type packageTransport interface {
	resolve(context.Context, GitHubDependency) (string, error)
	archive(context.Context, GitHubDependency, string) ([]byte, error)
}
type remoteResolver struct {
	ctx       context.Context
	cache     string
	network   bool
	transport packageTransport
	pins      map[string]RemotePin
	used      map[string]RemotePin
	checked   map[string]string
}

func packageCacheRoot() (string, error) {
	if p := os.Getenv("CRAFT_PACKAGE_CACHE"); p != "" {
		return filepath.Abs(p)
	}
	p, e := os.UserCacheDir()
	return filepath.Join(p, "Craft", "packages"), e
}
func newRemoteResolver(ctx context.Context, lock projectLock, network bool) (*remoteResolver, error) {
	cache, e := packageCacheRoot()
	if e != nil {
		return nil, e
	}
	r := &remoteResolver{ctx: ctx, cache: cache, network: network, transport: newGitHubTransport(), pins: map[string]RemotePin{}, used: map[string]RemotePin{}, checked: map[string]string{}}
	for _, p := range lock.Remotes {
		r.pins[p.key()] = p
	}
	return r, nil
}
func (r *remoteResolver) pinPath(p RemotePin) string {
	return filepath.Join(r.cache, filepath.FromSlash(p.Repository), p.Commit, p.Checksum)
}
func (r *remoteResolver) ensure(d GitHubDependency) (string, RemotePin, error) {
	if err := d.validate(); err != nil {
		return "", RemotePin{}, err
	}
	if err := r.ctx.Err(); err != nil {
		return "", RemotePin{}, err
	}
	if p, ok := r.used[d.key()]; ok {
		return r.checked[d.key()], p, nil
	}
	p, locked := r.pins[d.key()]
	if locked {
		dir := r.pinPath(p)
		if _, err := os.Lstat(dir); err == nil {
			sum, err := treeChecksum(dir)
			if err != nil || sum != p.Checksum {
				return "", p, fmt.Errorf("package cache integrity failure for %s; use a fresh CRAFT_PACKAGE_CACHE and restore", d.key())
			}
			r.used[d.key()] = p
			r.checked[d.key()] = dir
			return dir, p, nil
		} else if !os.IsNotExist(err) {
			return "", p, err
		}
	}
	if !r.network {
		return "", p, fmt.Errorf("package %s missing from verified cache/lock; run craft install online", d.key())
	}
	if !locked {
		commit, err := r.transport.resolve(r.ctx, d)
		if err != nil {
			return "", p, err
		}
		p = RemotePin{GitHubDependency: d, Commit: commit}
	}
	if !commitPattern.MatchString(p.Commit) {
		return "", p, fmt.Errorf("invalid commit")
	}
	b, err := r.transport.archive(r.ctx, d, p.Commit)
	if err != nil {
		return "", p, err
	}
	if err = os.MkdirAll(r.cache, 0700); err != nil {
		return "", p, err
	}
	stage, err := os.MkdirTemp(r.cache, ".stage-")
	if err != nil {
		return "", p, err
	}
	defer os.RemoveAll(stage)
	if err = extractPackage(r.ctx, b, stage); err != nil {
		return "", p, err
	}
	sum, err := treeChecksum(stage)
	if err != nil {
		return "", p, err
	}
	if locked && sum != p.Checksum {
		return "", p, fmt.Errorf("download checksum differs from locked source for %s", d.key())
	}
	p.Checksum = sum
	m, err := ReadManifest(stage)
	if err != nil {
		return "", p, err
	}
	if m.Entry != "library" || m.Version != d.Version {
		return "", p, fmt.Errorf("GitHub package must be entry=library and version=%s", d.Version)
	}
	target := r.pinPath(p)
	if err = os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		return "", p, err
	}
	if err = r.ctx.Err(); err != nil {
		return "", p, err
	}
	if err = os.Rename(stage, target); err != nil {
		existing, e := treeChecksum(target)
		if e != nil || existing != sum {
			return "", p, fmt.Errorf("cannot promote package cache: %w", err)
		}
	}
	r.used[d.key()] = p
	r.checked[d.key()] = target
	return target, p, nil
}
func (r *remoteResolver) usedPins() []RemotePin {
	out := []RemotePin{}
	for _, p := range r.used {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].key() < out[j].key() })
	return out
}

func safeArchivePath(name string) error {
	if name == "" || len(name) > 240 || strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "/") || path.Clean(name) != name {
		return fmt.Errorf("unsafe archive path")
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." || part == "." || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") || strings.ContainsAny(part, "<>\"|?*") {
			return fmt.Errorf("unsafe archive component")
		}
		for _, c := range part {
			if c < 32 || c == 127 {
				return fmt.Errorf("control character in archive path")
			}
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || regexp.MustCompile(`^(COM|LPT)[0-9]$`).MatchString(base) || strings.EqualFold(part, ".git") {
			return fmt.Errorf("reserved archive path")
		}
	}
	return nil
}
func extractPackage(ctx context.Context, b []byte, dir string) error {
	if len(b) > 32<<20 {
		return fmt.Errorf("archive exceeds 32 MiB")
	}
	gz, e := gzip.NewReader(bytes.NewReader(b))
	if e != nil {
		return fmt.Errorf("invalid package archive")
	}
	defer gz.Close()
	tr := tar.NewReader(io.LimitReader(gz, 72<<20))
	seen := map[string]bool{}
	spelling := map[string]string{}
	top := ""
	var total int64
	count := 0
	for {
		if e = ctx.Err(); e != nil {
			return e
		}
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("truncated/invalid package archive")
		}
		count++
		if count > 4096 {
			return fmt.Errorf("archive exceeds 4096 entries")
		}
		if h.Typeflag == tar.TypeXGlobalHeader {
			for key := range h.PAXRecords {
				if key != "comment" {
					return fmt.Errorf("unsupported global archive metadata")
				}
			}
			continue
		}
		name := strings.TrimSuffix(h.Name, "/")
		if err = safeArchivePath(name); err != nil {
			return err
		}
		parts := strings.Split(name, "/")
		if top == "" {
			top = parts[0]
		}
		if parts[0] != top {
			return fmt.Errorf("archive must have one root")
		}
		if h.Typeflag != tar.TypeDir && h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA {
			return fmt.Errorf("archive links/special files are unsupported")
		}
		for i := range parts {
			prefix := strings.Join(parts[:i+1], "/")
			key := strings.ToLower(prefix)
			if old, ok := spelling[key]; ok && old != prefix {
				return fmt.Errorf("case-colliding archive paths")
			}
			spelling[key] = prefix
		}
		key := strings.ToLower(name)
		if seen[key] {
			return fmt.Errorf("duplicate archive entry")
		}
		seen[key] = true
		if len(parts) == 1 {
			if h.Typeflag != tar.TypeDir {
				return fmt.Errorf("invalid archive root")
			}
			continue
		}
		dest := filepath.Join(dir, filepath.FromSlash(strings.Join(parts[1:], "/")))
		if h.Typeflag == tar.TypeDir {
			if err = os.MkdirAll(dest, 0700); err != nil {
				return err
			}
			continue
		}
		if h.Size < 0 || h.Size > 16<<20 {
			return fmt.Errorf("archive file exceeds 16 MiB")
		}
		total += h.Size
		if total > 64<<20 {
			return fmt.Errorf("archive exceeds 64 MiB expanded")
		}
		if path.Base(name) == ".gitmodules" {
			return fmt.Errorf("submodules are unsupported")
		}
		if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			return err
		}
		f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		_, copyErr := io.CopyN(f, tr, h.Size)
		closeErr := f.Close()
		if copyErr != nil {
			return fmt.Errorf("truncated archive file")
		}
		if closeErr != nil {
			return closeErr
		}
	}
	// Drain the gzip stream to validate its checksum/footer, with a bounded tail.
	if n, err := io.Copy(io.Discard, io.LimitReader(gz, 1<<20)); err != nil || n >= 1<<20 {
		return fmt.Errorf("invalid archive trailer")
	}
	if count == 0 {
		return fmt.Errorf("empty archive")
	}
	return nil
}
func treeChecksum(root string) (string, error) {
	info, e := os.Lstat(root)
	if e != nil {
		return "", e
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("invalid cache directory")
	}
	type entry struct {
		Path string `json:"path"`
		Hash string `json:"sha256"`
	}
	entries := []entry{}
	var total int64
	count := 0
	e = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		count++
		if count > 4097 {
			return fmt.Errorf("cache entry limit")
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("cache symlink rejected")
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("cache special file rejected")
		}
		b, err := readLimited(p, 16<<20)
		if err != nil {
			return err
		}
		total += int64(len(b))
		if total > 64<<20 {
			return fmt.Errorf("cache size limit")
		}
		if bytes.HasPrefix(b, []byte("version https://git-lfs.github.com/spec/v1")) {
			return fmt.Errorf("Git LFS is unsupported")
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		entries = append(entries, entry{filepath.ToSlash(rel), hex.EncodeToString(sum[:])})
		return nil
	})
	if e != nil {
		return "", e
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	data, _ := json.Marshal(entries)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
