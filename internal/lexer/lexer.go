package lexer

import (
	"craft/internal/diagnostics"
	"strconv"
	"unicode"
)

type Token struct {
	Kind, Text string
	Pos        diagnostics.Position
}
type scanner struct {
	input                []rune
	file                 string
	i, line, col, parens int
	comments             bool
	tokens               []Token
}

func Scan(file, source string) ([]Token, error) {
	return scan(file, source, false)
}
func ScanWithComments(file, source string) ([]Token, error) { return scan(file, source, true) }
func scan(file, source string, comments bool) ([]Token, error) {
	s := &scanner{input: []rune(source), file: file, line: 1, col: 1, comments: comments}
	for s.i < len(s.input) {
		c := s.peek(0)
		p := s.pos()
		if c == '\ufeff' && s.i == 0 {
			s.next()
			continue
		}
		if c == ' ' || c == '\t' || c == '\r' {
			s.next()
			continue
		}
		if c == '\n' {
			s.next()
			if s.parens == 0 {
				s.emit("newline", "\n", p)
			}
			continue
		}
		if c == '/' && s.peek(1) == '/' {
			start := s.i
			for s.i < len(s.input) && s.peek(0) != '\n' {
				s.next()
			}
			if s.comments {
				s.emit("comment", string(s.input[start:s.i]), p)
			}
			continue
		}
		start := s.i
		if unicode.IsLetter(c) || c == '_' {
			for unicode.IsLetter(s.peek(0)) || unicode.IsMark(s.peek(0)) || unicode.IsDigit(s.peek(0)) || s.peek(0) == '_' {
				s.next()
			}
			t := string(s.input[start:s.i])
			kind := "identifier"
			switch t {
			case "struct", "null", "func", "let", "var", "if", "else", "while", "for", "in", "return", "break", "continue", "true", "false", "try", "catch", "throw", "defer", "test", "assert":
				kind = t
			}
			s.emit(kind, t, p)
			continue
		}
		if c >= '0' && c <= '9' {
			for s.peek(0) >= '0' && s.peek(0) <= '9' {
				s.next()
			}
			kind := "integer"
			if s.peek(0) == '.' && s.peek(1) >= '0' && s.peek(1) <= '9' {
				kind = "float"
				s.next()
				for s.peek(0) >= '0' && s.peek(0) <= '9' {
					s.next()
				}
			}
			if s.peek(0) == 'e' || s.peek(0) == 'E' {
				kind = "float"
				s.next()
				if s.peek(0) == '+' || s.peek(0) == '-' {
					s.next()
				}
				n := s.i
				for s.peek(0) >= '0' && s.peek(0) <= '9' {
					s.next()
				}
				if n == s.i {
					return nil, diagnostics.New(p, "syntax error: exponent requires digits")
				}
			}
			s.emit(kind, string(s.input[start:s.i]), p)
			continue
		}
		if c == '"' {
			s.next()
			closed := false
			for s.i < len(s.input) {
				r := s.next()
				if r == '\n' || r == '\r' {
					break
				}
				if r == '"' {
					closed = true
					break
				}
				if r == '\\' {
					if s.peek(0) == '\n' || s.peek(0) == '\r' {
						break
					}
					s.next()
				}
			}
			if !closed {
				return nil, diagnostics.New(p, "syntax error: unterminated string; add a closing quote")
			}
			raw := string(s.input[start:s.i])
			value, err := strconv.Unquote(raw)
			if err != nil {
				return nil, diagnostics.New(p, "syntax error: invalid string escape")
			}
			s.emit("string", value, p)
			continue
		}
		pair := string([]rune{c, s.peek(1)})
		switch pair {
		case "==", "!=", ">=", "<=", "&&", "||", "+=", "-=", "*=", "/=", "%=", "..":
			s.next()
			s.next()
			s.emit(pair, pair, p)
			continue
		}
		switch c {
		case '?', '+', '-', '*', '/', '%', '=', '!', '>', '<', '(', ')', '[', ']', '.', '{', '}', ',', ':', ';':
			s.next()
			if c == '(' || c == '[' {
				s.parens++
			}
			if (c == ')' || c == ']') && s.parens > 0 {
				s.parens--
			}
			s.emit(string(c), string(c), p)
		default:
			return nil, diagnostics.New(p, "syntax error: unexpected character %q", c)
		}
	}
	s.emit("eof", "", s.pos())
	return s.tokens, nil
}
func (s *scanner) peek(n int) rune {
	if s.i+n >= len(s.input) {
		return 0
	}
	return s.input[s.i+n]
}
func (s *scanner) next() rune {
	if s.i >= len(s.input) {
		return 0
	}
	c := s.input[s.i]
	s.i++
	if c == '\n' {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	return c
}
func (s *scanner) pos() diagnostics.Position {
	return diagnostics.Position{File: s.file, Line: s.line, Column: s.col}
}
func (s *scanner) emit(k, t string, p diagnostics.Position) {
	s.tokens = append(s.tokens, Token{Kind: k, Text: t, Pos: p})
}
