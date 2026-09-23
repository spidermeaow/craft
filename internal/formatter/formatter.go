// Package formatter preserves comments and validates that formatting keeps the parsed AST.
package formatter

import (
	"craft/internal/ast"
	"craft/internal/diagnostics"
	"craft/internal/lexer"
	"craft/internal/parser"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func Format(file, source string) (string, error) {
	tokens, e := lexer.Scan(file, source)
	if e != nil {
		return "", e
	}
	before, e := parser.Parse(tokens)
	if e != nil {
		return "", e
	}
	tokens, e = lexer.ScanWithComments(file, source)
	if e != nil {
		return "", e
	}
	mapPositions := map[diagnostics.Position]bool{}
	var markExpr func(*ast.Expr)
	var markStmt func(*ast.Stmt)
	markExpr = func(x *ast.Expr) {
		pending := []*ast.Expr{x}
		for len(pending) > 0 {
			n := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			if n == nil {
				continue
			}
			if n.Kind == "map" {
				mapPositions[n.Pos] = true
			}
			pending = append(pending, n.Left, n.Right)
			pending = append(pending, n.Args...)
		}
	}
	markStmt = func(s *ast.Stmt) {
		if s == nil {
			return
		}
		markExpr(s.Expr)
		markExpr(s.End)
		markStmt(s.Body)
		markStmt(s.Else)
		for _, st := range s.Statements {
			markStmt(st)
		}
	}
	for _, f := range before.Functions {
		markStmt(f.Body)
	}
	for _, t := range before.Tests {
		markStmt(t.Body)
	}
	braces := []bool{}
	var out strings.Builder
	line := ""
	indent, groups := 0, 0
	lineIndent := 0
	previous := ""
	unary := false
	flush := func() {
		if strings.TrimSpace(line) != "" {
			out.WriteString(strings.Repeat("    ", lineIndent))
			out.WriteString(strings.TrimSpace(line))
			out.WriteByte('\n')
		}
		line = ""
		previous = ""
		unary = false
	}
	for _, t := range tokens {
		k := t.Kind
		if k == "eof" {
			break
		}
		if k == "newline" || k == ";" {
			flush()
			continue
		}
		if line == "" {
			lineIndent = indent
			if groups > 0 {
				lineIndent++
			}
		}
		if k == "comment" {
			if line != "" {
				line += " "
			}
			line += t.Text
			flush()
			continue
		}
		if k == "{" {
			braces = append(braces, mapPositions[t.Pos])
			if line != "" {
				line += " "
			}
			line += "{"
			flush()
			indent++
			continue
		}
		if k == "}" {
			isMap := false
			if len(braces) > 0 {
				isMap = braces[len(braces)-1]
				braces = braces[:len(braces)-1]
			}
			flush()
			indent--
			if indent < 0 {
				return "", fmt.Errorf("formatter: unmatched closing brace")
			}
			line = "}"
			lineIndent = indent
			if isMap {
				previous = "}"
				continue
			}
			flush()
			if indent == 0 {
				out.WriteByte('\n')
			}
			continue
		}
		text := t.Text
		if k == "string" {
			text = strconv.Quote(text)
		}
		space := line != ""
		isUnary := (k == "!" || k == "-" || k == "+") && expressionStart(previous)
		if k == "?" || k == ")" || k == "]" || k == "," || k == ":" || k == "." || k == ".." {
			space = false
		}
		if previous == "(" || previous == "[" || previous == "." || previous == ".." || unary {
			space = false
		}
		if k == "(" && (previous == "identifier" || previous == ")" || previous == "]") {
			space = false
		}
		if k == "[" && (previous == "identifier" || previous == "]" || previous == ")") {
			space = false
		}
		if space {
			line += " "
		}
		line += text
		if k == "(" || k == "[" {
			groups++
		}
		if k == ")" || k == "]" {
			groups--
		}
		previous = k
		unary = isUnary
	}
	flush()
	formatted := strings.TrimRight(out.String(), "\n")
	if formatted != "" {
		formatted += "\n"
	}
	afterTokens, e := lexer.Scan(file, formatted)
	if e != nil {
		return "", e
	}
	after, e := parser.Parse(afterTokens)
	if e != nil {
		return "", e
	}
	beforeShape, e := fingerprint(before)
	if e != nil {
		return "", e
	}
	afterShape, e := fingerprint(after)
	if e != nil {
		return "", e
	}
	if beforeShape != afterShape {
		return "", fmt.Errorf("formatter refused to change program structure: %s", file)
	}
	return formatted, nil
}
func expressionStart(previous string) bool {
	switch previous {
	case "", "(", "[", ",", ":", "=", "+=", "-=", "*=", "/=", "%=", "+", "-", "*", "/", "%", "!", "==", "!=", "<", ">", "<=", ">=", "&&", "||", "return", "throw", "assert", "in", "..":
		return true
	}
	return false
}
func fingerprint(p *ast.Program) (string, error) {
	depth := 0
	var err error
	var expr func(*ast.Expr)
	expr = func(x *ast.Expr) {
		if x == nil || err != nil {
			return
		}
		depth++
		defer func() { depth-- }()
		if depth > 512 {
			err = diagnostics.New(x.Pos, "syntax error: formatter expression limit is 512 levels")
			return
		}
		x.Pos = diagnostics.Position{}
		expr(x.Left)
		expr(x.Right)
		for _, a := range x.Args {
			expr(a)
		}
	}
	var stmt func(*ast.Stmt)
	stmt = func(s *ast.Stmt) {
		if s == nil {
			return
		}
		s.Pos = diagnostics.Position{}
		expr(s.Expr)
		expr(s.End)
		stmt(s.Body)
		stmt(s.Else)
		for _, child := range s.Statements {
			stmt(child)
		}
	}
	for _, s := range p.Structs {
		s.Pos = diagnostics.Position{}
		for i := range s.Fields {
			s.Fields[i].Pos = diagnostics.Position{}
		}
	}
	for _, f := range p.Functions {
		f.Pos = diagnostics.Position{}
		for i := range f.Params {
			f.Params[i].Pos = diagnostics.Position{}
		}
		stmt(f.Body)
	}
	for _, test := range p.Tests {
		test.Pos = diagnostics.Position{}
		stmt(test.Body)
	}
	if err != nil {
		return "", err
	}
	b, e := json.Marshal(p)
	return string(b), e
}
