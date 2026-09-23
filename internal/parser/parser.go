package parser

import (
	"craft/internal/ast"
	"craft/internal/diagnostics"
	"craft/internal/lexer"
)

type parser struct {
	tokens   []lexer.Token
	i, depth int
}

func Parse(tokens []lexer.Token) (*ast.Program, error) {
	p := &parser{tokens: tokens}
	program := &ast.Program{}
	p.separators()
	for !p.at("eof") {
		if p.at("identifier") && p.tok().Text == "import" {
			p.take()
			alias, e := p.need("identifier")
			if e != nil {
				return nil, e
			}
			dep, e := p.need("string")
			if e != nil {
				return nil, e
			}
			program.Imports = append(program.Imports, ast.Import{Alias: alias.Text, Dependency: dep.Text})
			p.separators()
			continue
		}
		exported := false
		if p.at("identifier") && p.tok().Text == "export" {
			p.take()
			exported = true
		}
		if p.at("struct") {
			s, e := p.structure()
			if e != nil {
				return nil, e
			}
			s.Exported = exported
			program.Structs = append(program.Structs, s)
			p.separators()
			continue
		}
		if p.accept("test") {
			if exported {
				return nil, diagnostics.New(p.tok().Pos, "syntax error: tests cannot be exported")
			}
			n, e := p.need("string")
			if e != nil {
				return nil, e
			}
			b, e := p.block()
			if e != nil {
				return nil, e
			}
			program.Tests = append(program.Tests, &ast.Test{Name: n.Text, Body: b, Pos: n.Pos})
			p.separators()
			continue
		}
		f, e := p.function()
		if e != nil {
			return nil, e
		}
		f.Exported = exported
		program.Functions = append(program.Functions, f)
		p.separators()
	}
	return program, nil
}
func (p *parser) tok() lexer.Token { return p.tokens[p.i] }
func (p *parser) at(k string) bool { return p.tok().Kind == k }
func (p *parser) take() lexer.Token {
	t := p.tok()
	if !p.at("eof") {
		p.i++
	}
	return t
}
func (p *parser) accept(k string) bool {
	if p.at(k) {
		p.take()
		return true
	}
	return false
}
func (p *parser) need(k string) (lexer.Token, error) {
	if p.at(k) {
		return p.take(), nil
	}
	return p.tok(), diagnostics.New(p.tok().Pos, "syntax error: expected %s, found %q", k, p.tok().Text)
}
func (p *parser) separators() {
	for p.at("newline") || p.at(";") {
		p.take()
	}
}
func (p *parser) newlines() {
	for p.at("newline") {
		p.take()
	}
}
func (p *parser) function() (*ast.Function, error) {
	t, e := p.need("func")
	if e != nil {
		return nil, e
	}
	name, e := p.need("identifier")
	if e != nil {
		return nil, e
	}
	f := &ast.Function{Name: name.Text, Return: ast.Void, Pos: t.Pos}
	if _, e = p.need("("); e != nil {
		return nil, e
	}
	if !p.at(")") {
		for {
			n, e := p.need("identifier")
			if e != nil {
				return nil, e
			}
			if _, e = p.need(":"); e != nil {
				return nil, e
			}
			typ, e := p.typeName()
			if e != nil {
				return nil, e
			}
			f.Params = append(f.Params, ast.Parameter{Name: n.Text, Type: typ, Pos: n.Pos})
			if !p.accept(",") {
				break
			}
		}
	}
	if _, e = p.need(")"); e != nil {
		return nil, e
	}
	if p.accept(":") {
		typ, e := p.typeName()
		if e != nil {
			return nil, e
		}
		f.Return = typ
	}
	f.Body, e = p.block()
	return f, e
}
func (p *parser) block() (*ast.Stmt, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 256 {
		return nil, diagnostics.New(p.tok().Pos, "syntax error: nesting exceeds 256 levels")
	}
	p.newlines()
	t, e := p.need("{")
	if e != nil {
		return nil, e
	}
	b := &ast.Stmt{Kind: "block", Pos: t.Pos}
	p.separators()
	for !p.at("}") {
		if p.at("eof") {
			return nil, diagnostics.New(p.tok().Pos, "syntax error: expected closing }")
		}
		s, e := p.statement()
		if e != nil {
			return nil, e
		}
		b.Statements = append(b.Statements, s)
		p.separators()
	}
	p.take()
	return b, nil
}
func (p *parser) statement() (*ast.Stmt, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 256 {
		return nil, diagnostics.New(p.tok().Pos, "syntax error: nesting exceeds 256 levels")
	}
	t := p.tok()
	s := &ast.Stmt{Kind: t.Kind, Pos: t.Pos}
	var e error
	switch t.Kind {
	case "{":
		return p.block()
	case "let", "var":
		p.take()
		n, err := p.need("identifier")
		if err != nil {
			return nil, err
		}
		s.Name = n.Text
		if _, e = p.need(":"); e != nil {
			return nil, e
		}
		typ, err := p.typeName()
		if err != nil {
			return nil, err
		}
		s.Type = typ
		if _, e = p.need("="); e != nil {
			return nil, e
		}
		s.Expr, e = p.expression(1)
	case "if":
		p.take()
		if p.accept("let") {
			s.Kind = "iflet"
			n, err := p.need("identifier")
			if err != nil {
				return nil, err
			}
			s.Name = n.Text
			if _, e = p.need("="); e != nil {
				return nil, e
			}
		}
		s.Expr, e = p.expression(1)
		if e != nil {
			return nil, e
		}
		s.Body, e = p.block()
		if e != nil {
			return nil, e
		}
		saved := p.i
		p.newlines()
		if p.accept("else") {
			if p.at("if") {
				s.Else, e = p.statement()
			} else {
				s.Else, e = p.block()
			}
		} else {
			p.i = saved
		}
		return s, e
	case "while":
		p.take()
		s.Expr, e = p.expression(1)
		if e != nil {
			return nil, e
		}
		s.Body, e = p.block()
		return s, e
	case "for":
		p.take()
		n, err := p.need("identifier")
		if err != nil {
			return nil, err
		}
		s.Name = n.Text
		if _, e = p.need("in"); e != nil {
			return nil, e
		}
		s.Expr, e = p.expression(1)
		if e != nil {
			return nil, e
		}
		if p.accept("..") {
			s.End, e = p.expression(1)
			if e != nil {
				return nil, e
			}
		}
		s.Body, e = p.block()
		return s, e
	case "return":
		p.take()
		if !p.at("newline") && !p.at(";") && !p.at("}") && !p.at("eof") {
			s.Expr, e = p.expression(1)
		}
	case "break", "continue":
		p.take()
	case "throw", "assert":
		p.take()
		s.Expr, e = p.expression(1)
	case "defer":
		p.take()
		s.Body, e = p.block()
		return s, e
	case "try":
		p.take()
		s.Body, e = p.block()
		if e != nil {
			return nil, e
		}
		p.newlines()
		if _, e = p.need("catch"); e != nil {
			return nil, e
		}
		n, err := p.need("identifier")
		if err != nil {
			return nil, err
		}
		s.Name = n.Text
		s.Else, e = p.block()
		return s, e
	default:
		s.Kind = "expression"
		s.Expr, e = p.expression(1)
	}
	if e != nil {
		return nil, e
	}
	if !p.at("newline") && !p.at(";") && !p.at("}") && !p.at("eof") {
		return nil, diagnostics.New(p.tok().Pos, "syntax error: separate statements with a newline or semicolon")
	}
	return s, nil
}
func precedence(k string) int {
	switch k {
	case "=", "+=", "-=", "*=", "/=", "%=":
		return 1
	case "||":
		return 2
	case "&&":
		return 3
	case "==", "!=":
		return 4
	case "<", ">", "<=", ">=":
		return 5
	case "+", "-":
		return 6
	case "*", "/", "%":
		return 7
	}
	return 0
}
func (p *parser) expression(min int) (*ast.Expr, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 256 {
		return nil, diagnostics.New(p.tok().Pos, "syntax error: nesting exceeds 256 levels")
	}
	t := p.take()
	x := &ast.Expr{Pos: t.Pos, Text: t.Text}
	var e error
	switch t.Kind {
	case "null":
		x.Kind = "literal"
		x.Type = ast.Null
	case "{":
		x.Kind = "map"
		p.newlines()
		for !p.at("}") {
			key, err := p.need("string")
			if err != nil {
				return nil, err
			}
			if _, err = p.need(":"); err != nil {
				return nil, err
			}
			a, err := p.expression(1)
			if err != nil {
				return nil, err
			}
			x.ArgNames = append(x.ArgNames, key.Text)
			x.Args = append(x.Args, a)
			p.newlines()
			if !p.accept(",") {
				break
			}
			p.newlines()
		}
		if _, e = p.need("}"); e != nil {
			return nil, e
		}
	case "integer":
		x.Kind = "literal"
		x.Type = ast.Int
	case "float":
		x.Kind = "literal"
		x.Type = ast.Float
	case "string":
		x.Kind = "literal"
		x.Type = ast.String
	case "true", "false":
		x.Kind = "literal"
		x.Type = ast.Bool
	case "identifier":
		x.Kind = "identifier"
	case "[":
		x.Kind = "array"
		for !p.at("]") {
			a, err := p.expression(1)
			if err != nil {
				return nil, err
			}
			x.Args = append(x.Args, a)
			if !p.accept(",") {
				break
			}
		}
		if _, e = p.need("]"); e != nil {
			return nil, e
		}
	case "!", "-", "+":
		x.Kind = "unary"
		x.Right, e = p.expression(8)
	case "(":
		x, e = p.expression(1)
		if e == nil {
			_, e = p.need(")")
		}
	default:
		return nil, diagnostics.New(t.Pos, "syntax error: expected expression, found %q", t.Text)
	}
	if e != nil {
		return nil, e
	}
	for {
		if p.accept(".") {
			n, err := p.need("identifier")
			if err != nil {
				return nil, err
			}
			x = &ast.Expr{Kind: "member", Pos: n.Pos, Left: x, Text: n.Text}
			continue
		}
		if p.accept("[") {
			index, err := p.expression(1)
			if err != nil {
				return nil, err
			}
			if _, e = p.need("]"); e != nil {
				return nil, e
			}
			x = &ast.Expr{Kind: "index", Pos: x.Pos, Left: x, Right: index}
			continue
		}
		if p.accept("(") {
			call := &ast.Expr{Kind: "call", Pos: x.Pos, Left: x}
			if !p.at(")") {
				for {
					name := ""
					if p.at("identifier") && p.i+1 < len(p.tokens) && p.tokens[p.i+1].Kind == ":" {
						name = p.take().Text
						p.take()
					}
					a, err := p.expression(1)
					if err != nil {
						return nil, err
					}
					call.Args = append(call.Args, a)
					call.ArgNames = append(call.ArgNames, name)
					if !p.accept(",") {
						break
					}
				}
			}
			if _, e = p.need(")"); e != nil {
				return nil, e
			}
			x = call
			continue
		}
		op := p.tok()
		prec := precedence(op.Kind)
		if prec < min {
			break
		}
		p.take()
		next := prec + 1
		kind := "binary"
		if prec == 1 {
			next = prec
			kind = "assignment"
		}
		r, err := p.expression(next)
		if err != nil {
			return nil, err
		}
		x = &ast.Expr{Kind: kind, Text: op.Text, Pos: op.Pos, Left: x, Right: r}
	}
	return x, nil
}

