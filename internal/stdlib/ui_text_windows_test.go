//go:build windows

package stdlib

import (
	"strings"
	"testing"
)

func TestTextInputValueLimits(t *testing.T) {
	for _, value := range []string{"", "สวัสดี", "👋🏽", strings.Repeat("a", 1024)} {
		if err := validateInputValue(value); err != nil {
			t.Fatalf("rejected valid input %q: %v", value, err)
		}
	}
	for _, value := range []string{"a\nb", "a\rb", "a\x00b", strings.Repeat("a", 1025), strings.Repeat("😀", 513)} {
		if err := validateInputValue(value); err == nil {
			t.Fatalf("accepted invalid input %q", value)
		}
	}
}

func TestFocusableOrderIncludesInputs(t *testing.T) {
	root := &UiNode{kind: "Column", id: "root", children: []*UiNode{
		{kind: "Button", id: "save", tabStop: true},
		{kind: "TextInput", id: "name", tabStop: true},
		{kind: "ScrollView", id: "scroll", children: []*UiNode{{kind: "TextInput", id: "notes", tabStop: true}}},
	}}
	var ids []string
	focusableIDs(root, &ids)
	if strings.Join(ids, ",") != "save,name,notes" {
		t.Fatalf("unexpected tab order: %v", ids)
	}
	if !hasFocusable(root, "name") || hasFocusable(root, "scroll") {
		t.Fatal("focusable lookup returned wrong kind")
	}
}
