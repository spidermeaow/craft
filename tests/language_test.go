package tests

import (
	"bytes"
	"context"
	"craft/internal/interpreter"
	"craft/internal/project"
	"strings"
	"testing"
	"time"
)

func TestLanguagePrograms(t *testing.T) {
	cases := []struct{ name, source, want string }{
		{"hello", `func main() { print("Hello World") }`, "Hello World\n"},
		{"arithmetic", `func main() { var n: Int = 10 + 2 * 3; n += 4; n *= 2; n -= 10; n /= 3; n %= 6; print(n, 7 / 2, 7 % 2, -2, +3, 1.5 + 2.5) }`, "4 3 1 -2 3 4\n"},
		{"recursion", `func fact(n: Int): Int { if n <= 1 { return 1 } else { return n * fact(n - 1) } }
func main() { print(fact(6)) }`, "720\n"},
		{"loops", `func main() { var total: Int = 0; for i in 0..8 { if i == 2 { continue }; if i == 5 { break }; total += i }; while total < 10 { total += 1 }; print(total) }`, "10\n"},
		{"short_circuit", `func bomb(): Bool { print(1 / 0); return true }
func main() { print(false && bomb(), true || bomb(), !false) }`, "false true true\n"},
		{"block_scope", `func main() { let x: Int = 1; { let x: String = "inner"; print(x) }; print(x) }`, "inner\n1\n"},
		{"else_if", `func main() { if false { print(1) } else if true { print(2) } else { print(3) } }`, "2\n"},
		{"strings", `func main() { print("hello " + "โลก", "a" < "b", "a" != "b") }`, "hello โลก true true\n"},
		{"range_once", `func main() { var end: Int = 3; for i in 0..end { end = 0; print(i) }; for j in 4..2 { print(j) } }`, "0\n1\n2\n"},
		{"return_loop", `func first(): Int { for i in 0..5 { if i == 2 { return i } }; return -1 }
func main() { print(first()) }`, "2\n"},
		{"limits", `func main() { print(-9223372036854775808, 9223372036854775807) }`, "-9223372036854775808 9223372036854775807\n"},
		{"multiline", "func main()\n{\n print(\n 1 +\n 2\n )\n if true\n { print(4) }\n else\n { print(5) }\n}", "3\n4\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: tc.source}}}
			program, _, e := p.Compile()
			if e != nil {
				t.Fatal(e)
			}
			var out bytes.Buffer
			if e = interpreter.Run(context.Background(), program, &out); e != nil {
				t.Fatal(e)
			}
			if out.String() != tc.want {
				t.Fatalf("got %q, want %q", out.String(), tc.want)
			}
		})
	}
}
func TestStaticErrors(t *testing.T) {
	cases := []struct{ source, want string }{
		{`func main() { let x: Int = "wrong" }`, "cannot assign String to Int"},
		{`func main(x: Int) {}`, "main must have no parameters"},
		{`func main(): Int { return 1 }`, "return Void"},
		{`func other() {}`, "add exactly one func main()"},
		{`func main() {} func main() {}`, "declared more than once"},
		{`func main() { print(missing) }`, "unknown variable"},
		{`func main() { break }`, "only valid inside a loop"},
		{`func main() { continue }`, "only valid inside a loop"},
		{`func main() { if 1 {} }`, "condition must be Bool"},
		{`func main() { for i in 0.0..2 {} }`, "range bounds must be Int"},
		{`func main() { print(1 + 1.5) }`, "cannot combine Int and Float"},
		{`func main() { print(true + false) }`, "cannot combine Bool and Bool"},
		{`func main() { let x: Void = print(1) }`, "invalid variable type"},
		{`func main() { let x: Int = 1; let x: Int = 2 }`, "already declared"},
		{`func main() { { let x: Int = 1 }; print(x) }`, "unknown variable"},
		{`func main() { 1 = 2 }`, "assignment target"},
		{`func f(x: Int): Int { return x } func main() { f() }`, "expects 1 arguments"},
		{`func f(x: Int): Int { return x } func main() { f("x") }`, "argument 1"},
		{`func f(): Int { if true { return 1 } } func main() {}`, "on every path"},
		{`func f(): Int { return "x" } func main() {}`, "return must be Int"},
		{`func main() { print(print(1)) }`, "cannot accept Void"},
		{`func main() { print(9223372036854775808) }`, "outside signed 64-bit"},
		{`func main() { print(1e9999) }`, "outside 64-bit"},
		{`func main() { let print: Int = 1; print(2) }`, "not callable"},
		{`func f(a: Int, a: Int) {} func main() {}`, "duplicate parameter"},
	}
	for _, tc := range cases {
		p := &project.Project{Sources: []project.Source{{Path: "src/bad.craft", Text: tc.source}}}
		_, _, e := p.Compile()
		if e == nil || !strings.Contains(e.Error(), tc.want) {
			t.Errorf("%s: want %q, got %v", tc.source, tc.want, e)
		}
	}
}
func TestRuntimeErrors(t *testing.T) {
	for _, expr := range []string{"1 / 0", "1 % 0", "1.0 / 0.0", "9223372036854775807 + 1", "-9223372036854775808 - 1", "9223372036854775807 * 2", "-(-9223372036854775808)", "-9223372036854775808 / -1", "1e308 * 1e308"} {
		p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: "func main() { print(" + expr + ") }"}}}
		program, _, e := p.Compile()
		if e != nil {
			t.Fatal(e)
		}
		var out bytes.Buffer
		e = interpreter.Run(context.Background(), program, &out)
		if e == nil || !strings.Contains(e.Error(), "E3001") {
			t.Errorf("%s: expected runtime error, got %v", expr, e)
		}
	}
}
func TestCancellation(t *testing.T) {
	p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: "func main() { while true {} }"}}}
	program, _, e := p.Compile()
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	var out bytes.Buffer
	if e = interpreter.Run(ctx, program, &out); e != context.DeadlineExceeded {
		t.Fatalf("expected timeout, got %v", e)
	}
}
func TestMultipleFiles(t *testing.T) {
	p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: "func main() { print(add(2, 3)) }"}, {Path: "src/math.craft", Text: "func add(a: Int, b: Int): Int { return a + b }"}}}
	program, _, e := p.Compile()
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	if e = interpreter.Run(context.Background(), program, &out); e != nil {
		t.Fatal(e)
	}
	if out.String() != "5\n" {
		t.Fatal(out.String())
	}
}