func (p *parser) typeName() (ast.Type, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 256 {
		return "", diagnostics.New(p.tok().Pos, "syntax error: type nesting exceeds 256 levels")
	}
	t, e := p.need("identifier")
	if e != nil {
		return "", e
	}
	typ := ast.Type(t.Text)
	if p.accept(".") {
		n, e := p.need("identifier")
		if e != nil {
			return "", e
		}
		typ = ast.Type(t.Text + "." + n.Text)
	}
	if t.Text == "Fn" && p.accept("<") {
		var parts []ast.Type
		for {
			v, e := p.typeName()
			if e != nil {
				return "", e
			}
			parts = append(parts, v)
			if !p.accept(",") {
				break
			}
		}
		if p.at(">=") {
			p.tokens[p.i].Kind = "="
			p.tokens[p.i].Text = "="
			p.tokens[p.i].Pos.Column++
		} else if _, e = p.need(">"); e != nil {
			return "", e
		}
		typ = ast.FunctionType(parts[0], parts[1:])
	}
	if t.Text == "Map" && p.accept("<") {
		key, err := p.need("identifier")
		if err != nil {
			return "", err
		}
		if key.Text != "String" {
			return "", diagnostics.New(key.Pos, "syntax error: Map keys must be String")
		}
		if _, e = p.need(","); e != nil {
			return "", e
		}
		value, err := p.typeName()
		if err != nil {
			return "", err
		}
		// A closing generic followed immediately by assignment is lexed as >=.
		if p.at(">=") {
			p.tokens[p.i].Kind = "="
			p.tokens[p.i].Text = "="
			p.tokens[p.i].Pos.Column++
		} else if _, e = p.need(">"); e != nil {
			return "", e
		}
		typ = value.Map()
	}
	suffixes := 0
	for {
		if suffixes+p.depth > 256 {
			return "", diagnostics.New(p.tok().Pos, "syntax error: type nesting exceeds 256 levels")
		}
		if p.accept("[") {
			if _, e = p.need("]"); e != nil {
				return "", e
			}
			typ = typ.Array()
		} else if p.accept("?") {
			typ = typ.Optional()
		} else {
			break
		}
		suffixes++
	}
	return typ, nil
}

func (p *parser) structure() (*ast.Struct, error) {
	t := p.take()
	n, e := p.need("identifier")
	if e != nil {
		return nil, e
	}
	s := &ast.Struct{Name: n.Text, Pos: t.Pos}
	p.newlines()
	if _, e = p.need("{"); e != nil {
		return nil, e
	}
	p.separators()
	for !p.at("}") {
		f := p.take()
		if f.Kind != "let" && f.Kind != "var" {
			return nil, diagnostics.New(f.Pos, "syntax error: struct field requires let or var")
		}
		n, e := p.need("identifier")
		if e != nil {
			return nil, e
		}
		if _, e = p.need(":"); e != nil {
			return nil, e
		}
		typ, e := p.typeName()
		if e != nil {
			return nil, e
		}
		s.Fields = append(s.Fields, ast.Field{Name: n.Text, Type: typ, Mutable: f.Kind == "var", Pos: n.Pos})
		if !p.at("}") && !p.at("newline") && !p.at(";") {
			return nil, diagnostics.New(p.tok().Pos, "syntax error: separate fields with a newline or semicolon")
		}
		p.separators()
	}
	p.take()
	return s, nil
}
