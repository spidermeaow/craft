package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const lockFileName = "craft.lock"

// ResolvedPackage is the stable, user-visible result of local dependency
// resolution. Source is relative to the root project whenever possible.
type ResolvedPackage struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Checksum string `json:"checksum"`
}

type projectLock struct {
	Remotes      []RemotePin       `json:"remotes,omitempty"`
	Format       int               `json:"format"`
	Root         string            `json:"root"`
	Dependencies map[string]string `json:"dependencies"`
	Packages     []ResolvedPackage `json:"packages"`
}

func resolvedPackages(root string, modules []Module, paths map[string]string) ([]ResolvedPackage, error) {
	packages := make([]ResolvedPackage, 0, len(modules)-1)
	for i, module := range modules {
		if i == 0 {
			continue
		}
		data, e := json.Marshal(module)
		if e != nil {
			return nil, e
		}
		sum := sha256.Sum256(data)
		source, e := filepath.Rel(root, paths[module.ID])
		if e != nil {
			source = paths[module.ID]
		}
		source = filepath.ToSlash(filepath.Clean(source))
		packages = append(packages, ResolvedPackage{ID: module.ID, Source: source, Checksum: hex.EncodeToString(sum[:])})
	}
	sort.Slice(packages, func(i, j int) bool { return packages[i].ID < packages[j].ID })
	return packages, nil
}

func currentLock(p *Project) projectLock {
	packages := make([]ResolvedPackage, len(p.Packages))
	copy(packages, p.Packages)
	dependencies := map[string]string{}
	if len(p.Modules) > 0 {
		for alias, id := range p.Modules[0].Dependencies {
			dependencies[alias] = id
		}
	}
	format := 1
	if len(p.Remotes) > 0 {
		format = 2
	}
	return projectLock{Format: format, Root: p.Manifest.Name + "@" + p.Manifest.Version, Dependencies: dependencies, Packages: packages, Remotes: p.Remotes}
}

