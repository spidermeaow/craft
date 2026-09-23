// Package runtime defines values shared by execution engines and the standard library.
package runtime

import (
	"craft/internal/ast"
	"craft/internal/diagnostics"
	"fmt"
	"strings"
)

type Function struct{ Declaration *ast.Function }

func (f Function) String() string { return "Function(" + f.Declaration.Name + ")" }

type Array struct{ Items []any }

func (a *Array) String() string {
	parts := make([]string, len(a.Items))
	for i, v := range a.Items {
		parts[i] = fmt.Sprint(v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Clone gives arrays value semantics, including arrays nested in other arrays.
func Clone(v any) any {
	switch x := v.(type) {
	case Optional:
		return Optional{Valid: x.Valid, Value: Clone(x.Value)}
	case *Map:
		m := NewMap()
		for _, k := range x.Keys {
			m.Set(k, Clone(x.Values[k]))
		}
		return m
	case *Struct:
		s := &Struct{Name: x.Name, Order: append([]string(nil), x.Order...), Fields: map[string]any{}}
		for k, v := range x.Fields {
			s.Fields[k] = Clone(v)
		}
		return s
	case *Random:
		r := *x
		return &r
	}
	if a, ok := v.(*Array); ok {
		copy := &Array{Items: make([]any, len(a.Items))}
		for i, item := range a.Items {
			copy.Items[i] = Clone(item)
		}
		return copy
	}
	return v
}

type Optional struct {
	Valid bool
	Value any
}

func Some(v any) Optional { return Optional{Valid: true, Value: Clone(v)} }
func (v Optional) String() string {
	if !v.Valid {
		return "null"
	}
	return fmt.Sprint(v.Value)
}

type Map struct {
	Keys   []string
	Values map[string]any
}

func NewMap() *Map { return &Map{Values: map[string]any{}} }
func (m *Map) Set(k string, v any) {
	if _, ok := m.Values[k]; !ok {
		m.Keys = append(m.Keys, k)
	}
	m.Values[k] = v
}
func (m *Map) String() string {
	p := []string{}
	for _, k := range m.Keys {
		p = append(p, fmt.Sprintf("%q: %v", k, m.Values[k]))
	}
	return "{" + strings.Join(p, ", ") + "}"
}

type Struct struct {
	Name   string
	Order  []string
	Fields map[string]any
}

func (s *Struct) String() string {
	p := []string{}
	for _, k := range s.Order {
		p = append(p, k+": "+fmt.Sprint(s.Fields[k]))
	}
	return s.Name + "(" + strings.Join(p, ", ") + ")"
}

// Random is a copyable SplitMix64 state. It is not a cryptographic generator.
type Random struct{ State uint64 }

func (r *Random) Next() uint64 {
	r.State += 0x9e3779b97f4a7c15
	z := r.State
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}
func (r *Random) String() string { return "Random" }

type Exception struct {
	Message, Kind string
	Detail        *diagnostics.Error
}

func (e *Exception) Error() string {
	if e.Detail != nil {
		return e.Detail.Error()
	}
	return e.Kind + ": " + e.Message
}
func (e *Exception) Unwrap() error {
	if e.Detail == nil {
		return nil
	}
	return e.Detail
}
func (e *Exception) String() string { return e.Kind + ": " + e.Message }
