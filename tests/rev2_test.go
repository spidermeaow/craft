package tests

import (
	"bytes"
	"context"
	"craft/internal/cli"
	"craft/internal/formatter"
	"craft/internal/interpreter"
	"craft/internal/project"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRev2Programs(t *testing.T) {
	cases := []struct{ name, source, want string }{
		{"optional", `func f(n: Int?): Int? { return n } func main(){let n: Int? = f(7); if let v = n {print(v)}; let missing: Int? = null; print(missing == null, n != null); if let v = missing {print(v)} else {print("empty")}}`, "7\ntrue true\nempty\n"},
		{"struct copies", `struct S {var values: Int[];let name: String} func main(){let a: S = S(name: "a",values: [1]);var b: S = a;b.values.append(2);b.values[0]=3;print(a.values,b.values)}`, "[1] [3, 2]\n"},
		{"map insertion order", `func main(){var m: Map<String,Int> = {"a":1,"b":2};m.set("a",3);m.remove("b");m.set("b",4);print(m.keys(),m.values(),m.length,m.containsKey("x"));if let v=m.get("a"){print(v)};print(m.get("x")==null)}`, "[a, b] [3, 4] 2 false\n3\ntrue\n"},
		{"map deep copies", `func main(){let a: Map<String,Int[]> = {"a":[1]};var b: Map<String,Int[]> = a;var v: Int[]=b.require("a");v.append(2);b.set("a",v);print(a.require("a"),b.require("a"))}`, "[1] [1, 2]\n"},
		{"present null map entry", `func main(){let m: Map<String,Int?> = {"a":null};if let v=m.get("a"){print(v==null)};print(m.get("b")==null)}`, "true\ntrue\n"},
		{"optional composites", `struct S{var n:Int} func main(){let n:S?=S(n:1);if let v=n{var copy:S=v;copy.n=2;print(v.n,copy.n)};let a:Int[]?=[];if let v=a{print(v.length)}}`, "1 2\n0\n"},
		{"collection utilities", `func main(){let a:Int[]=[3,1,2];print(a.sorted(),a.reverse(),a.slice(1,3),a.contains(1));if let i=a.indexOf(2){print(i)};print(a.indexOf(9)==null);print("โลกabc".substring(0,3),"aa".replace("a","b"),std.string.join(["a","b"],"/"))}`, "[1, 2, 3] [2, 1, 3] [1, 2] true\n2\ntrue\nโลก bb a/b\n"},
		{"conversions", `func main(){print(std.convert.toInt("-42"),std.convert.toFloat("1.25e1"),std.convert.toBool("false"));print(std.convert.tryInt(" 1")==null,std.convert.tryFloat("NaN")==null);print(std.string.fixed(2.5,0),std.math.round(-2.5),std.math.floor(-2.5),std.math.ceil(-2.5),std.math.pow(2.0,3.0))}`, "-42 12.5 false\ntrue true\n2 -3 -3 -2 8\n"},
		{"json exact values", `func main(){let j:JsonValue=std.json.parse("{\"n\":9007199254740993,\"x\":null}");print(j.asObject().require("n").asInt(),j.asObject().require("x").kind());print(std.json.stringify(std.json.fromObject({"a":std.json.fromInt(7)})))}`, "9007199254740993 null\n{\"a\":7}\n"},
		{"dates", `func main(){let d:DateTime=std.time.parse("2024-02-29T00:00:00Z");let later:DateTime=d.add(std.time.durationMilliseconds(86400000));print(later.format(),later.difference(d).milliseconds());print(std.time.fromUnixSeconds(0).withOffset(420).format())}`, "2024-03-01T00:00:00Z 86400000\n1970-01-01T07:00:00+07:00\n"},
		{"random copies", `func main(){var a:Random=std.math.random(42);var b:Random=a;assert a.nextInt(-9,9)==b.nextInt(-9,9);assert a.nextFloat()==b.nextFloat();print("same")}`, "same\n"},
		{"new catchable errors", `func main(){var n:Int=0;try{std.json.parse("{\"a\":1,\"a\":2}")}catch e{n+=1};try{std.convert.toBool("True")}catch e{n+=1};try{std.math.pow(-1.0,0.5)}catch e{n+=1};try{"x".substring(0,2)}catch e{n+=1};print(n)}`, "4\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := checked(t, tc.source)
			var out bytes.Buffer
			if e := interpreter.Run(context.Background(), p, &out); e != nil {
				t.Fatal(e)
			}
			if out.String() != tc.want {
				t.Fatalf("got %q want %q", out.String(), tc.want)
			}
		})
	}
}

