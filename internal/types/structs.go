package types

import (
	"craft/internal/ast"
	"craft/internal/diagnostics"
	"craft/internal/stdlib"
)

func (c *checker) field(t ast.Type, name string) (ast.Field, bool) {
	if s := c.structs[string(t)]; s != nil {
		for _, f := range s.Fields {
			if f.Name == name {
				return f, true
			}
		}
	}
	return ast.Field{}, false
}
func (c *checker) registerStructs(p *ast.Program) error {
	for _, s := range p.Structs {
		_, builtin := stdlib.Lookup(s.Name)
		if builtin || s.Name == "std" || s.Name == "Map" || s.Name == "Fn" || c.valid(ast.Type(s.Name), true) {
			return diagnostics.New(s.Pos, "type error: duplicate or reserved struct name %s", s.Name)
		}
		c.structs[s.Name] = s
	}
	for _, s := range p.Structs {
		seen := map[string]bool{}
		for _, f := range s.Fields {
			if seen[f.Name] || !c.valid(f.Type, false) {
				return diagnostics.New(f.Pos, "type error: duplicate field or invalid field type %s", f.Name)
			}
			seen[f.Name] = true
		}
	}
	// Only direct required fields make a value impossible to construct. Optional,
	// Array and Map edges terminate the cycle because they can start empty.
	state := map[string]int{}
	var visit func(string) error
	visit = func(name string) error {
		s := c.structs[name]
		if s == nil || state[name] == 2 {
			return nil
		}
		if state[name] == 1 {
			return diagnostics.New(s.Pos, "type error: recursive required value fields in struct %s", name)
		}
		state[name] = 1
		for _, f := range s.Fields {
			if e := visit(string(f.Type)); e != nil {
				return e
			}
		}
		state[name] = 2
		return nil
	}
	for _, s := range p.Structs {
		if e := visit(s.Name); e != nil {
			return e
		}
	}
	return nil
}
