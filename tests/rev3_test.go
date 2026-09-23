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

func TestRev3StaticRejection(t *testing.T) {
	for _, source := range []string{
		`func f():Int{return 1} func main(){std.task.spawn(f)}`,
		`func f(n:Int){} func main(){std.task.spawn(f,"wrong")}`,
		`func f(){} func main(){let f:Int=1;std.task.spawn(f)}`,
		`func f(){} func main(){std.timer.after(5,f)}`,
		`func f(){} func main(){std.task.spawn(f())}`,
		`func main(){std.task.spawn(print)}`,
		`func f(){} func main(){std.task.spawn(callback:f)}`,
		`func f(){} func main(){let t:Task=std.task.spawn(f);std.json.stringify(t)}`,
		`func f(){} func main(){let t:Timer=std.task.spawn(f)}`,
		`func main(){std.time.sleep("1")}`,
	} {
		p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: source}}}
		if _, _, e := p.Compile(); e == nil {
			t.Errorf("accepted %s", source)
		}
	}
}

func TestRev3ExamplesAndFormatting(t *testing.T) {
	p, e := project.LoadTests(filepath.Join("..", "examples", "rev3-tasks"))
	if e != nil {
		t.Fatal(e)
	}
	program, _, e := p.CompileMode(false)
	if e != nil {
		t.Fatal(e)
	}
	for _, test := range program.Tests {
		t.Run(test.Name, func(t *testing.T) {
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
			t.Fatalf("formatter %s: %v", s.Path, e)
		}
	}
}

func TestRev3Bundle(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "rev3-bundle")
	if e := project.Init(dir, true); e != nil {
		t.Fatal(e)
	}
	source := `func worker(){print("done")} func main(){let t:Timer=std.timer.after(std.time.durationMilliseconds(0),worker);t.join()}`
	if e := os.WriteFile(filepath.Join(dir, "src", "main.craft"), []byte(source), 0644); e != nil {
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
	var out, stderr bytes.Buffer
	if code := cli.RunWithInput(context.Background(), []string{"run", bundle}, dir, nil, &out, &stderr); code != 0 || out.String() != "done\n" {
		t.Fatalf("%d %s %s", code, &out, &stderr)
	}
}

func TestFmtCheckTypoDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.craft")
	source := []byte(`func main(){print(1)}`)
	if e := os.WriteFile(path, source, 0644); e != nil {
		t.Fatal(e)
	}
	var out, stderr bytes.Buffer
	code := cli.RunWithInput(context.Background(), []string{"fmt", "check"}, dir, nil, &out, &stderr)
	if code != 4 || !strings.Contains(stderr.String(), "craft fmt --check") {
		t.Fatalf("%d %s", code, &stderr)
	}
	after, e := os.ReadFile(path)
	if e != nil || !bytes.Equal(after, source) {
		t.Fatal("fmt typo wrote source")
	}
	if e := os.WriteFile(filepath.Join(dir, "check"), source, 0644); e != nil {
		t.Fatal(e)
	}
	stderr.Reset()
	cli.RunWithInput(context.Background(), []string{"fmt", "check"}, dir, nil, &out, &stderr)
	if strings.Contains(stderr.String(), "hint:") {
		t.Fatal("existing positional path treated as flag")
	}
}
