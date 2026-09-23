package stdlib

import (
	"bufio"
	"context"
	"craft/internal/ast"
	rt "craft/internal/runtime"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf8"
)

func init() {
	add := func(n string, r ast.Type, p ...ast.Type) { signatures[n] = Signature{Params: p, Result: r} }
	add("std.io.write", ast.Void, ast.String)
	add("std.io.readLine", ast.String.Optional())
	add("std.env.args", ast.String.Array())
	add("std.env.lookup", ast.String.Optional(), ast.String)
	for _, n := range []string{"toInt", "tryInt"} {
		r := ast.Int
		if n == "tryInt" {
			r = r.Optional()
		}
		add("std.convert."+n, r, ast.String)
	}
	for _, n := range []string{"toFloat", "tryFloat"} {
		r := ast.Float
		if n == "tryFloat" {
			r = r.Optional()
		}
		add("std.convert."+n, r, ast.String)
	}
	add("std.convert.toBool", ast.Bool, ast.String)
	add("std.convert.toString", ast.String, ast.String)
	add("std.string.fixed", ast.String, ast.Float, ast.Int)
	add("std.string.join", ast.String, ast.String.Array(), ast.String)
	for _, n := range []string{"round", "floor", "ceil"} {
		add("std.math."+n, ast.Float, ast.Float)
	}
	add("std.math.pow", ast.Float, ast.Float, ast.Float)
	add("std.math.random", ast.Random, ast.Int)
	add("std.fs.appendText", ast.Void, ast.String, ast.String)
	add("std.fs.createDirectory", ast.Void, ast.String)
	add("std.fs.listDirectory", ast.String.Array(), ast.String)
	add("std.fs.isDirectory", ast.Bool, ast.String)
	add("std.fs.removeFile", ast.Void, ast.String)
	add("std.fs.copyFile", ast.Void, ast.String, ast.String)
	add("std.fs.moveFile", ast.Void, ast.String, ast.String)
	add("std.fs.writeTextAtomic", ast.Void, ast.String, ast.String)
	add("std.fs.readBytes", ast.Bytes, ast.String)
	add("std.fs.writeBytesAtomic", ast.Void, ast.String, ast.Bytes)
	add("std.path.join", ast.String, ast.String.Array())
	add("std.path.clean", ast.String, ast.String)
	add("std.path.absolute", ast.String, ast.String)
	add("std.path.relative", ast.String, ast.String, ast.String)
	for _, n := range []string{"fileName", "extension", "parent"} {
		add("std.path."+n, ast.String, ast.String)
	}
	add("std.path.isAbsolute", ast.Bool, ast.String)
	registerJSON(add)
	registerTime(add)
}

// Session owns program input and arguments. No global reader or random state.
type Session struct {
	logger   *sessionLogger
	reader   *bufio.Reader
	args     []string
	canceled atomic.Bool
	reading  atomic.Bool
}

func NewSession(in io.Reader, args []string) *Session {
	if in == nil {
		in = strings.NewReader("")
	}
	return &Session{reader: bufio.NewReader(in), args: append([]string(nil), args...), logger: &sessionLogger{out: os.Stderr, level: 1}}
}
func (s *Session) Invoke(ctx context.Context, out io.Writer, name string, args []any) (any, error) {
	if strings.HasPrefix(name, "std.log.") {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return s.logger.invoke(name, args)
	}
	switch name {
	case "std.io.readLine":
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		if s.canceled.Load() {
			return nil, fmt.Errorf("console input unavailable after a cancelled read; start a new run")
		}
		if !s.reading.CompareAndSwap(false, true) {
			return nil, fmt.Errorf("console input already has a reader")
		}
		type result struct {
			v any
			e error
		}
		done := make(chan result, 1)
		go func() { v, e := readLine(s.reader); s.reading.Store(false); done <- result{v, e} }()
		select {
		case <-ctx.Done():
			s.canceled.Store(true)
			return nil, ctx.Err()
		case r := <-done:
			return r.v, r.e
		}
	case "std.env.args":
		return stringArray(s.args), nil
	}
	return Invoke(ctx, out, name, args)
}

const MaxLineBytes = 1024 * 1024

func readLine(r *bufio.Reader) (any, error) {
	data := []byte{}
	tooLong := false
	for {
		part, e := r.ReadSlice('\n')
		if len(data)+len(part) > MaxLineBytes+2 {
			tooLong = true
		} else if !tooLong {
			data = append(data, part...)
		}
		if e == bufio.ErrBufferFull {
			continue
		}
		if e != nil && e != io.EOF {
			return nil, e
		}
		if len(data) == 0 && e == io.EOF && !tooLong {
			return rt.Optional{}, nil
		}
		if len(data) > 0 && data[len(data)-1] == '\n' {
			data = data[:len(data)-1]
			if len(data) > 0 && data[len(data)-1] == '\r' {
				data = data[:len(data)-1]
			}
		}
		if tooLong || len(data) > MaxLineBytes {
			return nil, fmt.Errorf("readLine limit is 1 MiB; line discarded")
		}
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("readLine requires UTF-8 input")
		}
		return rt.Some(string(data)), nil
	}
}

