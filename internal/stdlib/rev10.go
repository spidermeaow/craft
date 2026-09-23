package stdlib

import (
	"craft/internal/ast"
	rt "craft/internal/runtime"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	signatures["std.log.setLevel"] = Signature{Params: []ast.Type{ast.String}, Result: ast.Void}
	signatures["std.log.write"] = Signature{Params: []ast.Type{ast.String, ast.String, ast.String.Map(), ast.String.Array()}, Result: ast.Void}
	signatures["std.config.required"] = Signature{Params: []ast.Type{ast.String, ast.String.Optional()}, Result: ast.String}
	signatures["std.config.defaultValue"] = Signature{Params: []ast.Type{ast.String.Optional(), ast.String}, Result: ast.String}
	signatures["std.config.intRange"] = Signature{Params: []ast.Type{ast.String, ast.String, ast.Int, ast.Int}, Result: ast.Int}
	signatures["std.config.bool"] = Signature{Params: []ast.Type{ast.String, ast.String}, Result: ast.Bool}
}

func toolingError(kind, message string) error { return &rt.Exception{Kind: kind, Message: message} }

func configInvoke(name string, a []any) (any, error) {
	fail := func(message string) (any, error) {
		return nil, toolingError("ConfigError", "setting "+strconv.Quote(a[0].(string))+": "+message)
	}
	switch name {
	case "std.config.defaultValue":
		v := a[0].(rt.Optional)
		if v.Valid {
			return v.Value, nil
		}
		return a[1], nil
	case "std.config.required":
		v := a[1].(rt.Optional)
		if !v.Valid || strings.TrimSpace(v.Value.(string)) == "" {
			return fail("required nonblank value")
		}
		return v.Value, nil
	case "std.config.intRange":
		lo, hi := a[2].(int64), a[3].(int64)
		if lo > hi {
			return fail("invalid range bounds")
		}
		v, err := strconv.ParseInt(a[1].(string), 10, 64)
		if err != nil || v < lo || v > hi {
			return fail("expected decimal Int within configured range")
		}
		return v, nil
	case "std.config.bool":
		if a[1] == "true" {
			return true, nil
		}
		if a[1] == "false" {
			return false, nil
		}
		return fail("expected true or false")
	}
	return nil, toolingError("ConfigError", "unknown configuration operation")
}

type sessionLogger struct {
	mu    sync.Mutex
	out   io.Writer
	level int
}

func logLevel(s string) int {
	switch s {
	case "debug":
		return 0
	case "info":
		return 1
	case "warn":
		return 2
	case "error":
		return 3
	}
	return -1
}
func (s *Session) SetLogWriter(out io.Writer) {
	s.logger.mu.Lock()
	defer s.logger.mu.Unlock()
	if out == nil {
		out = io.Discard
	}
	s.logger.out = out
}

// Fork shares the execution logger but not input or program arguments.
func (s *Session) Fork() *Session {
	child := NewSession(nil, nil)
	child.logger = s.logger
	return child
}
func (l *sessionLogger) invoke(name string, a []any) (any, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	level := logLevel(a[0].(string))
	if level < 0 {
		return nil, toolingError("LogError", "invalid log level")
	}
	if name == "std.log.setLevel" {
		l.level = level
		return nil, nil
	}
	message := a[1].(string)
	fields := a[2].(*rt.Map)
	sensitive := a[3].(*rt.Array)
	if len(message) > 4096 || len(fields.Keys) > 64 || len(sensitive.Items) > 64 {
		return nil, toolingError("LogError", "log limits exceeded")
	}
	hidden := map[string]bool{}
	for _, k := range sensitive.Items {
		hidden[k.(string)] = true
	}
	values := map[string]string{}
	for _, k := range fields.Keys {
		v := fields.Values[k].(string)
		if len(k) > 128 || len(v) > 4096 {
			return nil, toolingError("LogError", "field limits exceeded")
		}
		if hidden[k] {
			v = "[REDACTED]"
		}
		values[k] = v
	}
	if level < l.level {
		return nil, nil
	}
	record := struct {
		Time    string            `json:"time"`
		Level   string            `json:"level"`
		Message string            `json:"message"`
		Fields  map[string]string `json:"fields"`
	}{time.Now().UTC().Format(time.RFC3339Nano), a[0].(string), message, values}
	b, err := json.Marshal(record)
	if err != nil {
		return nil, toolingError("LogError", "could not encode log")
	}
	b = append(b, '\n')
	n, err := l.out.Write(b)
	if err != nil || n != len(b) {
		return nil, toolingError("LogError", "could not write complete log record")
	}
	return nil, nil
}
