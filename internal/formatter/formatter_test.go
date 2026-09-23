package formatter

import (
	"strings"
	"testing"
)

func TestFormatPreservesCommentsAndIsIdempotent(t *testing.T) {
	cases := []string{
		"// leading\nfunc main(){var a:Int[]=[1,2];a.append(3)// trailing\nprint(a[0])}\n",
		"func main(){print( // call\n 1+2,\n -3)}",
		"func main(){try{throw Exception(\"// not a comment\")}catch e{print(e.message)}}",
		"func main(){for i in 1..(3+1){if i==2{continue};print(i)}}\ntest \"a\"{assert true}",
		"func main(){defer{print(\"done\")};return // don't join return and call\nprint(1)}",
		"func main(){let a:Int[][]=[[],[1,2]];print(a)}",
	}
	for _, source := range cases {
		out, e := Format("main.craft", source)
		if e != nil {
			t.Fatal(e)
		}
		again, e := Format("main.craft", out)
		if e != nil || out != again {
			t.Fatalf("non-idempotent: %v\n%s\n%s", e, out, again)
		}
		for _, line := range strings.Split(source, "\n") {
			if strings.Contains(line, "// leading") && !strings.Contains(out, "// leading") {
				t.Fatal("lost comment")
			}
		}
	}
}
func TestFormatRejectsMalformedSource(t *testing.T) {
	if _, e := Format("bad.craft", "func main( {"); e == nil {
		t.Fatal("accepted malformed source")
	}
}
