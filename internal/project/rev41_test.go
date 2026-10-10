package project

import (
	"strings"
	"testing"
)

func TestRev41ProjectModulesAcrossFiles(t *testing.T) {
	p := &Project{Modules: []Module{{
		ID: "app@1", Name: "app", Version: "1", Dependencies: map[string]string{}, Sources: []Source{
			{Path: "src/main.craft", Text: "import module sheet \"pages.excel\"\nfunc main() { assert sheet.value() == 3 }"},
			{Path: "src/pages/one.craft", Text: "module pages.excel\nfunc helper(): Int { return 3 }"},
			{Path: "src/pages/two.craft", Text: "module pages.excel\nexport func value(): Int { return helper() }"},
		},
	}}}
	if _, _, err := p.Compile(); err != nil {
		t.Fatal(err)
	}
}

func TestRev41ProjectModuleVisibilityAndCycles(t *testing.T) {
	for _, tc := range []struct {
		name    string
		sources []Source
		want    string
	}{
		{"private", []Source{{Path: "src/main.craft", Text: "import module sheet \"pages.excel\"\nfunc main() { sheet.hidden() }"}, {Path: "src/sheet.craft", Text: "module pages.excel\nfunc hidden() {}"}}, "private"},
		{"cycle", []Source{{Path: "src/main.craft", Text: "func main() {}"}, {Path: "src/a.craft", Text: "module a\nimport module b \"b\"\nexport func one() {}"}, {Path: "src/b.craft", Text: "module b\nimport module a \"a\"\nexport func two() {}"}}, "cycle"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &Project{Modules: []Module{{ID: "app@1", Name: "app", Version: "1", Sources: tc.sources}}}
			_, _, err := p.Compile()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %s, got %v", tc.want, err)
			}
		})
	}
}

func TestRev41ExplicitPackageAndProjectImports(t *testing.T) {
	p := &Project{Modules: []Module{
		{ID: "app@1", Name: "app", Version: "1", Dependencies: map[string]string{"lib": "lib@1"}, Sources: []Source{
			{Path: "src/main.craft", Text: "import package external \"lib\"\nimport module internal \"calc\"\nfunc main() { assert internal.value() == external.value() }"},
			{Path: "src/calc.craft", Text: "module calc\nexport func value(): Int { return 7 }"},
		}},
		{ID: "lib@1", Name: "lib", Version: "1", Sources: []Source{{Path: "src/lib.craft", Text: "export func value(): Int { return 7 }"}}},
	}}
	if _, _, err := p.Compile(); err != nil {
		t.Fatal(err)
	}
}