var decimalFloat = regexp.MustCompile(`^[+-]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][+-]?[0-9]+)?$`)

func finite(v float64) (any, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil, fmt.Errorf("numeric result must be finite")
	}
	return v, nil
}
func stringArray(values []string) *rt.Array {
	a := &rt.Array{}
	for _, v := range values {
		a.Items = append(a.Items, v)
	}
	return a
}
func arrayStrings(a *rt.Array) []string {
	values := make([]string, len(a.Items))
	for i, v := range a.Items {
		values[i] = v.(string)
	}
	return values
}
func invokeRev2(out io.Writer, name string, args []any) (any, error) {
	switch name {
	case "std.io.write":
		_, e := io.WriteString(out, args[0].(string))
		return nil, e
	case "std.env.lookup":
		v, ok := os.LookupEnv(args[0].(string))
		if !ok {
			return rt.Optional{}, nil
		}
		return rt.Some(v), nil
	case "std.convert.toInt", "std.convert.tryInt":
		v, e := strconv.ParseInt(args[0].(string), 10, 64)
		if name == "std.convert.tryInt" {
			if e != nil {
				return rt.Optional{}, nil
			}
			return rt.Some(v), nil
		}
		if e != nil {
			return nil, fmt.Errorf("invalid Int: %w", e)
		}
		return v, nil
	case "std.convert.toFloat", "std.convert.tryFloat":
		text := args[0].(string)
		v, e := strconv.ParseFloat(text, 64)
		if !decimalFloat.MatchString(text) || math.IsNaN(v) || math.IsInf(v, 0) {
			e = fmt.Errorf("expected a finite decimal Float")
		}
		if name == "std.convert.tryFloat" {
			if e != nil {
				return rt.Optional{}, nil
			}
			return rt.Some(v), nil
		}
		if e != nil {
			return nil, fmt.Errorf("invalid Float: %w", e)
		}
		return v, nil
	case "std.convert.toBool":
		if args[0] == "true" {
			return true, nil
		}
		if args[0] == "false" {
			return false, nil
		}
		return nil, fmt.Errorf("Bool text must be true or false")
	case "std.convert.toString":
		return fmt.Sprint(args[0]), nil
	case "std.string.fixed":
		n := args[1].(int64)
		if n < 0 || n > 18 {
			return nil, fmt.Errorf("fixed digits must be between 0 and 18")
		}
		return strconv.FormatFloat(args[0].(float64), 'f', int(n), 64), nil
	case "std.string.join":
		return strings.Join(arrayStrings(args[0].(*rt.Array)), args[1].(string)), nil
	case "std.math.round":
		return math.Round(args[0].(float64)), nil
	case "std.math.floor":
		return math.Floor(args[0].(float64)), nil
	case "std.math.ceil":
		return math.Ceil(args[0].(float64)), nil
	case "std.math.pow":
		return finite(math.Pow(args[0].(float64), args[1].(float64)))
	case "std.math.random":
		return &rt.Random{State: uint64(args[0].(int64))}, nil
	case "std.path.join":
		return filepath.Join(arrayStrings(args[0].(*rt.Array))...), nil
	case "std.path.clean":
		return filepath.Clean(args[0].(string)), nil
	case "std.path.absolute":
		path, e := filepath.Abs(args[0].(string))
		if e != nil {
			return nil, fsFailure("make absolute", args[0].(string), e)
		}
		return path, nil
	case "std.path.relative":
		path, e := filepath.Rel(args[0].(string), args[1].(string))
		if e != nil {
			return nil, toolingError("FsPathError", fmt.Sprintf("path relative from %q to %q failed", args[0].(string), args[1].(string)))
		}
		return path, nil
	case "std.path.fileName":
		return filepath.Base(args[0].(string)), nil
	case "std.path.extension":
		return filepath.Ext(args[0].(string)), nil
	case "std.path.parent":
		return filepath.Dir(args[0].(string)), nil
	case "std.path.isAbsolute":
		return filepath.IsAbs(args[0].(string)), nil
	}
	if strings.HasPrefix(name, "std.fs.") {
		return filesystem(name, args)
	}
	if strings.HasPrefix(name, "std.json.") {
		return jsonInvoke(name, args)
	}
	if strings.HasPrefix(name, "std.time.") {
		return timeInvoke(name, args)
	}
	return nil, fmt.Errorf("unknown standard library function %s", name)
}

