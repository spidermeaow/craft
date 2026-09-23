package ast

import "strings"

type Import struct{ Alias, Dependency string }

// Fn<Return,Parameter,...> is an immutable, non-capturing function type.
func FunctionType(result Type, params []Type) Type {
	parts := []string{string(result)}
	for _, p := range params {
		parts = append(parts, string(p))
	}
	return Type("Fn<" + strings.Join(parts, ",") + ">")
}
func (t Type) Function() (Type, []Type, bool) {
	s := string(t)
	if !strings.HasPrefix(s, "Fn<") || !strings.HasSuffix(s, ">") {
		return "", nil, false
	}
	parts := SplitTypes(s[3 : len(s)-1])
	if len(parts) == 0 {
		return "", nil, false
	}
	return parts[0], parts[1:], true
}
func SplitTypes(s string) []Type {
	depth, start := 0, 0
	var result []Type
	for i, r := range s {
		switch r {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				result = append(result, Type(s[start:i]))
				start = i + 1
			}
		}
	}
	return append(result, Type(s[start:]))
}
func (f *Function) Type() Type {
	params := make([]Type, len(f.Params))
	for i, p := range f.Params {
		params[i] = p.Type
	}
	return FunctionType(f.Return, params)
}
