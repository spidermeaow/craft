package parser

import (
	"craft/internal/lexer"
	"testing"
)

func TestRev41MultilineAndTrailingCommas(t *testing.T) {
	source := "func total(\n a: Int,\n b: Int,\n): Int {\n return a +\n b\n}\nfunc main() {\n let xs: Int[] = [\n 1,\n 2,\n ]\n let value: Int = total(\n xs[0],\n xs[1],\n )\n assert value == 3 &&\n xs[1] == 2\n}\n"
	tokens, err := lexer.Scan("multiline.craft", source)
	if err != nil {
		t.Fatal(err)
	}
	program, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Functions) != 2 || len(program.Functions[0].Params) != 2 {
		t.Fatalf("unexpected program: %+v", program)
	}
}

func TestRev41StatementBoundaryAndModuleImports(t *testing.T) {
	valid := "module pages.excel\nimport module data \"models.sheet\"\nimport package gui \"gui\"\nexport func value(): Int { return data.answer() }\n"
	tokens, err := lexer.Scan("module.craft", valid)
	if err != nil {
		t.Fatal(err)
	}
	program, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if program.Module != "pages.excel" || len(program.Imports) != 2 || program.Imports[0].Kind != "module" || program.Imports[1].Kind != "package" {
		t.Fatalf("unexpected imports: %+v", program)
	}
	invalid := "func main() { let a: Int = 1\n + 2 }"
	tokens, err = lexer.Scan("boundary.craft", invalid)
	if err != nil {
		t.Fatal(err)
	}
	boundary, err := Parse(tokens)
	if err != nil || len(boundary.Functions[0].Body.Statements) != 2 {
		t.Fatalf("newline before operator changed statement boundary: %v", err)
	}
}

func TestRev41MultilineGenericType(t *testing.T) {
	source := "func use(f: Fn<\n Int,\n Int,\n>): Int { return f(2) }\n"
	tokens, err := lexer.Scan("generic.craft", source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(tokens); err != nil {
		t.Fatal(err)
	}
}

func TestRev41MemberCallWithOpeningParenOnNextLine(t *testing.T) {
	source := "func main() {\n var fields: Int[] = []\n fields.append\n (\n  1,\n )\n}\n"
	tokens, err := lexer.Scan("member-call.craft", source)
	if err != nil {
		t.Fatal(err)
	}
	program, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	call := program.Functions[0].Body.Statements[1].Expr
	if call.Kind != "call" || call.Left.Kind != "member" || call.Left.Text != "append" || len(call.Args) != 1 {
		t.Fatalf("expected append call, got %+v", call)
	}
}
