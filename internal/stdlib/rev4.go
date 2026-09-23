package stdlib

import (
	"context"
	"craft/internal/ast"
	rt "craft/internal/runtime"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// Bytes is immutable. Its backing string never escapes as mutable Go memory.
type Bytes struct{ data string }

func (b Bytes) String() string { return fmt.Sprintf("Bytes(%d)", len(b.data)) }

const MaxBodyBytes = 16 * 1024 * 1024

type Context struct {
	ctx      context.Context
	expired  atomic.Bool
	stop     func()
	lifetime *Context
}

func (c *Context) isExpired() bool {
	return c.expired.Load() || c.lifetime != nil && c.lifetime.expired.Load()
}

type requestContextKey struct{}

func validRequestPath(u *url.URL) bool {
	escaped := strings.ToLower(u.EscapedPath())
	return strings.HasPrefix(u.Path, "/") && !strings.Contains(escaped, "%2f") && !strings.Contains(escaped, "%5c") && !strings.ContainsAny(u.Path, "\\\x00")
}

func (c *Context) String() string { return "Context" }
func NativeStructs() []*ast.Struct {
	fields := func(names []string, types []ast.Type) *ast.Struct {
		s := &ast.Struct{}
		for i, n := range names {
			s.Fields = append(s.Fields, ast.Field{Name: n, Type: types[i], Mutable: true})
		}
		return s
	}
	request := fields([]string{"method", "rawTarget", "path", "query", "headers", "body", "context"}, []ast.Type{ast.String, ast.String, ast.String, ast.String.Array().Map(), ast.String.Array().Map(), ast.Bytes, ast.Context})
	request.Name = "HttpRequest"
	response := fields([]string{"status", "headers", "body"}, []ast.Type{ast.Int, ast.String.Array().Map(), ast.Bytes})
	response.Name = "HttpResponse"
	config := fields([]string{"address", "maxConcurrent", "maxConnections", "maxBodyBytes", "maxHeaderBytes", "readTimeoutMs", "writeTimeoutMs", "idleTimeoutMs", "requestTimeoutMs", "shutdownTimeoutMs"}, []ast.Type{ast.String, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int})
	config.Name = "HttpConfig"
	return append([]*ast.Struct{request, response, config}, dbStructs()...)
}
func nativeStruct(name string, values ...any) *rt.Struct {
	for _, s := range NativeStructs() {
		if s.Name == name {
			v := &rt.Struct{Name: name, Fields: map[string]any{}}
			for i, f := range s.Fields {
				v.Order = append(v.Order, f.Name)
				v.Fields[f.Name] = values[i]
			}
			return v
		}
	}
	panic("unknown native struct")
}
func init() {
	add := func(n string, r ast.Type, p ...ast.Type) { signatures[n] = Signature{Params: p, Result: r} }
	add("std.bytes.fromString", ast.Bytes, ast.String)
	add("std.bytes.fromArray", ast.Bytes, ast.Int.Array())
	add("std.context.current", ast.Context)
	add("std.http.config", "HttpConfig", ast.String)
	add("std.http.response", "HttpResponse", ast.Int, ast.Bytes)
	add("std.http.request", "HttpRequest", ast.String, ast.String, ast.Bytes)
	add("std.http.stop", ast.Void, ast.Context)
	signatures["std.http.serve"] = Signature{Result: ast.Void, Variadic: true}
}
func rev4Method(t ast.Type, n string) (Signature, bool) {
	if t == ast.Bytes {
		switch n {
		case "length":
			return Signature{Result: ast.Int}, true
		case "text":
			return Signature{Result: ast.String}, true
		case "toArray":
			return Signature{Result: ast.Int.Array()}, true
		}
	}
	if t == ast.Context {
		switch n {
		case "cancelled":
			return Signature{Result: ast.Bool}, true
		case "reason":
			return Signature{Result: ast.String}, true
		case "remainingMilliseconds":
			return Signature{Result: ast.Int.Optional()}, true
		}
	}
	return Signature{}, false
}
func rev4ValueMethod(v any, n string) (any, error) {
	switch b := v.(type) {
	case Bytes:
		switch n {
		case "length":
			return int64(len(b.data)), nil
		case "text":
			if !utf8.ValidString(b.data) {
				return nil, fmt.Errorf("body is not UTF-8")
			}
			return b.data, nil
		case "toArray":
			a := &rt.Array{}
			for _, v := range []byte(b.data) {
				a.Items = append(a.Items, int64(v))
			}
			return a, nil
		}
	case *Context:
		if b.isExpired() {
			return nil, fmt.Errorf("request context has expired")
		}
		switch n {
		case "cancelled":
			return b.ctx.Err() != nil, nil
		case "reason":
			if e := context.Cause(b.ctx); e != nil {
				return e.Error(), nil
			}
			return "", nil
		case "remainingMilliseconds":
			if d, ok := b.ctx.Deadline(); ok {
				n := time.Until(d).Milliseconds()
				if n < 0 {
					n = 0
				}
				return rt.Some(n), nil
			}
			return rt.Optional{}, nil
		}
	}
	return nil, fmt.Errorf("unknown native method %s", n)
}
func invokeRev4(ctx context.Context, name string, args []any) (any, error) {
	switch name {
	case "std.bytes.fromString":
		s := args[0].(string)
		if len(s) > MaxBodyBytes {
			return nil, fmt.Errorf("Bytes limit is 16 MiB")
		}
		return Bytes{s}, nil
	case "std.bytes.fromArray":
		a := args[0].(*rt.Array)
		if len(a.Items) > MaxBodyBytes {
			return nil, fmt.Errorf("Bytes limit is 16 MiB")
		}
		b := make([]byte, len(a.Items))
		for i, v := range a.Items {
			n := v.(int64)
			if n < 0 || n > 255 {
				return nil, fmt.Errorf("byte must be 0..255")
			}
			b[i] = byte(n)
		}
		return Bytes{string(b)}, nil
	case "std.context.current":
		if request, ok := ctx.Value(requestContextKey{}).(*Context); ok {
			return &Context{ctx: ctx, stop: request.stop, lifetime: request}, nil
		}
		return &Context{ctx: ctx}, nil
	case "std.http.config":
		return nativeStruct("HttpConfig", args[0], int64(64), int64(128), int64(1024*1024), int64(32768), int64(10000), int64(15000), int64(30000), int64(10000), int64(5000)), nil
	case "std.http.response":
		return nativeStruct("HttpResponse", args[0], rt.NewMap(), args[1]), nil
	case "std.http.request":
		u, e := url.ParseRequestURI(args[1].(string))
		if e != nil {
			return nil, e
		}
		if !validRequestPath(u) {
			return nil, fmt.Errorf("invalid request path")
		}
		q, e := url.ParseQuery(u.RawQuery)
		if e != nil {
			return nil, e
		}
		return nativeStruct("HttpRequest", args[0], args[1], u.Path, multiMap(q, false), rt.NewMap(), args[2], &Context{ctx: ctx}), nil
	case "std.http.stop":
		c := args[0].(*Context)
		if c.isExpired() {
			return nil, fmt.Errorf("request context has expired")
		}
		if c.stop == nil {
			return nil, fmt.Errorf("context does not belong to a server")
		}
		c.stop()
		return nil, nil
	}
	return nil, fmt.Errorf("unknown native function %s", name)
}
func multiMap(values map[string][]string, lower bool) *rt.Map {
	m := rt.NewMap()
	for _, k := range sortedKeysAny(values) {
		key := k
		if lower {
			key = strings.ToLower(k)
		}
		a := &rt.Array{}
		if old, ok := m.Values[key]; ok {
			a = old.(*rt.Array)
		}
		for _, v := range values[k] {
			a.Items = append(a.Items, v)
		}
		m.Set(key, a)
	}
	return m
}
func sortedKeysAny(values map[string][]string) []string {
	keys := []string{}
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
