package lexer

import (
	"strings"
	"testing"
)

func TestRangeFloatCommentsAndLocations(t *testing.T) {
	tokens, err := Scan("sample.craft", "// comment\r\nlet ชื่อ: String = \"hello\\nโลก\"\nfor i in 0..5 { print(1.25e2) }")
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Kind != "let" || tokens[1].Pos.Line != 2 || tokens[1].Pos.Column != 1 {
		t.Fatalf("wrong location: %+v", tokens[1])
	}
	foundRange, foundFloat, foundString := false, false, false
	for _, tok := range tokens {
		if tok.Kind == ".." {
			foundRange = true
		}
		if tok.Kind == "float" && tok.Text == "1.25e2" {
			foundFloat = true
		}
		if tok.Kind == "string" && tok.Text == "hello\nโลก" {
			foundString = true
		}
	}
	if !foundRange || !foundFloat || !foundString {
		t.Fatalf("missing tokens: %+v", tokens)
	}
}
func TestRejectMalformedInput(t *testing.T) {
	for _, source := range []string{"\"unterminated", "\"bad\\q\"", "1e+", "@"} {
		t.Run(source, func(t *testing.T) {
			_, err := Scan("bad.craft", source)
			if err == nil || !strings.Contains(err.Error(), "bad.craft:1:") {
				t.Fatalf("expected located error, got %v", err)
			}
		})
	}
}
func FuzzScan(f *testing.F) {
	for _, s := range []string{"func main() {}", "\"x\\\"", "0..5", "let ชื่อ: String = \"ไทย\""} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 65536 {
			t.Skip()
		}
		tokens, e := Scan("fuzz.craft", s)
		if e == nil && (len(tokens) == 0 || tokens[len(tokens)-1].Kind != "eof") {
			t.Fatal("missing EOF")
		}
	})
}
