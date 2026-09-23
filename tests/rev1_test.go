package tests

import (
	"bytes"
	"context"
	"craft/internal/ast"
	"craft/internal/cli"
	"craft/internal/diagnostics"
	"craft/internal/interpreter"
	"craft/internal/project"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func checked(t *testing.T, source string) *ast.Program {
	t.Helper()
	p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: source}}}
	program, _, e := p.Compile()
	if e != nil {
		t.Fatal(e)
	}
	return program
}
func TestRev1Behavior(t *testing.T) {
	cases := []struct{ name, source, want string }{
		{"mutable", `func main() { var n: Int = 0; while n < 3 { n += 1 }; print(n) }`, "3\n"},
		{"array", `func main() { var a: Int[] = []; a.append(2); a.append(3); a[0] += 8; print(a.remove(1), a.length, a[0]) }`, "3 1 10\n"},
		{"array_copy", `func main() { let a: Int[][] = [[1]]; var b: Int[][] = a; b[0][0] = 9; b[0].append(2); print(a, b) }`, "[[1]] [[9, 2]]\n"},
		{"argument_copy", `func change(a: Int[]): Int[] { var b: Int[] = a; b.append(2); return b } func main() { let a: Int[] = [1]; print(change(a), a) }`, "[1, 2] [1]\n"},
		{"iteration_snapshot", `func main() { var a: Int[] = [1, 2]; for n in a { a.append(n); print(n) }; print(a) }`, "1\n2\n[1, 2, 1, 2]\n"},
		{"string", `func main() { let s: String = " Craft "; print(s.trim().upper(), s.lower().contains("craft"), s.startsWith(" "), s.endsWith(" "), "โลก".length); print("a,b".split(",")) }`, "CRAFT true true true 3\n[a, b]\n"},
		{"catch", `func fail(): Int { throw Exception("oops") } func main() { try { print(fail()) } catch error { print(error.message, error.type, error.code) }; print("continued") }`, "oops Exception E3002\ncontinued\n"},
		{"runtime_catch", `func main() { let a: Int[] = [1]; try { print(a[9]) } catch e { print(e.type, e.code) }; try { print(1 / 0) } catch e { print(e.code) } }`, "RuntimeError E3001\nE3001\n"},
		{"defer_lifo", `func value(): Int { var n: Int = 1; defer { print("first", n) }; defer { n = 2; print("second") }; return n } func main() { print(value()) }`, "second\nfirst 2\n1\n"},
		{"defer_exception", `func fail() { defer { print("cleanup") }; throw Exception("oops") } func main() { try { fail() } catch e { print(e.message) } }`, "cleanup\noops\n"},
		{"defer_binding", `func main() { var x: Int = 1; { defer { print(x) }; let x: String = "shadow"; print(x) }; x = 2 }`, "shadow\n2\n"},
		{"defer_return_snapshot", `func values(): Int[] { var a: Int[] = [1]; defer { a[0] = 9 }; return a } func main() { print(values()) }`, "[1]\n"},
		{"cleanup_errors", `func fail() { defer { print("last cleanup") }; defer { throw Exception("cleanup") }; throw Exception("original") } func main() { try { fail() } catch e { print(e.message) } }`, "last cleanup\noriginal\n"},
		{"catch_return", `func value(): Int { try { throw Exception("bad") } catch e { return 8 } } func main() { print(value()) }`, "8\n"},
		{"named", `func diff(a: Int, b: Int): Int { return a - b } func value(n: Int): Int { print(n); return n } func main() { print(diff(b: value(2), a: value(5))) }`, "2\n5\n3\n"},
		{"stdlib", `func main() { print(std.math.sqrt(9.0), std.math.abs(-2.0), std.json.isValid("{\"a\":1}"), std.json.isValid("bad")); print(std.json.quote("Craft")) }`, "3 2 true false\n\"Craft\"\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := checked(t, tc.source)
			var out bytes.Buffer
			if e := interpreter.Run(context.Background(), p, &out); e != nil {
				t.Fatal(e)
			}
			if out.String() != tc.want {
				t.Fatalf("got %q, want %q", out.String(), tc.want)
			}
		})
	}
}
func TestRev1StaticSafety(t *testing.T) {
	cases := []struct{ source, code string }{
		{`func main() { let n: Int = 1; n += 1 }`, "E2003"},
		{`func main() { let a: Int[] = [1]; a[0] = 2 }`, "E2003"},
		{`func main() { let a: Int[][] = [[1]]; a[0].append(2) }`, "E2003"},
		{`func main() { let a: Int[] = []; a.append(1) }`, "E2003"},
		{`func change(a: Int[]) { a.append(1) } func main() {}`, "E2003"},
		{`func main() { for n in 0..3 { n += 1 } }`, "E2003"},
		{`func main() { var x: Int = 1; x = "bad" }`, "E2002"},
		{`func main() { print(x); let x: Int = 1 }`, "E2001"},
		{`func main() { let a: Int[] = [1, "bad"] }`, "E2002"},
		{`func main() { let a: Int[] = [1]; print(a["x"]) }`, "E2002"},
		{`func main() { throw "bad" }`, "E2002"},
		{`func main() { try {} }`, "E1001"},
		{`func main() { defer { return } }`, "E2002"},
		{`func main() { while true { defer { break }; break } }`, "E2002"},
		{`func main() { defer { defer {} } }`, "E2002"},
		{`func f(): Int { try { return 1 } catch e { print(e.message) } } func main() {}`, "E2002"},
		{`func f(): Int { while true { break; return 1 } } func main() {}`, "E2002"},
		{`func main() { assert 1 }`, "E2002"},
		{`func f(a: Int, b: Int) {} func main() { f(a: 1, a: 2) }`, "E2002"},
		{`func f(a: Int, b: Int) {} func main() { f(1, b: 2) }`, "E2002"},
		{`func f(a: Int) {} func main() { f(b: 2) }`, "E2002"},
		{`func main() {} test "bad" { return }`, "E2002"},
	}
	for _, tc := range cases {
		p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: tc.source}}}
		_, _, e := p.Compile()
		var d *diagnostics.Error
		if !errors.As(e, &d) || d.Code != tc.code {
			t.Errorf("%s: want %s, got %v", tc.source, tc.code, e)
		}
	}
}
func TestStackTraceAcrossFilesAndRethrow(t *testing.T) {
	p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: `func main() { middle() }`}, {Path: "src/math.craft", Text: `func middle() { try { divide() } catch e { throw e } }
func divide() { print(1 / 0) }`}}}
	program, sources, e := p.Compile()
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	e = interpreter.Run(context.Background(), program, &out)
	var d *diagnostics.Error
	if !errors.As(e, &d) {
		t.Fatalf("no diagnostic: %v", e)
	}
	if len(d.Frames) != 3 || d.Frames[0].Name != "divide" || d.Frames[1].Name != "middle" || d.Frames[2].Name != "main" {
		t.Fatalf("bad stack: %+v", d.Frames)
	}
	if d.Pos.File != "src/math.craft" || d.Pos.Line != 2 {
		t.Fatalf("bad origin: %+v", d.Pos)
	}
	formatted := diagnostics.Format(e, sources)
	if !strings.Contains(formatted, "at main (src/main.craft:1:") || !strings.Contains(formatted, "^") {
		t.Fatal(formatted)
	}
}
func TestRev1CLIExitCodesAndTestDiscovery(t *testing.T) {
	root := filepath.Join(t.TempDir(), "hello")
	if e := project.Init(root, true); e != nil {
		t.Fatal(e)
	}
	invoke := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		code := cli.Run(context.Background(), args, root, &out, &errOut)
		return code, out.String(), errOut.String()
	}
	path := filepath.Join(root, "src", "main.craft")
	for _, tc := range []struct {
		src  string
		code int
	}{{"func main() {", 2}, {`func main() { let x: Int = "bad" }`, 3}, {`func main() { print(1 / 0) }`, 1}, {`func other() {}`, 4}} {
		if e := os.WriteFile(path, []byte(tc.src), 0644); e != nil {
			t.Fatal(e)
		}
		if got, _, errOut := invoke("run"); got != tc.code {
			t.Fatalf("exit %d, want %d: %s", got, tc.code, errOut)
		}
	}
	if got, _, _ := invoke("unknown"); got != 5 {
		t.Fatal(got)
	}
	if e := os.WriteFile(path, []byte(`func main() { throw Exception("main must not run") } func add(a: Int, b: Int): Int { return a + b }`), 0644); e != nil {
		t.Fatal(e)
	}
	testFile := filepath.Join(root, "tests", "math.craft")
	if e := os.WriteFile(testFile, []byte(`test "addition" { assert add(2, 3) == 5 } test "failure" { assert false } test "after failure" { assert true }`), 0644); e != nil {
		t.Fatal(e)
	}
	got, out, errOut := invoke("test")
	if got != 1 || !strings.Contains(out, "3 passed, 1 failed, 4 total") || !strings.Contains(out, "PASS after failure") || !strings.Contains(errOut, "E3003") || !strings.Contains(errOut, "tests/math.craft") {
		t.Fatalf("%d %s %s", got, out, errOut)
	}
	if e := os.WriteFile(testFile, []byte(`test "addition" { assert add(2, 3) == 5 }`), 0644); e != nil {
		t.Fatal(e)
	}
	if got, out, errOut = invoke("test"); got != 0 || !strings.Contains(out, "2 passed, 0 failed") {
		t.Fatalf("%d %s %s", got, out, errOut)
	}
}
func TestFmtCLILeavesAllFilesUntouchedOnSyntaxError(t *testing.T) {
	root := filepath.Join(t.TempDir(), "hello")
	if e := project.Init(root, true); e != nil {
		t.Fatal(e)
	}
	main := filepath.Join(root, "src", "main.craft")
	original := `func main(){print("hello")}`
	if e := os.WriteFile(main, []byte(original), 0644); e != nil {
		t.Fatal(e)
	}
	bad := filepath.Join(root, "src", "z.craft")
	if e := os.WriteFile(bad, []byte("func broken("), 0644); e != nil {
		t.Fatal(e)
	}
	var out, errOut bytes.Buffer
	if code := cli.Run(context.Background(), []string{"fmt"}, root, &out, &errOut); code != 2 {
		t.Fatalf("%d %s", code, errOut.String())
	}
	data, e := os.ReadFile(main)
	if e != nil || string(data) != original {
		t.Fatal("fmt partially changed source before detecting syntax error")
	}
}
