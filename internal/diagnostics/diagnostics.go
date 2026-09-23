package diagnostics

import (
	"errors"
	"fmt"
	"strings"
)

type Position struct {
	File         string
	Line, Column int
}
type Error struct {
	Pos     Position
	Message string
	Code    string
	Exit    int
	Frames  []Frame
}
type Frame struct {
	Name string
	Pos  Position
}

func (e *Error) Error() string {
	if e.Pos.File == "" {
		return e.Code + " " + e.Message
	}
	return fmt.Sprintf("%s:%d:%d: %s %s", e.Pos.File, e.Pos.Line, e.Pos.Column, e.Code, e.Message)
}
func New(p Position, format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	code, exit := "E3001", 1
	switch {
	case strings.HasPrefix(message, "syntax error"):
		code, exit = "E1001", 2
	case strings.HasPrefix(message, "entry point error"):
		code, exit = "E4001", 4
	case strings.HasPrefix(message, "manifest error"):
		code, exit = "E4002", 4
	case strings.HasPrefix(message, "type error"):
		code, exit = "E2002", 3
		if strings.Contains(message, "unknown variable") || strings.Contains(message, "unknown function") {
			code = "E2001"
		}
		if strings.Contains(message, "immutable") {
			code = "E2003"
		}
	}
	return &Error{Pos: p, Message: message, Code: code, Exit: exit}
}
func ExitCode(err error) int {
	var d *Error
	if errors.As(err, &d) {
		return d.Exit
	}
	return 1
}
func Format(err error, sources map[string]string) string {
	var e *Error
	if !errors.As(err, &e) {
		return err.Error()
	}
	trace := ""
	for _, f := range e.Frames {
		trace += fmt.Sprintf("\n  at %s (%s:%d:%d)", f.Name, f.Pos.File, f.Pos.Line, f.Pos.Column)
	}
	source, exists := sources[e.Pos.File]
	if !exists {
		return e.Error() + trace
	}
	lines := strings.Split(source, "\n")
	if e.Pos.Line < 1 || e.Pos.Line > len(lines) {
		return e.Error() + trace
	}
	line := strings.TrimSuffix(lines[e.Pos.Line-1], "\r")
	prefix := []rune(line)
	col := e.Pos.Column - 1
	if col > len(prefix) {
		col = len(prefix)
	}
	if col < 0 {
		col = 0
	}
	indent := ""
	for _, r := range prefix[:col] {
		if r == '\t' {
			indent += "    "
		} else {
			indent += " "
		}
	}
	return fmt.Sprintf("%s\n  %s\n  %s^%s", e.Error(), strings.ReplaceAll(line, "\t", "    "), indent, trace)
}
