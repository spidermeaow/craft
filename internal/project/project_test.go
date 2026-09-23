package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectBundleAndClean(t *testing.T) {
	root := filepath.Join(t.TempDir(), "hello")
	if e := Init(root, true); e != nil {
		t.Fatal(e)
	}
	if e := Init(root, false); e == nil {
		t.Fatal("init overwrote existing project")
	}
	p, e := Load(filepath.Join(root, "src"))
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = p.Compile(); e != nil {
		t.Fatal(e)
	}
	path, e := p.Build()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.Build(); e != nil {
		t.Fatalf("rebuild: %v", e)
	}
	bundle, e := LoadBundle(path)
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = bundle.Compile(); e != nil {
		t.Fatal(e)
	}
	keep := filepath.Join(root, "dist", "keep.txt")
	if e = os.WriteFile(keep, []byte("keep"), 0644); e != nil {
		t.Fatal(e)
	}
	if e = Clean(root); e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(path); !os.IsNotExist(e) {
		t.Fatal("bundle not removed")
	}
	if _, e = os.Stat(keep); e != nil {
		t.Fatal("clean removed unrelated data")
	}
}
func TestManifest(t *testing.T) {
	for _, text := range []string{"name = 'hello' # comment\nversion = \"0.1.0\"\nedition = '2026'", "name = \"hello\"\nversion = \"0.1.0\"\nedition = \"2026\"\nentry = \"main\""} {
		root := t.TempDir()
		if e := os.WriteFile(filepath.Join(root, "craft.toml"), []byte(text), 0644); e != nil {
			t.Fatal(e)
		}
		if _, e := ReadManifest(root); e != nil {
			t.Fatal(e)
		}
	}
}
func TestRejectInvalidBundles(t *testing.T) {
	for _, content := range []string{`{}`, `{"format":"craft-source-bundle","version":2,"name":"hello","sources":[]}`, `{"format":"craft-source-bundle","version":1,"name":"hello","sources":[{"path":"../escape.craft","text":""}]}`} {
		path := filepath.Join(t.TempDir(), "bad.craftbundle")
		if e := os.WriteFile(path, []byte(content), 0644); e != nil {
			t.Fatal(e)
		}
		if _, e := LoadBundle(path); e == nil {
			t.Fatal("accepted invalid bundle")
		}
	}
}
func TestCleanProtectsUnrecognizedArtifact(t *testing.T) {
	root := filepath.Join(t.TempDir(), "hello")
	if e := Init(root, true); e != nil {
		t.Fatal(e)
	}
	if e := os.Mkdir(filepath.Join(root, "dist"), 0755); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(root, "dist", "hello.craftbundle")
	if e := os.WriteFile(path, []byte("important"), 0644); e != nil {
		t.Fatal(e)
	}
	if e := Clean(root); e == nil {
		t.Fatal("clean accepted unknown artifact")
	}
	data, e := os.ReadFile(path)
	if e != nil || string(data) != "important" {
		t.Fatal("artifact changed")
	}
}

func TestLegacyBundleRequiresRebuildButCanBeCleaned(t *testing.T) {
	root := filepath.Join(t.TempDir(), "hello")
	if e := Init(root, true); e != nil {
		t.Fatal(e)
	}
	if e := os.Mkdir(filepath.Join(root, "dist"), 0755); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(root, "dist", "hello.craftbundle")
	legacy := []byte(`{"format":"craft-source-bundle","version":1,"name":"hello","sources":[{"path":"src/main.craft","text":"func main() {}"}]}`)
	if e := os.WriteFile(path, legacy, 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := LoadBundle(path); e == nil {
		t.Fatal("legacy bundle executed without migration")
	}
	p, e := Load(root)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.Build(); e != nil {
		t.Fatalf("rebuild legacy: %v", e)
	}
	if _, e = LoadBundle(path); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, legacy, 0644); e != nil {
		t.Fatal(e)
	}
	if e = Clean(root); e != nil {
		t.Fatal(e)
	}
}
