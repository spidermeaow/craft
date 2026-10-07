package tests

import (
	"bytes"
	"context"
	"craft/internal/cli"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRev12OfflineCacheMissIsProjectError(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join("..", "examples", "rev11-github")
	for _, name := range []string{"craft.toml", "craft.lock"} {
		data, err := os.ReadFile(filepath.Join(source, name))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "src"), 0700); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(source, "src", "main.craft"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "src", "main.craft"), data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CRAFT_PACKAGE_CACHE", filepath.Join(root, "empty-cache"))
	var out, errOut bytes.Buffer
	code := cli.Run(context.Background(), []string{"install", "--offline"}, root, &out, &errOut)
	if code != 4 || !strings.Contains(errOut.String(), "E4002") || !strings.Contains(errOut.String(), "run craft install online") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), errOut.String())
	}
}
