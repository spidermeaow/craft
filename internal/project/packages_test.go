package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePackageFixture(t *testing.T, root, manifest, source string) {
	t.Helper()
	if e := os.MkdirAll(filepath.Join(root, "src"), 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(root, "craft.toml"), []byte(manifest), 0644); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(root, "src", "main.craft"), []byte(source), 0644); e != nil {
		t.Fatal(e)
	}
}

func TestRev8ManifestDependenciesAndLegacyCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, dependency string
	}{
		{"table", "[dependencies]\nlib = { path = '../lib', version = \"1.2.3\" } # local"},
		{"legacy", "dependency.lib = '../lib'\ndependency-version.lib = '1.2.3'"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			manifest := "name = \"app\"\nversion = \"1.0.0\"\nedition = \"2026\"\n" + tc.dependency + "\n"
			if e := os.WriteFile(filepath.Join(root, "craft.toml"), []byte(manifest), 0644); e != nil {
				t.Fatal(e)
			}
			m, e := ReadManifest(root)
			if e != nil {
				t.Fatal(e)
			}
			if m.Dependencies["lib"] != "../lib" || m.Versions["lib"] != "1.2.3" || m.DependencyPositions["lib"].Line == 0 {
				t.Fatalf("unexpected dependency: %#v", m)
			}
		})
	}
}

func TestRev8ManifestRejectsMixedDependencyForms(t *testing.T) {
	root := t.TempDir()
	manifest := "name = \"app\"\nversion = \"1.0.0\"\nedition = \"2026\"\ndependency-version.lib = \"2.0.0\"\n[dependencies] # canonical\nlib = { path = \"../lib\", version = \"1.0.0\" }\n"
	if e := os.WriteFile(filepath.Join(root, "craft.toml"), []byte(manifest), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := ReadManifest(root); e == nil || !strings.Contains(e.Error(), "duplicate dependency version lib") {
		t.Fatalf("mixed dependency forms accepted: %v", e)
	}
}

func TestRev8PackageLockDetectsSourceChanges(t *testing.T) {
	workspace := t.TempDir()
	lib := filepath.Join(workspace, "lib")
	app := filepath.Join(workspace, "app")
	writePackageFixture(t, lib, "name = \"lib\"\nversion = \"1.2.3\"\nedition = \"2026\"\nentry = \"library\"\n", "export func value():Int{return 1}")
	writePackageFixture(t, app, "name = \"app\"\nversion = \"1.0.0\"\nedition = \"2026\"\n[dependencies]\nlib = { path = \"../lib\", version = \"1.2.3\" }\n", `import lib "lib" func main(){print(lib.value())}`)

	p, e := Load(app)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Packages) != 1 || p.Packages[0].ID != "lib@1.2.3" || p.Packages[0].Source != "../lib" {
		t.Fatalf("unexpected resolution: %#v", p.Packages)
	}
	path, count, e := WritePackageLock(app)
	if e != nil || count != 1 || filepath.Base(path) != "craft.lock" {
		t.Fatalf("lock: %s %d %v", path, count, e)
	}
	if _, locked, e := LoadPackageGraph(app); e != nil || !locked {
		t.Fatalf("locked graph: locked=%v error=%v", locked, e)
	}
	if e = os.WriteFile(filepath.Join(lib, "src", "main.craft"), []byte("export func value():Int{return 2}"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = Load(app); e == nil || !strings.Contains(e.Error(), "craft.lock is out of date") {
		t.Fatalf("changed package accepted: %v", e)
	}
	if _, _, e = WritePackageLock(app); e != nil {
		t.Fatalf("refresh lock: %v", e)
	}
	if _, e = Load(app); e != nil {
		t.Fatalf("refreshed graph: %v", e)
	}
}

func TestRev8PackageDiagnostics(t *testing.T) {
	root := t.TempDir()
	writePackageFixture(t, root, "name = \"app\"\nversion = \"1.0.0\"\nedition = \"2026\"\n[dependencies]\nlib = { path = \"../missing\", version = \"1.0.0\" }\n", "func main(){}")
	_, e := Load(root)
	if e == nil || !strings.Contains(e.Error(), "craft.toml:5:1") || !strings.Contains(e.Error(), "dependency lib") || !strings.Contains(e.Error(), "missing") {
		t.Fatalf("unexpected diagnostic: %v", e)
	}
}

func TestRev8TransitivePackagesAreDeterministic(t *testing.T) {
	workspace := t.TempDir()
	base := filepath.Join(workspace, "base")
	middle := filepath.Join(workspace, "middle")
	app := filepath.Join(workspace, "app")
	writePackageFixture(t, base, "name = \"base\"\nversion = \"1.0.0\"\nedition = \"2026\"\nentry = \"library\"\n", "export func value():Int{return 7}")
	writePackageFixture(t, middle, "name = \"middle\"\nversion = \"1.0.0\"\nedition = \"2026\"\nentry = \"library\"\n[dependencies]\nbase = { path = \"../base\", version = \"1.0.0\" }\n", `import base "base" export func value():Int{return base.value()}`)
	writePackageFixture(t, app, "name = \"app\"\nversion = \"1.0.0\"\nedition = \"2026\"\n[dependencies]\nmiddle = { path = \"../middle\", version = \"1.0.0\" }\n", `import middle "middle" func main(){print(middle.value())}`)

	if _, _, e := WritePackageLock(app); e != nil {
		t.Fatal(e)
	}
	first, e := os.ReadFile(filepath.Join(app, "craft.lock"))
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = WritePackageLock(app); e != nil {
		t.Fatal(e)
	}
	second, e := os.ReadFile(filepath.Join(app, "craft.lock"))
	if e != nil {
		t.Fatal(e)
	}
	if string(first) != string(second) || !strings.Contains(string(first), "base@1.0.0") || !strings.Contains(string(first), "middle@1.0.0") {
		t.Fatalf("lock is not deterministic:\n%s\n%s", first, second)
	}
	p, e := Load(app)
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = p.Compile(); e != nil {
		t.Fatal(e)
	}
	bundlePath, e := p.Build()
	if e != nil {
		t.Fatal(e)
	}
	firstBundle, e := os.ReadFile(bundlePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.Build(); e != nil {
		t.Fatal(e)
	}
	secondBundle, e := os.ReadFile(bundlePath)
	if e != nil {
		t.Fatal(e)
	}
	if string(firstBundle) != string(secondBundle) {
		t.Fatal("bundle changed without input changes")
	}
}

func TestRev8DependencyCycleReportsManifestLocation(t *testing.T) {
	workspace := t.TempDir()
	a := filepath.Join(workspace, "a")
	b := filepath.Join(workspace, "b")
	writePackageFixture(t, a, "name = \"a\"\nversion = \"1.0.0\"\nedition = \"2026\"\nentry = \"library\"\n[dependencies]\nb = { path = \"../b\", version = \"1.0.0\" }\n", "export func a(){}")
	writePackageFixture(t, b, "name = \"b\"\nversion = \"1.0.0\"\nedition = \"2026\"\nentry = \"library\"\n[dependencies]\na = { path = \"../a\", version = \"1.0.0\" }\n", "export func b(){}")
	_, e := Load(a)
	if e == nil || !strings.Contains(e.Error(), "craft.toml:6:1") || !strings.Contains(e.Error(), "dependency cycle: a@1.0.0 -> b@1.0.0 -> a@1.0.0") {
		t.Fatalf("unexpected cycle diagnostic: %v", e)
	}
}
