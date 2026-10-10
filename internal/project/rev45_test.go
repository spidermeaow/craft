package project

import (
	"craft/internal/toolchain"
	"strings"
	"testing"
)

func TestRev45BundledDBUsesCraftVersion(t *testing.T) {
	root := t.TempDir()
	base := "name = \"app\"\nversion = \"0.1.0\"\nedition = \"2026\"\nentry = \"main\"\n[dependencies]\n"
	manifest, err := readManifestData(root, []byte(base+"db = { path = \"../db\" }\n"))
	if err != nil || manifest.Versions["db"] != toolchain.Version {
		t.Fatalf("bundled db version: %v, %v", manifest.Versions["db"], err)
	}
	if _, err := readManifestData(root, []byte(base+"db = { path = \"../db\", version = \"0.1.0\" }\n")); err == nil || !strings.Contains(err.Error(), "uses the Craft version") {
		t.Fatalf("explicit bundled version accepted: %v", err)
	}
}
