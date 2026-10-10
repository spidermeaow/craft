package project

import (
	"craft/internal/toolchain"
	"strings"
	"testing"
)

func TestRev44BundledAPIUsesCraftVersion(t *testing.T) {
	root := t.TempDir()
	base := "name = \"app\"\nversion = \"0.1.0\"\nedition = \"2026\"\nentry = \"main\"\n[dependencies]\n"
	manifest, err := readManifestData(root, []byte(base+"api = { path = \"../api\" }\n"))
	if err != nil || manifest.Versions["api"] != toolchain.Version {
		t.Fatalf("bundled api version: %v, %v", manifest.Versions["api"], err)
	}
	for _, dependency := range []string{
		base + "api = { path = \"../api\", version = \"0.1.0\" }\n",
		"name = \"app\"\nversion = \"0.1.0\"\nedition = \"2026\"\nentry = \"main\"\ndependency.api = \"../api\"\ndependency-version.api = \"0.1.0\"\n",
	} {
		if _, err := readManifestData(root, []byte(dependency)); err == nil || !strings.Contains(err.Error(), "uses the Craft version") {
			t.Fatalf("explicit bundled version accepted: %s, %v", dependency, err)
		}
	}
}