func TestRev2RejectsUnsafePrograms(t *testing.T) {
	for _, source := range []string{
		`func main(){let n:Int=null}`,
		`func main(){let n:Int?=1;print(n+1)}`,
		`func main(){let n:Int?=1;if n!=null{let x:Int=n}}`,
		`func main(){if let n=1{print(n)}}`,
		`func main(){let n:Int?=1;if let v=n{v=2}}`,
		`func main(){let n:Int?=null;if let v=n{};print(v)}`,
		`struct S{var n:Int} func main(){let s:S=S(n:1);s.n=2}`,
		`struct S{let a:Int[]} func main(){var s:S=S(a:[]);s.a.append(1)}`,
		`struct S{var a:Int[]} func f(s:S){s.a.append(1)} func main(){}`,
		`struct S{var n:Int} func main(){let s:S=S(1)}`,
		`struct S{var n:Int} func main(){let s:S=S(wrong:1)}`,
		`struct S{var n:S} func main(){}`,
		`struct S{var n:Void} func main(){}`,
		`struct S{var n:Int;var n:Int} func main(){}`,
		`func main(){let m:Map<String,Int>={};m.set("a",1)}`,
		`func main(){var m:Map<String,Int>={"a":1};m.set("b","bad")}`,
		`func main(){let m:Map<String,Int>={"a":1,"a":2}}`,
		`func main(){var m:Map<Int,Int>={}}`,
		`func main(){let a:Int[]=[];let b:Int[]=[];assert a==b}`,
		`struct S{} func main(){assert S()==S()}`,
		`func main(){let a:Random=std.math.random(1);a.nextFloat()}`,
		`func main(){std.convert.toString([1])}`,
	} {
		p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: source}}}
		if _, _, e := p.Compile(); e == nil {
			t.Errorf("accepted unsafe program: %s", source)
		}
	}
}

func TestRev2InputAndArgumentsInProjectAndBundle(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "input-app")
	if e := project.Init(dir, true); e != nil {
		t.Fatal(e)
	}
	source := `func main(){print(std.env.args());while true{if let line=std.io.readLine(){print("line:",line)}else{break}}}`
	if e := os.WriteFile(filepath.Join(dir, "src/main.craft"), []byte(source), 0644); e != nil {
		t.Fatal(e)
	}
	p, e := project.Load(dir)
	if e != nil {
		t.Fatal(e)
	}
	bundle, e := p.Build()
	if e != nil {
		t.Fatal(e)
	}
	for _, args := range [][]string{{"run", "--", "one", "--flag", "โลก"}, {"run", bundle, "--", "one", "--flag", "โลก"}} {
		var out, stderr bytes.Buffer
		exit := cli.RunWithInput(context.Background(), args, dir, strings.NewReader("โลก\r\n\nlast"), &out, &stderr)
		if exit != 0 || out.String() != "[one, --flag, โลก]\nline: โลก\nline: \nline: last\n" {
			t.Fatalf("exit=%d out=%q error=%s", exit, out.String(), stderr.String())
		}
	}
}

func TestRev2FormatterPreservesMapExpressions(t *testing.T) {
	sources := []string{
		`struct S {var a:Map<String,Int>; let n:String?} func main(){var s:S=S(a:{"x":1},n:null);if let n=s.n{print(n)};print(s.a)}`,
		`func main(){let a:Map<String,Int>[]=[{"x":1},{"y":2}];print(a)}`,
		`func main(){let m:Map<String,Map<String,Int>>={"a":{"b":2}};print(m.require("a").require("b"))}`,
	}
	for _, source := range sources {
		formatted, e := formatter.Format("main.craft", source)
		if e != nil {
			t.Fatal(e)
		}
		twice, e := formatter.Format("main.craft", formatted)
		if e != nil || formatted != twice {
			t.Fatalf("not idempotent: %v\n%s\n%s", e, formatted, twice)
		}
		var before, after bytes.Buffer
		if e = interpreter.Run(context.Background(), checked(t, source), &before); e != nil {
			t.Fatal(e)
		}
		if e = interpreter.Run(context.Background(), checked(t, formatted), &after); e != nil {
			t.Fatal(e)
		}
		if before.String() != after.String() {
			t.Fatal("format changed behavior")
		}
	}
}

func TestRev2ExampleProjects(t *testing.T) {
	for _, name := range []string{"rev2-bank", "rev2-features", "rev2-cli"} {
		p, e := project.LoadTests(filepath.Join("..", "examples", name))
		if e != nil {
			t.Fatal(e)
		}
		program, _, e := p.Compile()
		if e != nil {
			t.Fatal(e)
		}
		for _, test := range program.Tests {
			var out bytes.Buffer
			if e = interpreter.RunTest(context.Background(), program, test, &out); e != nil {
				t.Fatalf("%s/%s: %v", name, test.Name, e)
			}
		}
	}
}
