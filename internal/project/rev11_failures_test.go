package project

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRev11IdentityConflictAndFailurePreservesFiles(t *testing.T) {
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	root := t.TempDir()
	newInstallApp(t, root)
	f := newPackageTransport()
	for _, repo := range []string{"dev/one", "dev/two"} {
		f.add(t, repo, "1.0.0", strings.Repeat("a", 40), "same", "export func value(){}", "")
	}
	fixtureInstall(t, root, InstallOptions{Spec: "github.com/dev/one@v1.0.0", Alias: "one"}, f)
	original, _ := os.ReadFile(filepath.Join(root, "craft.toml"))
	lock, _ := os.ReadFile(filepath.Join(root, lockFileName))
	for _, o := range []InstallOptions{{Spec: "github.com/dev/two@v1.0.0", Alias: "one"}, {Spec: "github.com/dev/two@v1.0.0", Alias: "two"}} {
		if _, err := installWithTransport(context.Background(), root, o, f); err == nil {
			t.Fatal("collision accepted")
		}
		current, _ := os.ReadFile(filepath.Join(root, "craft.toml"))
		currentLock, _ := os.ReadFile(filepath.Join(root, lockFileName))
		if !bytes.Equal(original, current) || !bytes.Equal(lock, currentLock) {
			t.Fatal("failed transaction modified project")
		}
	}
	// Cache missing with different bytes for the locked commit must not rewrite the lock.
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	f.add(t, "dev/one", "1.0.0", strings.Repeat("a", 40), "same", "export func changed(){}", "")
	if _, err := installWithTransport(context.Background(), root, InstallOptions{}, f); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("changed archive: %v", err)
	}
}

func TestRev11PrivateExportsAndStructIdentity(t *testing.T) {
	for _, source := range []string{`import a "a" func main(){a.hidden()}`, `import a "a" import b "b" func main(){let value:a.Value=b.value()}`, `import a "a" import a "b" func main(){}`} {
		p := &Project{Modules: []Module{
			{ID: "app@1", Name: "app", Version: "1", Dependencies: map[string]string{"a": "a@1", "b": "b@1"}, Sources: []Source{{Path: "src/main.craft", Text: source}}},
			{ID: "a@1", Name: "a", Version: "1", Sources: []Source{{Path: "src/lib.craft", Text: `export struct Value{let n:Int} func hidden(){} export func value():Value{return Value(n:1)}`}}},
			{ID: "b@1", Name: "b", Version: "1", Sources: []Source{{Path: "src/lib.craft", Text: `export struct Value{let n:Int} export func value():Value{return Value(n:2)}`}}},
		}}
		if _, _, err := p.Compile(); err == nil {
			t.Fatal("namespace safety failed")
		}
	}
}

func TestRev11ArchiveDuplicateAndSize(t *testing.T) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for i := 0; i < 2; i++ {
		tw.WriteHeader(&tar.Header{Name: "root/file", Mode: 0600, Size: 1})
		tw.Write([]byte("x"))
	}
	tw.Close()
	gz.Close()
	if err := extractPackage(context.Background(), b.Bytes(), t.TempDir()); err == nil {
		t.Fatal("duplicate accepted")
	}
	b.Reset()
	gz = gzip.NewWriter(&b)
	tw = tar.NewWriter(gz)
	tw.WriteHeader(&tar.Header{Name: "root/huge", Mode: 0600, Size: 17 << 20})
	tw.Close()
	gz.Close()
	if err := extractPackage(context.Background(), b.Bytes(), t.TempDir()); err == nil {
		t.Fatal("oversized file accepted")
	}
	g := newGitHubTransport()
	u, _ := url.Parse("https://untrusted.example/archive")
	if err := g.client.CheckRedirect(&http.Request{URL: u}, nil); err == nil {
		t.Fatal("untrusted redirect accepted")
	}
}

func TestRev11RecoveryPreservesEdits(t *testing.T) {
	root := t.TempDir()
	newInstallApp(t, root)
	old, _ := os.ReadFile(filepath.Join(root, "craft.toml"))
	tx := installTransaction{OldManifest: old, NewManifest: append(append([]byte{}, old...), []byte("# install\n")...)}
	b, _ := json.Marshal(tx)
	os.WriteFile(filepath.Join(root, installJournal), b, 0600)
	edited := append(append([]byte{}, old...), []byte("# user edit\n")...)
	os.WriteFile(filepath.Join(root, "craft.toml"), edited, 0600)
	if err := recoverInstallation(root); err == nil {
		t.Fatal("overwrote user edits")
	}
	actual, _ := os.ReadFile(filepath.Join(root, "craft.toml"))
	if !bytes.Equal(actual, edited) {
		t.Fatal("edit lost")
	}
}

func TestRev11TransactionInjectedFailure(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		root := t.TempDir()
		newInstallApp(t, root)
		if _, _, e := WritePackageLock(root); e != nil {
			t.Fatal(e)
		}
		old, _ := os.ReadFile(filepath.Join(root, "craft.toml"))
		oldLock, _ := os.ReadFile(filepath.Join(root, lockFileName))
		next := append(append([]byte{}, old...), []byte("# changed\n")...)
		calls := 0
		err := commitInstallationWith(root, old, next, []byte("new lock"), func(p string, b []byte) error {
			calls++
			if calls == failAt {
				return fmt.Errorf("injected write failure")
			}
			return replacePackageFile(p, b)
		})
		if err == nil {
			t.Fatal("failure ignored")
		}
		actual, _ := os.ReadFile(filepath.Join(root, "craft.toml"))
		actualLock, _ := os.ReadFile(filepath.Join(root, lockFileName))
		if !bytes.Equal(old, actual) || !bytes.Equal(oldLock, actualLock) {
			t.Fatal("rollback did not preserve old pair")
		}
		if err = checkInstallJournal(root); err != nil {
			t.Fatal("rollback left journal")
		}
	}
}

func TestRev11ThreeProjectVersionsAndLocalRemoteSubpackage(t *testing.T) {
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	f := newPackageTransport()
	for i, v := range []string{"1.0.0", "1.2.0", "2.0.0"} {
		f.add(t, "dev/versioned", v, strings.Repeat(fmt.Sprint(i+1), 40), "versioned", `export func value():Int{return 1}`, "")
		root := t.TempDir()
		newInstallApp(t, root)
		fixtureInstall(t, root, InstallOptions{Spec: "github.com/dev/versioned@v" + v}, f)
		if _, e := Load(root); e != nil {
			t.Fatal(e)
		}
	}
	commit := strings.Repeat("d", 40)
	f.tags["dev/outer@v1.0.0"] = commit
	f.archives["dev/outer@"+commit] = archiveFixture(t, map[string]string{
		"craft.toml":                   "name='outer'\nversion='1.0.0'\nedition='2026'\nentry='library'\n[dependencies]\ninner={path='packages/inner',version='1.0.0'}",
		"src/lib.craft":                `import inner "inner" export func value():Int{return inner.value()}`,
		"packages/inner/craft.toml":    "name='inner'\nversion='1.0.0'\nedition='2026'\nentry='library'",
		"packages/inner/src/lib.craft": "export func value():Int{return 3}",
	})
	root := t.TempDir()
	newInstallApp(t, root)
	fixtureInstall(t, root, InstallOptions{Spec: "github.com/dev/outer@v1.0.0"}, f)
	if _, e := Load(root); e != nil {
		t.Fatal(e)
	}
}
