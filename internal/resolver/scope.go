// Package resolver owns lexical symbols independently from execution storage.
package resolver

import "craft/internal/ast"

type Symbol struct {
	Type    ast.Type
	Mutable bool
}
type Scope struct {
	Parent *Scope
	Values map[string]Symbol
}

func New(parent *Scope) *Scope { return &Scope{Parent: parent, Values: map[string]Symbol{}} }
func (s *Scope) Find(name string) (Symbol, bool) {
	for ; s != nil; s = s.Parent {
		if v, ok := s.Values[name]; ok {
			return v, true
		}
	}
	return Symbol{}, false
}
