package tests

import (
	"bytes"
	"context"
	"craft/internal/formatter"
	"craft/internal/interpreter"
	"craft/internal/project"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRev4FunctionValues(t *testing.T) {
	source := `struct Box{let call:Fn<Int,Int>} func inc(n:Int):Int{return n+1} func choose():Fn<Int,Int>{return inc} func apply(f:Fn<Int,Int>,n:Int):Int{return f(n)} func work(){} func main(){let f:Fn<Int,Int>=choose();let fs:Fn<Int,Int>[]=[f];let m:Map<String,Fn<Int,Int>>={"inc":f};let box:Box=Box(call:f);assert apply(f,1)==2;assert fs[0](2)==3;assert m.require("inc")(3)==4;assert box.call(4)==5;let task:Fn<Void>=work;let t:Task=std.task.spawn(task);t.join();print("ok")}`
	p := checked(t, source)
	var out bytes.Buffer
	if e := interpreter.Run(context.Background(), p, &out); e != nil {
		t.Fatal(e)
	}
	if out.String() != "ok\n" {
		t.Fatal(out.String())
	}
	if _, e := formatter.Format("functions.craft", source); e != nil {
		t.Fatal(e)
	}
}
func TestRev4RejectFunctionMismatches(t *testing.T) {
	for _, source := range []string{
		`func f(n:Int):Int{return n} func main(){let x:Fn<String,Int>=f}`,
		`func f(n:Int):Int{return n} func main(){let x:Fn<Int,Int>=f;x(n:1)}`,
		`func f():Int{return 1} func main(){std.task.spawn(f)}`,
		`func f(){} func main(){std.json.stringify(f)}`,
		`func f(r:HttpRequest):HttpResponse{return std.http.response(200,std.bytes.fromString(""))} func main(){std.http.serve(std.http.config("127.0.0.1:0"),f,0)}`,
	} {
		p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: source}}}
		if _, _, e := p.Compile(); e == nil {
			t.Errorf("accepted %s", source)
		}
	}
}
func TestRev4CraftFrameworkAndApp(t *testing.T) {
	for _, dir := range []string{"../packages/api-framework", "../examples/api-server"} {
		p, e := project.LoadTests(dir)
		if e != nil {
			t.Fatal(e)
		}
		program, _, e := p.CompileMode(false)
		if e != nil {
			t.Fatal(e)
		}
		for _, test := range program.Tests {
			t.Run(filepath.Base(dir)+"/"+test.Name, func(t *testing.T) {
				var out bytes.Buffer
				if e := interpreter.RunTest(context.Background(), program, test, &out); e != nil {
					t.Fatal(e)
				}
			})
		}
		for _, s := range p.Sources {
			formatted, e := formatter.Format(s.Path, s.Text)
			if e != nil {
				t.Fatal(e)
			}
			again, e := formatter.Format(s.Path, formatted)
			if e != nil || again != formatted {
				t.Fatalf("format %s: %v", s.Path, e)
			}
		}
	}
}
func TestRev4TenThousandJoinedTasks(t *testing.T) {
	p := checked(t, `func work(){} func main(){for i in 0..10000{let t:Task=std.task.spawn(work);t.join()};print("ok")}`)
	var out bytes.Buffer
	if e := interpreter.Run(context.Background(), p, &out); e != nil {
		t.Fatal(e)
	}
}
func TestRev4PortableModuleBundle(t *testing.T) {
	base := t.TempDir()
	lib := filepath.Join(base, "lib")
	app := filepath.Join(base, "app")
	for _, dir := range []string{lib, app} {
		if e := project.Init(dir, true); e != nil {
			t.Fatal(e)
		}
	}
	write := func(path, text string) {
		t.Helper()
		if e := os.WriteFile(path, []byte(text), 0644); e != nil {
			t.Fatal(e)
		}
	}
	write(filepath.Join(lib, "src/main.craft"), `export struct Value{let n:Int} export func make():Value{return Value(n:7)} func private(){}`)
	write(filepath.Join(app, "craft.toml"), "name = \"app\"\nversion = \"0.1.0\"\nedition = \"2026\"\ndependency.lib = \"../lib\"\ndependency-version.lib = \"0.1.0\"\n")
	write(filepath.Join(app, "src/main.craft"), `import lib "lib" func main(){let v:lib.Value=lib.make();print(v.n)}`)
	p, e := project.Load(app)
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = p.Compile(); e != nil {
		t.Fatal(e)
	}
	path, e := p.Build()
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	portable := filepath.Join(t.TempDir(), "copy.craftbundle")
	if e = os.WriteFile(portable, data, 0644); e != nil {
		t.Fatal(e)
	}
	// Change originals: the portable bundle must keep its embedded source snapshot.
	write(filepath.Join(lib, "src/main.craft"), "broken original")
	b, e := project.LoadBundle(portable)
	if e != nil {
		t.Fatal(e)
	}
	program, _, e := b.Compile()
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	if e = interpreter.Run(context.Background(), program, &out); e != nil || out.String() != "7\n" {
		t.Fatalf("%v %s", e, &out)
	}
	write(filepath.Join(lib, "src/main.craft"), `func private(){}`)
	write(filepath.Join(app, "src/main.craft"), `import lib "lib" func main(){lib.private()}`)
	p, e = project.Load(app)
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = p.Compile(); e == nil || !strings.Contains(e.Error(), "private") {
		t.Fatalf("private symbol accepted: %v", e)
	}
}
