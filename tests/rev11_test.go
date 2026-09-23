package tests

import (
	"bytes"
	"context"
	"craft/internal/cli"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRev11LiveGitHub(t *testing.T) {
	if os.Getenv("CRAFT_REV11_LIVE") != "1" {
		t.Skip("set CRAFT_REV11_LIVE=1 for public GitHub acceptance")
	}
	root := t.TempDir()
	app := filepath.Join(root, "consumer")
	t.Setenv("CRAFT_PACKAGE_CACHE", filepath.Join(root, "cache"))
	run := func(dir string, args ...string) string {
		t.Helper()
		var out, err bytes.Buffer
		if code := cli.Run(context.Background(), args, dir, &out, &err); code != 0 {
			t.Fatalf("%v exit %d: %s", args, code, err.String())
		}
		return out.String()
	}
	run(root, "new", app)
	t.Log(run(app, "install", "github.com/spidermeaow/craft-test@v1.0.0", "--alias", "greeting"))
	source := `import greeting "greeting" func main(){assert greeting.isBlank("  ");print(greeting.greet("Craft"))}`
	if e := os.WriteFile(filepath.Join(app, "src", "main.craft"), []byte(source), 0600); e != nil {
		t.Fatal(e)
	}
	run(app, "check")
	run(app, "test")
	if out := run(app, "run"); out != "Hello, Craft!\n" {
		t.Fatal(out)
	}
	run(app, "package", "check")
	t.Log(run(app, "package", "list"))
	run(app, "install", "--offline")
	before, e := os.ReadFile(filepath.Join(app, "craft.lock"))
	if e != nil {
		t.Fatal(e)
	}
	// Fresh cache forces a fetch of the locked commit, never a new tag selection.
	t.Setenv("CRAFT_PACKAGE_CACHE", filepath.Join(root, "second-cache"))
	run(app, "install")
	after, _ := os.ReadFile(filepath.Join(app, "craft.lock"))
	if !bytes.Equal(before, after) {
		t.Fatal("restore changed lock")
	}
	if bytes.Contains(after, []byte(root)) {
		t.Fatal("absolute cache path leaked")
	}
	run(app, "build")
	bundle := filepath.Join(app, "dist", "consumer.craftbundle")
	b, e := os.ReadFile(bundle)
	if e != nil {
		t.Fatal(e)
	}
	var header struct{ Language string }
	if json.Unmarshal(b, &header) != nil || header.Language != "0.1.11" {
		t.Fatal("bundle version")
	}
	t.Setenv("CRAFT_PACKAGE_CACHE", filepath.Join(root, "empty-cache"))
	if out := run(root, "run", bundle); !strings.Contains(out, "Hello, Craft!") {
		t.Fatal(out)
	}
	t.Log("Installed, checked, tested, ran, restored exact commit, and ran portable bundle without cache.")
}