func readProjectLock(path string) (projectLock, error) {
	data, e := readLimited(path, 1024*1024)
	if e != nil {
		return projectLock{}, e
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var lock projectLock
	if e = decoder.Decode(&lock); e != nil {
		return projectLock{}, fmt.Errorf("invalid craft.lock: %w", e)
	}
	if e = decoder.Decode(&struct{}{}); e != io.EOF {
		return projectLock{}, fmt.Errorf("invalid craft.lock: trailing data")
	}
	if (lock.Format != 1 && lock.Format != 2) || lock.Root == "" || lock.Dependencies == nil || lock.Packages == nil {
		return projectLock{}, fmt.Errorf("invalid craft.lock format")
	}
	for alias, id := range lock.Dependencies {
		if !validAlias(alias) || id == "" {
			return projectLock{}, fmt.Errorf("invalid craft.lock dependency %q", alias)
		}
	}
	seen := map[string]bool{}
	if lock.Format == 1 && len(lock.Remotes) > 0 {
		return projectLock{}, fmt.Errorf("remote pins require lock format 2")
	}
	for i, p := range lock.Remotes {
		if p.validate() != nil || !commitPattern.MatchString(p.Commit) || !checksumPattern.MatchString(p.Checksum) || (i > 0 && lock.Remotes[i-1].key() >= p.key()) {
			return projectLock{}, fmt.Errorf("invalid remote lock pin")
		}
	}
	for i, pkg := range lock.Packages {
		if pkg.ID == "" || pkg.Source == "" || len(pkg.Checksum) != 64 || seen[pkg.ID] || (i > 0 && lock.Packages[i-1].ID >= pkg.ID) {
			return projectLock{}, fmt.Errorf("invalid craft.lock package entry %q", pkg.ID)
		}
		if _, e := hex.DecodeString(pkg.Checksum); e != nil {
			return projectLock{}, fmt.Errorf("invalid craft.lock checksum for %s", pkg.ID)
		}
		if strings.HasPrefix(pkg.Source, "github:") {
			valid := false
			for _, pin := range lock.Remotes {
				prefix := "github:" + pin.Repository + "@" + pin.Commit
				if pkg.Source == prefix {
					valid = true
				}
				if suffix, ok := strings.CutPrefix(pkg.Source, prefix+"/"); ok && safeArchivePath(suffix) == nil {
					valid = true
				}
			}
			if !valid {
				return projectLock{}, fmt.Errorf("invalid remote lock source for %s", pkg.ID)
			}
		} else if clean := filepath.ToSlash(filepath.Clean(pkg.Source)); clean != pkg.Source {
			return projectLock{}, fmt.Errorf("invalid craft.lock source for %s", pkg.ID)
		}
		seen[pkg.ID] = true
	}
	return lock, nil
}

func verifyProjectLock(p *Project) (bool, error) {
	path := filepath.Join(p.Root, lockFileName)
	if _, e := os.Lstat(path); os.IsNotExist(e) {
		return false, nil
	} else if e != nil {
		return false, e
	}
	actual, e := readProjectLock(path)
	if e != nil {
		return true, e
	}
	expected := currentLock(p)
	a, _ := json.Marshal(actual)
	b, _ := json.Marshal(expected)
	if !bytes.Equal(a, b) {
		return true, fmt.Errorf("craft.lock is out of date; review changes then run craft package lock (remote tag changes require explicit craft install <source@tag>)")
	}
	return true, nil
}

// LoadPackageGraph resolves a project without requiring main and reports
// whether a matching lockfile exists.
func LoadPackageGraph(start string) (*Project, bool, error) {
	p, e := loadGraphMode(start, false, true)
	if e != nil {
		return nil, false, e
	}
	locked, e := verifyProjectLock(p)
	return p, locked, e
}

// WritePackageLock resolves the current local graph while intentionally
// ignoring a stale lock, then atomically writes the new lock.
func WritePackageLock(start string) (string, int, error) {
	root, err := Find(start)
	if err != nil {
		return "", 0, err
	}
	unlock, err := lockInstallation(root)
	if err != nil {
		return "", 0, err
	}
	defer unlock()
	p, e := loadGraphMode(start, false, false)
	if e != nil {
		return "", 0, e
	}
	lock := currentLock(p)
	data, e := json.MarshalIndent(lock, "", "  ")
	if e != nil {
		return "", 0, e
	}
	data = append(data, '\n')
	path := filepath.Join(p.Root, lockFileName)
	if info, statErr := os.Lstat(path); statErr == nil {
		if !info.Mode().IsRegular() {
			return "", 0, fmt.Errorf("refusing to overwrite non-regular craft.lock")
		}
		if _, e = readProjectLock(path); e != nil {
			return "", 0, fmt.Errorf("refusing to overwrite unrecognized craft.lock: %w", e)
		}
	} else if !os.IsNotExist(statErr) {
		return "", 0, statErr
	}
	f, e := os.CreateTemp(p.Root, ".craft-lock-*")
	if e != nil {
		return "", 0, e
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, e = f.Write(data); e != nil {
		f.Close()
		return "", 0, e
	}
	if e = f.Close(); e != nil {
		return "", 0, e
	}
	if e = os.Rename(temp, path); e != nil {
		return "", 0, e
	}
	return path, len(p.Packages), nil
}

func PackageLines(p *Project) []string {
	lines := []string{p.Manifest.Name + "@" + p.Manifest.Version + " (root)"}
	for _, pkg := range p.Packages {
		lines = append(lines, fmt.Sprintf("%s %s sha256:%s", pkg.ID, pkg.Source, pkg.Checksum[:12]))
	}
	for _, pin := range p.Remotes {
		lines = append(lines, fmt.Sprintf("github %s commit:%s source-sha256:%s", pin.key(), pin.Commit, pin.Checksum))
	}
	return lines
}

func LockPath(root string) string { return filepath.Join(root, lockFileName) }

func hasDependencies(p *Project) bool { return len(p.Packages) > 0 }

func PackageSummary(p *Project, locked bool) string {
	state := "unlocked"
	if locked {
		state = "locked"
	}
	return fmt.Sprintf("Package graph valid: %d dependencies, %s", len(p.Packages), state)
}

func packageHint(p *Project, locked bool) string {
	if hasDependencies(p) && !locked {
		return "Run craft package lock to pin this dependency graph."
	}
	return ""
}

func PackageCheckMessages(p *Project, locked bool) []string {
	messages := []string{PackageSummary(p, locked)}
	if hint := strings.TrimSpace(packageHint(p, locked)); hint != "" {
		messages = append(messages, hint)
	}
	return messages
}
