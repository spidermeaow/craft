package parser

import (
	"craft/internal/lexer"
	"testing"
)

func TestPrecedenceAndRightAssociativeAssignment(t *testing.T) {
	tokens, e := lexer.Scan("main.craft", "func main() {\n var a: Int = 1 + 2 * 3\n a = a = 9\n}")
	if e != nil {
		t.Fatal(e)
	}
	p, e := Parse(tokens)
	if e != nil {
		t.Fatal(e)
	}
	init := p.Functions[0].Body.Statements[0].Expr
	if init.Text != "+" || init.Right.Text != "*" {
		t.Fatalf("wrong arithmetic tree: %+v", init)
	}
	assign := p.Functions[0].Body.Statements[1].Expr
	if assign.Kind != "assignment" || assign.Right.Kind != "assignment" {
		t.Fatalf("wrong assignment tree: %+v", assign)
	}
}
func TestMalformedPrograms(t *testing.T) {
	for _, s := range []string{"func main() {", "func main( {}", "func main() { print(1 2) }", "func main() { let x: Int = 1 let y: Int = 2 }", "let x: Int = 1"} {
		tokens, e := lexer.Scan("bad.craft", s)
		if e != nil {
			continue
		}
		if _, e = Parse(tokens); e == nil {
			t.Errorf("accepted %q", s)
		}
	}
}
func FuzzParse(f *testing.F) {
	for _, s := range []string{"func main() {}", "func main() { print(1 + 2) }", "func main() { if true {} else {} }"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 65536 {
			t.Skip()
		}
		tokens, e := lexer.Scan("fuzz.craft", s)
		if e == nil {
			_, _ = Parse(tokens)
		}
	})
}
