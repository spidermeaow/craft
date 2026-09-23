package project

import (
	"strings"
	"testing"
)

func TestModuleGraphDiagnostics(t *testing.T) {
	for _, tc := range []struct{ name, root, lib, want string }{
		{"private", `import lib "lib" func main(){lib.hidden()}`, `func hidden(){}`, "private"},
		{"missing", `import lib "missing" func main(){}`, `export func value(){}`, "import"},
		{"duplicate", `import a "lib" import b "lib" func main(){}`, `export func value(){}`, "duplicate"},
		{"alias", `import lib "lib" func main(){let lib:Int=1}`, `export func value(){}`, "shadows"},
		{"reserved", `func main(){}`, `export struct HttpRequest{let value:Int}`, "reserved"},
		{"identity", `import lib "lib" struct Value{let n:Int} func main(){let v:Value=lib.value()}`, `export struct Value{let n:Int} export func value():Value{return Value(n:1)}`, "type error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &Project{Modules: []Module{
				{ID: "app@1", Name: "app", Version: "1", Dependencies: map[string]string{"lib": "lib@1"}, Sources: []Source{{Path: "src/main.craft", Text: tc.root}}},
				{ID: "lib@1", Name: "lib", Version: "1", Sources: []Source{{Path: "src/lib.craft", Text: tc.lib}}},
			}}
			if _, _, e := p.Compile(); e == nil || !strings.Contains(e.Error(), tc.want) {
				t.Fatalf("expected %s, got %v", tc.want, e)
			}
		})
	}
	mods := []Module{
		{ID: "a@1", Name: "a", Version: "1", Dependencies: map[string]string{"b": "b@1"}, Sources: []Source{{Path: "src/a.craft", Text: "func main(){}"}}},
		{ID: "b@1", Name: "b", Version: "1", Dependencies: map[string]string{"a": "a@1"}, Sources: []Source{{Path: "src/b.craft", Text: "func value(){}"}}},
	}
	if e := validateModules(mods, false); e == nil || !strings.Contains(e.Error(), "cycle") {
		t.Fatalf("cycle: %v", e)
	}
}