func Mutates(t ast.Type, name string) bool {
	return (t.IsArray() && (name == "append" || name == "remove")) || (t.IsMap() && (name == "set" || name == "remove")) || t == ast.Random
}
func methodRev2(t ast.Type, name string) (Signature, bool) {
	sig := func(r ast.Type, p ...ast.Type) (Signature, bool) { return Signature{Params: p, Result: r}, true }
	if t == ast.String {
		switch name {
		case "replace":
			return sig(ast.String, ast.String, ast.String)
		case "substring":
			return sig(ast.String, ast.Int, ast.Int)
		}
	}
	if t.IsArray() {
		switch name {
		case "contains", "indexOf":
			if t.Element().Primitive() {
				r := ast.Bool
				if name == "indexOf" {
					r = ast.Int.Optional()
				}
				return sig(r, t.Element())
			}
		case "slice":
			return sig(t, ast.Int, ast.Int)
		case "reverse":
			return sig(t)
		case "sorted":
			if e := t.Element(); e == ast.Int || e == ast.Float || e == ast.String {
				return sig(t)
			}
		}
	}
	if t.IsMap() {
		e := t.MapElement()
		switch name {
		case "get":
			return sig(e.Optional(), ast.String)
		case "require":
			return sig(e, ast.String)
		case "set":
			return sig(ast.Void, ast.String, e)
		case "remove", "containsKey":
			return sig(ast.Bool, ast.String)
		case "keys":
			return sig(ast.String.Array())
		case "values":
			return sig(e.Array())
		}
	}
	if t == ast.Random {
		switch name {
		case "nextInt":
			return sig(ast.Int, ast.Int, ast.Int)
		case "nextFloat":
			return sig(ast.Float)
		}
	}
	if t == ast.JsonValue {
		return jsonMethodSignature(name)
	}
	if t == ast.DateTime || t == ast.Duration {
		return timeMethodSignature(t, name)
	}
	return Signature{}, false
}
func sliceBounds(args []any, n int) (int, int, error) {
	a, b := args[0].(int64), args[1].(int64)
	if a < 0 || b < a || b > int64(n) {
		return 0, 0, fmt.Errorf("slice bounds [%d, %d) invalid for length %d", a, b, n)
	}
	return int(a), int(b), nil
}
func arrayRev2(a *rt.Array, name string, args []any) (any, error) {
	switch name {
	case "contains", "indexOf":
		for i, v := range a.Items {
			if v == args[0] {
				if name == "contains" {
					return true, nil
				}
				return rt.Some(int64(i)), nil
			}
		}
		if name == "contains" {
			return false, nil
		}
		return rt.Optional{}, nil
	case "slice":
		lo, hi, e := sliceBounds(args, len(a.Items))
		if e != nil {
			return nil, e
		}
		return rt.Clone(&rt.Array{Items: a.Items[lo:hi]}), nil
	case "reverse":
		b := rt.Clone(a).(*rt.Array)
		for i, j := 0, len(b.Items)-1; i < j; i, j = i+1, j-1 {
			b.Items[i], b.Items[j] = b.Items[j], b.Items[i]
		}
		return b, nil
	case "sorted":
		b := rt.Clone(a).(*rt.Array)
		sort.SliceStable(b.Items, func(i, j int) bool {
			switch x := b.Items[i].(type) {
			case int64:
				return x < b.Items[j].(int64)
			case float64:
				return x < b.Items[j].(float64)
			case string:
				return x < b.Items[j].(string)
			}
			return false
		})
		return b, nil
	}
	return nil, fmt.Errorf("unknown Array method %s", name)
}
func ValueMethod(v any, name string, args []any) (any, error) {
	switch v.(type) {
	case Bytes, *Context:
		return rev4ValueMethod(v, name)
	}
	switch x := v.(type) {
	case *rt.Map:
		switch name {
		case "set":
			x.Set(args[0].(string), rt.Clone(args[1]))
			return nil, nil
		case "get", "require":
			v, ok := x.Values[args[0].(string)]
			if name == "get" {
				if !ok {
					return rt.Optional{}, nil
				}
				return rt.Some(v), nil
			}
			if !ok {
				return nil, fmt.Errorf("Map key %q not found", args[0])
			}
			return rt.Clone(v), nil
		case "containsKey":
			_, ok := x.Values[args[0].(string)]
			return ok, nil
		case "remove":
			k := args[0].(string)
			if _, ok := x.Values[k]; !ok {
				return false, nil
			}
			delete(x.Values, k)
			for i, v := range x.Keys {
				if v == k {
					x.Keys = append(x.Keys[:i], x.Keys[i+1:]...)
					break
				}
			}
			return true, nil
		case "keys":
			return stringArray(x.Keys), nil
		case "values":
			a := &rt.Array{}
			for _, k := range x.Keys {
				a.Items = append(a.Items, rt.Clone(x.Values[k]))
			}
			return a, nil
		}
	case *rt.Random:
		if name == "nextFloat" {
			return float64(x.Next()>>11) * (1.0 / (1 << 53)), nil
		}
		lo, hi := args[0].(int64), args[1].(int64)
		if hi <= lo {
			return nil, fmt.Errorf("random range requires min < max")
		}
		width := uint64(hi) - uint64(lo)
		threshold := -width % width
		for {
			n := x.Next()
			if n >= threshold {
				return int64(uint64(lo) + n%width), nil
			}
		}
	case JSON:
		return jsonMethod(x, name, args)
	case DateTime:
		return dateMethod(x, name, args)
	case Duration:
		return durationMethod(x, name, args)
	}
	return nil, fmt.Errorf("unknown method %s", name)
}
