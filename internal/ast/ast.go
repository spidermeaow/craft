package ast

import (
	"craft/internal/diagnostics"
	"strings"
)

type Type string

const (
	Int       Type = "Int"
	Float     Type = "Float"
	Bool      Type = "Bool"
	String    Type = "String"
	Void      Type = "Void"
	Exception Type = "Exception"
	Null      Type = "null"
	JsonValue Type = "JsonValue"
	DateTime  Type = "DateTime"
	Duration  Type = "Duration"
	Random    Type = "Random"
	Task      Type = "Task"
	Timer     Type = "Timer"
	Bytes     Type = "Bytes"
	Context   Type = "Context"
)

func (t Type) IsArray() bool    { return strings.HasSuffix(string(t), "[]") }
func (t Type) Element() Type    { return Type(strings.TrimSuffix(string(t), "[]")) }
func (t Type) Array() Type      { return Type(string(t) + "[]") }
func (t Type) IsOptional() bool { return strings.HasSuffix(string(t), "?") }
func (t Type) Base() Type       { return Type(strings.TrimSuffix(string(t), "?")) }
func (t Type) Optional() Type   { return Type(string(t) + "?") }
func (t Type) IsMap() bool {
	return strings.HasPrefix(string(t), "Map<String,") && strings.HasSuffix(string(t), ">")
}
func (t Type) MapElement() Type {
	return Type(strings.TrimSuffix(strings.TrimPrefix(string(t), "Map<String,"), ">"))
}
func (t Type) Map() Type       { return Type("Map<String," + string(t) + ">") }
func (t Type) Primitive() bool { return t == Int || t == Float || t == Bool || t == String }

type Program struct {
	Imports   []Import
	Structs   []*Struct
	Functions []*Function
	Tests     []*Test
}
type Struct struct {
	Exported bool
	Name     string
	Fields   []Field
	Pos      diagnostics.Position
}
type Field struct {
	Name    string
	Type    Type
	Mutable bool
	Pos     diagnostics.Position
}
type Test struct {
	Name string
	Body *Stmt
	Pos  diagnostics.Position
}
type Parameter struct {
	Name string
	Type Type
	Pos  diagnostics.Position
}
type Function struct {
	Exported bool
	Name     string
	Params   []Parameter
	Return   Type
	Body     *Stmt
	Pos      diagnostics.Position
}

// Kind discriminates statement and expression nodes. Frontend nodes contain no runtime state.
type Stmt struct {
	Kind       string
	Pos        diagnostics.Position
	Name       string
	Type       Type
	Expr, End  *Expr
	Body, Else *Stmt
	Statements []*Stmt
}
type Expr struct {
	Kind        string
	Pos         diagnostics.Position
	Text        string
	Type        Type
	Left, Right *Expr
	Args        []*Expr
	ArgNames    []string
	// ArgOrder maps source-order arguments to formal parameters after checking.
	ArgOrder []int
}
