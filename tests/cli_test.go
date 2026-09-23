package tests

import (
	"bytes"
	"context"
	"craft/internal/cli"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckDoesNotRunAndReportsContext(t *testing.T) {
	dir := t.TempDir()
	var out, stderr bytes.Buffer
	projectDir := filepath.Join(dir, "sample")
	if cli.Run(context.Background(), []string{"new", projectDir}, dir, &out, &stderr) != 0 {
		t.Fatal(stderr.String())
	}
	path := filepath.Join(projectDir, "src", "main.craft")
	if e := os.WriteFile(path, []byte("func main() { print(1 / 0) }"), 0644); e != nil {
		t.Fatal(e)
	}
	out.Reset()
	stderr.Reset()
	if cli.Run(context.Background(), []string{"check"}, projectDir, &out, &stderr) != 0 {
		t.Fatal(stderr.String())
	}
	if out.String() != "Check succeeded.\n" {
		t.Fatal(out.String())
	}
	if e := os.WriteFile(path, []byte("func main() { let x: Int = \"bad\" }"), 0644); e != nil {
		t.Fatal(e)
	}
	out.Reset()
	stderr.Reset()
	if cli.Run(context.Background(), []string{"run"}, projectDir, &out, &stderr) == 0 {
		t.Fatal("invalid program ran")
	}
	if !strings.Contains(stderr.String(), "src/main.craft:1:") || !strings.Contains(stderr.String(), "^") {
		t.Fatal(stderr.String())
	}
}

// This test builds and invokes the real executable when the user runs go test ./....
func TestCLIEndToEnd(t *testing.T) {
	root, e := filepath.Abs("..")
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	binary := filepath.Join(temp, "craft")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/craft")
	build.Dir = root
	if output, e := build.CombinedOutput(); e != nil {
		t.Fatalf("build: %v\n%s", e, output)
	}
	run := func(dir string, args ...string) (string, error) {
		cmd := exec.Command(binary, args...)
		cmd.Dir = dir
		b, e := cmd.CombinedOutput()
		return string(b), e
	}
	if out, e := run(temp, "version"); e != nil || !strings.Contains(out, "Craft 0.1.11") {
		t.Fatalf("version: %v %s", e, out)
	}
	if out, e := run(temp, "new", "hello"); e != nil {
		t.Fatalf("new: %v %s", e, out)
	}
	dir := filepath.Join(temp, "hello")
	if out, e := run(dir, "run"); e != nil || out != "Hello World\n" {
		t.Fatalf("run: %v %q", e, out)
	}
	if out, e := run(dir, "check"); e != nil || !strings.Contains(out, "Check succeeded") {
		t.Fatalf("check: %v %s", e, out)
	}
	if out, e := run(dir, "test"); e != nil || !strings.Contains(out, "1 passed, 0 failed, 1 total") {
		t.Fatalf("test: %v %s", e, out)
	}
	if out, e := run(dir, "fmt"); e != nil {
		t.Fatalf("fmt: %v %s", e, out)
	}
	if out, e := run(dir, "fmt", "--check"); e != nil {
		t.Fatalf("fmt --check: %v %s", e, out)
	}
	if out, e := run(dir, "build", "--release"); e != nil {
		t.Fatalf("build: %v %s", e, out)
	}
	bundle := filepath.Join(dir, "dist", "hello.craftbundle")
	if out, e := run(temp, "run", bundle); e != nil || out != "Hello World\n" {
		t.Fatalf("bundle: %v %q", e, out)
	}
	if out, e := run(dir, "clean"); e != nil {
		t.Fatalf("clean: %v %s", e, out)
	}
	if _, e := os.Stat(bundle); !os.IsNotExist(e) {
		t.Fatal("clean did not remove bundle")
	}
	if out, e := run(temp, "unknown"); e == nil || !strings.Contains(out, "unknown command") {
		t.Fatalf("bad command: %v %s", e, out)
	}
}

func TestRev8PackageCLI(t *testing.T) {
	workspace := t.TempDir()
	lib := filepath.Join(workspace, "lib")
	app := filepath.Join(workspace, "app")
	for _, dir := range []string{lib, app} {
		if e := os.MkdirAll(filepath.Join(dir, "src"), 0755); e != nil {
			t.Fatal(e)
		}
	}
	if e := os.WriteFile(filepath.Join(lib, "craft.toml"), []byte("name = \"lib\"\nversion = \"1.0.0\"\nedition = \"2026\"\nentry = \"library\"\n"), 0644); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(lib, "src", "lib.craft"), []byte("export func value():Int{return 8}"), 0644); e != nil {
		t.Fatal(e)
	}
	manifest := "name = \"app\"\nversion = \"1.0.0\"\nedition = \"2026\"\n[dependencies]\nlib = { path = \"../lib\", version = \"1.0.0\" }\n"
	if e := os.WriteFile(filepath.Join(app, "craft.toml"), []byte(manifest), 0644); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(app, "src", "main.craft"), []byte(`import lib "lib" func main(){print(lib.value())}`), 0644); e != nil {
		t.Fatal(e)
	}
	run := func(args ...string) (int, string, string) {
		var stdout, stderr bytes.Buffer
		code := cli.Run(context.Background(), args, app, &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
	if code, out, stderr := run("package", "check"); code != 0 || !strings.Contains(out, "1 dependencies, unlocked") || stderr != "" {
		t.Fatalf("check: code=%d out=%q err=%q", code, out, stderr)
	}
	if code, out, stderr := run("package", "list"); code != 0 || !strings.Contains(out, "app@1.0.0 (root)") || !strings.Contains(out, "lib@1.0.0 ../lib sha256:") || stderr != "" {
		t.Fatalf("list: code=%d out=%q err=%q", code, out, stderr)
	}
	if code, out, stderr := run("package", "lock"); code != 0 || !strings.Contains(out, "Locked 1 package(s)") || stderr != "" {
		t.Fatalf("lock: code=%d out=%q err=%q", code, out, stderr)
	}
	if code, out, stderr := run("package", "check"); code != 0 || !strings.Contains(out, "1 dependencies, locked") || stderr != "" {
		t.Fatalf("locked check: code=%d out=%q err=%q", code, out, stderr)
	}
	if code, out, stderr := run("run"); code != 0 || out != "8\n" || stderr != "" {
		t.Fatalf("run: code=%d out=%q err=%q", code, out, stderr)
	}
}
