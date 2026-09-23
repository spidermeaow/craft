package project

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"craft/internal/interpreter"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type packageFixtureTransport struct {
	tags                map[string]string
	archives            map[string][]byte
	resolves, downloads int
}

func (f *packageFixtureTransport) resolve(ctx context.Context, d GitHubDependency) (string, error) {
	f.resolves++
	if e := ctx.Err(); e != nil {
		return "", e
	}
	v, ok := f.tags[d.key()]
	if !ok {
		return "", fmt.Errorf("missing fixture tag")
	}
	return v, nil
}
func (f *packageFixtureTransport) archive(ctx context.Context, d GitHubDependency, c string) ([]byte, error) {
	f.downloads++
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	v, ok := f.archives[d.Repository+"@"+c]
	if !ok {
		return nil, fmt.Errorf("missing fixture commit")
	}
	return v, nil
}
func archiveFixture(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for name, value := range files {
		if e := tw.WriteHeader(&tar.Header{Name: "root/" + name, Mode: 0600, Size: int64(len(value))}); e != nil {
			t.Fatal(e)
		}
		if _, e := tw.Write([]byte(value)); e != nil {
			t.Fatal(e)
		}
	}
	if e := tw.Close(); e != nil {
		t.Fatal(e)
	}
	if e := gz.Close(); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func newPackageTransport() *packageFixtureTransport {
	return &packageFixtureTransport{tags: map[string]string{}, archives: map[string][]byte{}}
}
func (f *packageFixtureTransport) add(t *testing.T, repo, version, commit, name, source, extra string) {
	d := GitHubDependency{repo, "v" + version, version}
	f.tags[d.key()] = commit
	f.archives[repo+"@"+commit] = archiveFixture(t, map[string]string{"craft.toml": fmt.Sprintf("name = %q\nversion = %q\nedition = '2026'\nentry = 'library'\n%s", name, version, extra), "src/lib.craft": source})
}
func newInstallApp(t *testing.T, root string) {
	t.Helper()
	writePackageFixture(t, root, "name='app'\nversion='1.0.0'\nedition='2026'\nentry='main'\n# preserved comment\n", "func main() {}")
}
func fixtureInstall(t *testing.T, root string, o InstallOptions, f *packageFixtureTransport) {
	t.Helper()
	if _, e := installWithTransport(context.Background(), root, o, f); e != nil {
		t.Fatal(e)
	}
}

func TestRev11InstallRestoreMovedTagAndOffline(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("CRAFT_PACKAGE_CACHE", cache)
	root := t.TempDir()
	newInstallApp(t, root)
	f := newPackageTransport()
	c1 := strings.Repeat("a", 40)
	c2 := strings.Repeat("b", 40)
	f.add(t, "dev/lib", "1.0.0", c1, "my-lib", "export func value():Int{return 1}", "")
	fixtureInstall(t, root, InstallOptions{Spec: "github.com/dev/lib@v1.0.0"}, f)
	before, e := os.ReadFile(filepath.Join(root, "craft.lock"))
	if e != nil {
		t.Fatal(e)
	}
	manifest, _ := os.ReadFile(filepath.Join(root, "craft.toml"))
	if !strings.Contains(string(manifest), "# preserved comment") || !strings.Contains(string(manifest), "my_lib =") {
		t.Fatal("manifest preservation/alias")
	}
	f.add(t, "dev/lib", "1.0.0", c2, "my-lib", "export func value():Int{return 2}", "")
	fixtureInstall(t, root, InstallOptions{}, f)
	fixtureInstall(t, root, InstallOptions{Offline: true}, f)
	after, _ := os.ReadFile(filepath.Join(root, "craft.lock"))
	if !bytes.Equal(before, after) || f.resolves != 1 || f.downloads != 1 {
		t.Fatal("tag moved implicitly")
	}
	if _, e = Load(root); e != nil {
		t.Fatal(e)
	}
	// Restore the exact commit into another cache, without resolving the moved tag.
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	if _, e = installWithTransport(context.Background(), root, InstallOptions{Offline: true}, f); e == nil {
		t.Fatal("offline miss accepted")
	}
	fixtureInstall(t, root, InstallOptions{}, f)
	if f.resolves != 1 || f.downloads != 2 {
		t.Fatal("restore re-resolved tag")
	}
	fixtureInstall(t, root, InstallOptions{Spec: "github.com/dev/lib@v1.0.0", Refresh: true}, f)
	after, _ = os.ReadFile(filepath.Join(root, "craft.lock"))
	if !bytes.Contains(after, []byte(c2)) {
		t.Fatal("explicit refresh ignored")
	}
	lock, e := readProjectLock(filepath.Join(root, "craft.lock"))
	if e != nil {
		t.Fatal(e)
	}
	r, _ := newRemoteResolver(context.Background(), lock, false)
	if e = os.WriteFile(filepath.Join(r.pinPath(lock.Remotes[0]), "README.md"), []byte("tampered"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = Load(root); e == nil || !strings.Contains(e.Error(), "integrity") {
		t.Fatalf("tampered cache: %v", e)
	}
}

func TestRev11TransitiveVersionsAndNamespaces(t *testing.T) {
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	f := newPackageTransport()
	f.add(t, "dev/base", "1.0.0", strings.Repeat("a", 40), "base", "export struct Value{let n:Int} export func value():Value{return Value(n:1)}", "")
	f.add(t, "dev/second", "1.0.0", strings.Repeat("b", 40), "second", "export struct Value{let n:Int} export func value():Value{return Value(n:2)}", "")
	f.add(t, "dev/middle", "1.0.0", strings.Repeat("c", 40), "middle", `import b "b" export func number():Int{return b.value().n}`, "[dependencies]\nb={github='dev/base',tag='v1.0.0',version='1.0.0'}\n")
	root := t.TempDir()
	newInstallApp(t, root)
	for _, spec := range []string{"github.com/dev/base@v1.0.0", "github.com/dev/second@v1.0.0", "github.com/dev/middle@v1.0.0"} {
		fixtureInstall(t, root, InstallOptions{Spec: spec}, f)
	}
	source := `import one "base" import two "second" import mid "middle" func main(){let a:one.Value=one.value();let b:two.Value=two.value();assert a.n==1;assert b.n==2;assert mid.number()==1;print("namespace passed")}`
	os.WriteFile(filepath.Join(root, "src/main.craft"), []byte(source), 0600)
	p, e := Load(root)
	if e != nil {
		t.Fatal(e)
	}
	program, _, e := p.Compile()
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	if e = interpreter.Run(context.Background(), program, &out); e != nil {
		t.Fatal(e)
	}
	if out.String() != "namespace passed\n" {
		t.Fatal(out.String())
	}
	bundle, e := p.Build()
	if e != nil {
		t.Fatal(e)
	}
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	portable, e := LoadBundle(bundle)
	if e != nil {
		t.Fatal(e)
	}
	program, _, e = portable.Compile()
	if e != nil {
		t.Fatal(e)
	}
	if e = interpreter.Run(context.Background(), program, io.Discard); e != nil {
		t.Fatal(e)
	}
	// A second project may choose a different version. A single graph may not.
	f.add(t, "dev/base", "2.0.0", strings.Repeat("d", 40), "base", "export func value():Int{return 20}", "")
	other := t.TempDir()
	newInstallApp(t, other)
	fixtureInstall(t, other, InstallOptions{Spec: "github.com/dev/base@v2.0.0"}, f)
	conflict := t.TempDir()
	newInstallApp(t, conflict)
	fixtureInstall(t, conflict, InstallOptions{Spec: "github.com/dev/middle@v1.0.0"}, f)
	before, _ := os.ReadFile(filepath.Join(conflict, "craft.toml"))
	if _, e = installWithTransport(context.Background(), conflict, InstallOptions{Spec: "github.com/dev/base@v2.0.0"}, f); e == nil || !strings.Contains(e.Error(), "conflict") {
		t.Fatalf("conflict: %v", e)
	}
	after, _ := os.ReadFile(filepath.Join(conflict, "craft.toml"))
	if !bytes.Equal(before, after) {
		t.Fatal("failed install modified manifest")
	}
}

func TestRev11ManifestAndEscaping(t *testing.T) {
	for _, raw := range []string{`{github='dev/lib',tag='main',version='1.0.0'}`, `{github='dev/lib',tag='v1.0.0',version='2.0.0'}`, `{path='../lib',github='dev/lib',tag='v1.0.0',version='1.0.0'}`} {
		_, e := readManifestData(t.TempDir(), []byte("name='app'\nversion='1'\nedition='2026'\n[dependencies]\nlib="+raw))
		if e == nil {
			t.Fatal("invalid manifest accepted")
		}
	}
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	root := t.TempDir()
	newInstallApp(t, root)
	f := newPackageTransport()
	f.add(t, "dev/bad", "1.0.0", strings.Repeat("a", 40), "bad", "export func value(){}", "[dependencies]\nout={path='../../../../',version='1'}")
	if _, e := installWithTransport(context.Background(), root, InstallOptions{Spec: "github.com/dev/bad@v1.0.0"}, f); e == nil {
		t.Fatal("remote escaped")
	}
	f.add(t, "dev/a", "1.0.0", strings.Repeat("b", 40), "a", "export func value(){}", "[dependencies]\nb={github='dev/b',tag='v1.0.0',version='1.0.0'}")
	f.add(t, "dev/b", "1.0.0", strings.Repeat("c", 40), "b", "export func value(){}", "[dependencies]\na={github='dev/a',tag='v1.0.0',version='1.0.0'}")
	if _, e := installWithTransport(context.Background(), root, InstallOptions{Spec: "github.com/dev/a@v1.0.0"}, f); e == nil || !strings.Contains(e.Error(), "cycle") {
		t.Fatalf("cycle: %v", e)
	}
}

func TestRev11ArchiveRejection(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", "root/../../escape", "root/a\\b", "root/NUL.txt", "root/file:stream", "root/trailing.", "root/.git/config"} {
		b := archiveFixture(t, map[string]string{name: "bad"})
		if e := extractPackage(context.Background(), b, t.TempDir()); e == nil {
			t.Fatalf("accepted %s", name)
		}
	}
	for _, kind := range []byte{tar.TypeSymlink, tar.TypeLink, tar.TypeFifo} {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		tw := tar.NewWriter(gz)
		tw.WriteHeader(&tar.Header{Name: "root/link", Typeflag: kind, Linkname: "../../out"})
		tw.Close()
		gz.Close()
		if e := extractPackage(context.Background(), b.Bytes(), t.TempDir()); e == nil {
			t.Fatal("special file accepted")
		}
	}
	b := archiveFixture(t, map[string]string{"A/file": "x", "a/other": "y"})
	if e := extractPackage(context.Background(), b, t.TempDir()); e == nil {
		t.Fatal("case collision accepted")
	}
	b = archiveFixture(t, map[string]string{"src/lib.craft": "abc"})
	if e := extractPackage(context.Background(), b[:len(b)-7], t.TempDir()); e == nil {
		t.Fatal("truncated gzip accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := extractPackage(ctx, b, t.TempDir()); e == nil {
		t.Fatal("cancel ignored")
	}
}

func TestRev11TransactionRecoveryAndExclusion(t *testing.T) {
	for _, stage := range []int{0, 1, 2} {
		root := t.TempDir()
		newInstallApp(t, root)
		old, _ := os.ReadFile(filepath.Join(root, "craft.toml"))
		next := append(append([]byte{}, old...), []byte("# new\n")...)
		tx := installTransaction{OldManifest: old, NewManifest: next, NewLock: []byte("{}")}
		data, _ := json.Marshal(tx)
		os.WriteFile(filepath.Join(root, installJournal), data, 0600)
		if stage >= 1 {
			os.WriteFile(filepath.Join(root, "craft.toml"), next, 0600)
		}
		if stage >= 2 {
			os.WriteFile(filepath.Join(root, lockFileName), tx.NewLock, 0600)
		}
		if _, e := Load(root); e == nil {
			t.Fatal("read during transaction accepted")
		}
		if e := recoverInstallation(root); e != nil {
			t.Fatal(e)
		}
		after, _ := os.ReadFile(filepath.Join(root, "craft.toml"))
		if !bytes.Equal(after, old) {
			t.Fatal("original not restored")
		}
		if _, e := os.Stat(filepath.Join(root, lockFileName)); !os.IsNotExist(e) {
			t.Fatal("partial lock retained")
		}
	}
	root := t.TempDir()
	unlock, e := lockInstallation(root)
	if e != nil {
		t.Fatal(e)
	}
	if second, e := lockInstallation(root); e == nil {
		second()
		t.Fatal("concurrent installer accepted")
	}
	unlock()
	unlock, e = lockInstallation(root)
	if e != nil {
		t.Fatal("lock not released")
	}
	unlock()
}

func TestRev11TransportLimits(t *testing.T) {
	for _, status := range []int{403, 429, 404} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
		g := &githubTransport{client: srv.Client()}
		if _, e := g.get(context.Background(), srv.URL, 32); e == nil {
			t.Fatal("HTTP error ignored")
		}
		srv.Close()
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, strings.Repeat("x", 33)) }))
	defer srv.Close()
	g := &githubTransport{client: srv.Client()}
	if _, e := g.get(context.Background(), srv.URL, 32); e == nil {
		t.Fatal("size ignored")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := g.get(ctx, srv.URL, 32); e == nil {
		t.Fatal("cancellation ignored")
	}
}

func TestRev11ConcurrentCachePromotion(t *testing.T) {
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	base := newPackageTransport()
	base.add(t, "dev/lib", "1.0.0", strings.Repeat("a", 40), "lib", "export func value(){}", "")
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := newRemoteResolver(context.Background(), projectLock{}, true)
			if e != nil {
				t.Error(e)
				return
			}
			r.transport = &packageFixtureTransport{tags: base.tags, archives: base.archives}
			if _, _, e = r.ensure(GitHubDependency{"dev/lib", "v1.0.0", "1.0.0"}); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
}
