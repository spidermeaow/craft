package interpreter

import (
	"context"
	"craft/internal/ast"
	"craft/internal/diagnostics"
	rt "craft/internal/runtime"
	"craft/internal/stdlib"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type cell struct{ value any }
type environment struct {
	parent *environment
	values map[string]*cell
}

func local(p *environment) *environment { return &environment{parent: p, values: map[string]*cell{}} }
func (e *environment) find(n string) *cell {
	for ; e != nil; e = e.parent {
		if v, ok := e.values[n]; ok {
			return v
		}
	}
	return nil
}
func (e *environment) bind(n string, v any) { e.values[n] = &cell{value: rt.Clone(v)} }

// Capture lexical bindings now, but observe later updates to existing mutable cells.
func (e *environment) capture() *environment {
	if e == nil {
		return nil
	}
	copy := local(e.parent.capture())
	for k, v := range e.values {
		copy.values[k] = v
	}
	return copy
}

type flow struct {
	kind  string
	value any
}
type deferred struct {
	body *ast.Stmt
	env  *environment
}
type frame struct {
	name     string
	caller   diagnostics.Position
	cleanups []deferred
	children []*task
}
type machine struct {
	ctx            context.Context
	out            io.Writer
	structs        map[string]*ast.Struct
	library        *stdlib.Session
	functions      map[string]*ast.Function
	frames         []*frame
	tasks          *taskGroup
	self           *task
	cleaning       bool
	steps          uint64
	servingRequest bool
}

func newMachine(ctx context.Context, p *ast.Program, out io.Writer) *machine {
	m := &machine{ctx: ctx, out: &lockedWriter{out: out}, tasks: newTaskGroup(), functions: map[string]*ast.Function{}, structs: map[string]*ast.Struct{}, library: stdlib.NewSession(nil, nil)}
	for _, s := range stdlib.NativeStructs() {
		m.structs[s.Name] = s
	}
	for _, s := range p.Structs {
		m.structs[s.Name] = s
	}
	for _, f := range p.Functions {
		m.functions[f.Name] = f
	}
	return m
}

// Run requires a checked program. Tests use separate machines and do not invoke main.
func Run(ctx context.Context, p *ast.Program, out io.Writer) error {
	return RunWithInput(ctx, p, nil, out, nil)
}
func RunWithInput(ctx context.Context, p *ast.Program, in io.Reader, out io.Writer, args []string) error {
	return RunWithWriters(ctx, p, in, out, nil, args)
}
func RunWithWriters(ctx context.Context, p *ast.Program, in io.Reader, out, logOut io.Writer, args []string) error {
	m := newMachine(ctx, p, out)
	m.library = stdlib.NewSession(in, args)
	if logOut != nil {
		m.library.SetLogWriter(logOut)
	}
	f := m.functions["main"]
	_, e := m.call(f, nil, f.Pos)
	return e
}
func RunTest(ctx context.Context, p *ast.Program, t *ast.Test, out io.Writer) error {
	return RunTestWithWriters(ctx, p, t, out, nil)
}
func RunTestWithWriters(ctx context.Context, p *ast.Program, t *ast.Test, out, logOut io.Writer) error {
	m := newMachine(ctx, p, out)
	if logOut != nil {
		m.library.SetLogWriter(logOut)
	}
	_, e := m.call(&ast.Function{Name: "test " + t.Name, Return: ast.Void, Body: t.Body, Pos: t.Pos}, nil, t.Pos)
	return e
}
func (m *machine) trace(pos diagnostics.Position) []diagnostics.Frame {
	frames := []diagnostics.Frame{}
	for i := len(m.frames) - 1; i >= 0; i-- {
		f := m.frames[i]
		frames = append(frames, diagnostics.Frame{Name: f.name, Pos: pos})
		pos = f.caller
	}
	return frames
}
func (m *machine) exception(kind, message, code string, p diagnostics.Position) *rt.Exception {
	return &rt.Exception{Kind: kind, Message: message, Detail: &diagnostics.Error{Pos: p, Message: kind + ": " + message, Code: code, Exit: 1, Frames: m.trace(p)}}
}
func (m *machine) wrap(pos diagnostics.Position, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var ex *rt.Exception
	if errors.As(err, &ex) {
		if ex.Detail == nil && strings.HasPrefix(ex.Kind, "Db") {
			return m.exception(ex.Kind, ex.Message, "E3001", pos)
		}
		return ex
	}
	var d *diagnostics.Error
	if errors.As(err, &d) {
		if d.Pos.File != "" {
			pos = d.Pos
		}
		return m.exception("RuntimeError", strings.TrimPrefix(d.Message, "runtime error: "), "E3001", pos)
	}
	return m.exception("RuntimeError", err.Error(), "E3001", pos)
}
func (m *machine) call(f *ast.Function, args []any, at diagnostics.Position) (any, error) {
	if len(m.frames) >= 256 {
		return nil, m.exception("RuntimeError", "call depth exceeds 256; check recursive calls", "E3001", at)
	}
	parentDBContext := m.ctx
	var closeDBScope func()
	m.ctx, closeDBScope = stdlib.BeginDBScope(m.ctx)
	defer func() { closeDBScope(); m.ctx = parentDBContext }()
	fr := &frame{name: f.Name, caller: at}
	m.frames = append(m.frames, fr)
	defer func() { m.frames = m.frames[:len(m.frames)-1] }()
	env := local(nil)
	for i, a := range f.Params {
		env.bind(a.Name, args[i])
	}
	r, err := m.statement(f.Body, env)
	if e := m.finishChildren(fr); err == nil {
		err = e
	}
	// The original error wins; all remaining deferred blocks are still attempted.
	oldContext, oldCleaning := m.ctx, m.cleaning
	cleanupStarted := false
	m.cleaning = true
	defer func() {
		m.ctx, m.cleaning = oldContext, oldCleaning
	}()
	for i := len(fr.cleanups) - 1; i >= 0; i-- {
		// Normal defer retains existing timing. After cancellation, allow a bounded
		// cleanup window; nested cleanup calls inherit it instead of renewing it.
		if !oldCleaning && !cleanupStarted && m.ctx.Err() != nil {
			cleanupContext, cancelCleanup := context.WithTimeout(context.WithoutCancel(oldContext), 5*time.Second)
			defer cancelCleanup()
			m.ctx = cleanupContext
			cleanupStarted = true
		}
		cleanup := fr.cleanups[i]
		_, e := m.statement(cleanup.body, cleanup.env)
		if err == nil && e != nil {
			err = e
		}
	}
	return rt.Clone(r.value), err
}
func (m *machine) statement(s *ast.Stmt, env *environment) (result flow, err error) {
	defer func() { err = m.wrap(s.Pos, err) }()
	if e := m.ctx.Err(); e != nil {
		return flow{}, e
	}
	m.steps++
	if m.steps%64 == 0 {
		m.tasks.yield()
	}
	switch s.Kind {
	case "block":
		env = local(env)
		for _, st := range s.Statements {
			r, e := m.statement(st, env)
			if e != nil || r.kind != "" {
				return r, e
			}
		}
	case "let", "var":
		v, e := m.expression(s.Expr, env)
		if e != nil {
			return flow{}, e
		}
		env.bind(s.Name, v)
	case "expression":
		_, e := m.expression(s.Expr, env)
		return flow{}, e
	case "return":
		var v any
		var e error
		if s.Expr != nil {
			v, e = m.expression(s.Expr, env)
		}
		return flow{kind: "return", value: rt.Clone(v)}, e
	case "break", "continue":
		return flow{kind: s.Kind}, nil
	case "iflet":
		v, e := m.expression(s.Expr, env)
		if e != nil {
			return flow{}, e
		}
		opt := v.(rt.Optional)
		if opt.Valid {
			child := local(env)
			child.bind(s.Name, opt.Value)
			return m.statement(s.Body, child)
		}
		if s.Else != nil {
			return m.statement(s.Else, env)
		}
	case "if":
		v, e := m.expression(s.Expr, env)
		if e != nil {
			return flow{}, e
		}
		if v.(bool) {
			return m.statement(s.Body, env)
		}
		if s.Else != nil {
			return m.statement(s.Else, env)
		}
	case "while":
		for {
			v, e := m.expression(s.Expr, env)
			if e != nil {
				return flow{}, e
			}
			if !v.(bool) {
				break
			}
			r, e := m.statement(s.Body, env)
			if e != nil || r.kind == "return" {
				return r, e
			}
			if r.kind == "break" {
				break
			}
		}
	case "for":
		start, e := m.expression(s.Expr, env)
		if e != nil {
			return flow{}, e
		}
		runIteration := func(v any) (flow, error) {
			iteration := local(env)
			iteration.bind(s.Name, v)
			return m.statement(s.Body, iteration)
		}
		if s.End == nil {
			values := rt.Clone(start).(*rt.Array)
			for _, v := range values.Items {
				r, e := runIteration(v)
				if e != nil || r.kind == "return" {
					return r, e
				}
				if r.kind == "break" {
					break
				}
			}
		} else {
			end, e := m.expression(s.End, env)
			if e != nil {
				return flow{}, e
			}
			for i := start.(int64); i < end.(int64); i++ {
				r, e := runIteration(i)
				if e != nil || r.kind == "return" {
					return r, e
				}
				if r.kind == "break" {
					break
				}
			}
		}
	case "throw":
		v, e := m.expression(s.Expr, env)
		if e != nil {
			return flow{}, e
		}
		ex := v.(*rt.Exception)
		if ex.Detail == nil {
			ex = m.exception(ex.Kind, ex.Message, "E3002", s.Pos)
		}
		return flow{}, ex
	case "try":
		r, e := m.statement(s.Body, env)
		if e == nil {
			return r, nil
		}
		var ex *rt.Exception
		if !errors.As(e, &ex) {
			return r, e
		}
		caught := local(env)
		caught.bind(s.Name, ex)
		return m.statement(s.Else, caught)
	case "assert":
		v, e := m.expression(s.Expr, env)
		if e != nil {
			return flow{}, e
		}
		if !v.(bool) {
			return flow{}, m.exception("AssertionError", "assertion failed", "E3003", s.Pos)
		}
	case "defer":
		fr := m.frames[len(m.frames)-1]
		fr.cleanups = append(fr.cleanups, deferred{body: s.Body, env: env.capture()})
	}
	return flow{}, nil
}
func (m *machine) expression(x *ast.Expr, env *environment) (result any, err error) {
	defer func() { err = m.wrap(x.Pos, err) }()
	if e := m.ctx.Err(); e != nil {
		return nil, e
	}
	switch x.Kind {
	case "none":
		return rt.Optional{}, nil
	case "some":
		v, e := m.expression(x.Left, env)
		if e != nil {
			return nil, e
		}
		return rt.Some(v), nil
	case "literal":
		switch x.Type {
		case ast.Null:
			return nil, nil
		case ast.Int:
			v, _ := strconv.ParseInt(x.Text, 10, 64)
			return v, nil
		case ast.Float:
			v, _ := strconv.ParseFloat(x.Text, 64)
			return v, nil
		case ast.Bool:
			return x.Text == "true", nil
		case ast.String:
			return x.Text, nil
		}
	case "identifier":
		if value := env.find(x.Text); value != nil {
			return value.value, nil
		}
		return rt.Function{Declaration: m.functions[x.Text]}, nil
	case "map":
		a := rt.NewMap()
		for i, item := range x.Args {
			v, e := m.expression(item, env)
			if e != nil {
				return nil, e
			}
			a.Set(x.ArgNames[i], rt.Clone(v))
		}
		return a, nil
	case "array":
		a := &rt.Array{}
		for _, item := range x.Args {
			v, e := m.expression(item, env)
			if e != nil {
				return nil, e
			}
			a.Items = append(a.Items, rt.Clone(v))
		}
		return a, nil
	case "index":
		a, i, e := m.index(x, env)
		if e != nil {
			return nil, e
		}
		return a.Items[i], nil
	case "member":
		v, e := m.expression(x.Left, env)
		if e != nil {
			return nil, e
		}
		switch value := v.(type) {
		case string:
			return int64(utf8.RuneCountInString(value)), nil
		case *rt.Array:
			return int64(len(value.Items)), nil
		case *rt.Map:
			return int64(len(value.Keys)), nil
		case *rt.Struct:
			return value.Fields[x.Text], nil
		case *rt.Exception:
			switch x.Text {
			case "message":
				return value.Message, nil
			case "type":
				return value.Kind, nil
			case "code":
				if value.Detail != nil {
					return value.Detail.Code, nil
				}
				return "E3002", nil
			case "file":
				if value.Detail != nil {
					return value.Detail.Pos.File, nil
				}
				return "", nil
			case "line":
				if value.Detail != nil {
					return int64(value.Detail.Pos.Line), nil
				}
				return int64(0), nil
			case "column":
				if value.Detail != nil {
					return int64(value.Detail.Pos.Column), nil
				}
				return int64(0), nil
			}
		}
	case "unary":
		if x.Text == "-" && x.Right.Kind == "literal" && x.Right.Type == ast.Int && x.Right.Text == "9223372036854775808" {
			return int64(math.MinInt64), nil
		}
		v, e := m.expression(x.Right, env)
		if e != nil {
			return nil, e
		}
		switch x.Text {
		case "!":
			return !v.(bool), nil
		case "+":
			return v, nil
		case "-":
			switch n := v.(type) {
			case int64:
				if n == math.MinInt64 {
					return nil, fmt.Errorf("Int overflow")
				}
				return -n, nil
			case float64:
				return -n, nil
			}
		}
	case "assignment":
		var old any
		var write func(any) error
		if x.Left.Kind == "identifier" {
			target := env.find(x.Left.Text)
			old = target.value
			write = func(v any) error { target.value = rt.Clone(v); return nil }
		} else if x.Left.Kind == "member" {
			value, e := m.expression(x.Left.Left, env)
			if e != nil {
				return nil, e
			}
			st := value.(*rt.Struct)
			old = st.Fields[x.Left.Text]
			write = func(v any) error { st.Fields[x.Left.Text] = rt.Clone(v); return nil }
		} else {
			a, i, e := m.index(x.Left, env)
			if e != nil {
				return nil, e
			}
			old = a.Items[i]
			write = func(v any) error {
				if i >= len(a.Items) {
					return fmt.Errorf("array index changed during assignment")
				}
				a.Items[i] = rt.Clone(v)
				return nil
			}
		}
		v, e := m.expression(x.Right, env)
		if e != nil {
			return nil, e
		}
		if x.Text != "=" {
			v, e = calculate(x.Text[:1], old, v, x.Pos)
			if e != nil {
				return nil, e
			}
		}
		if e = write(v); e != nil {
			return nil, e
		}
		return rt.Clone(v), nil
	case "binary":
		l, e := m.expression(x.Left, env)
		if e != nil {
			return nil, e
		}
		if x.Text == "&&" && !l.(bool) {
			return false, nil
		}
		if x.Text == "||" && l.(bool) {
			return true, nil
		}
		r, e := m.expression(x.Right, env)
		if e != nil {
			return nil, e
		}
		return calculate(x.Text, l, r, x.Pos)
	case "call":
		if _, _, ok := x.Type.Function(); ok {
			v, e := m.expression(x.Left, env)
			if e != nil {
				return nil, e
			}
			args := make([]any, len(x.Args))
			for i, a := range x.Args {
				v, e := m.expression(a, env)
				if e != nil {
					return nil, e
				}
				args[i] = rt.Clone(v)
			}
			return m.call(v.(rt.Function).Declaration, args, x.Pos)
		}
		name := stdlib.QualifiedName(x.Left)
		if name == "std.task.spawn" || name == "std.timer.after" || name == "std.timer.every" {
			return m.startTask(x, env, name)
		}
		_, builtin := stdlib.Lookup(name)
		var receiver any
		if !builtin && x.Left.Kind == "member" {
			var e error
			receiver, e = m.expression(x.Left.Left, env)
			if e != nil {
				return nil, e
			}
		}
		args := make([]any, len(x.Args))
		for i, a := range x.Args {
			v, e := m.expression(a, env)
			if e != nil {
				return nil, e
			}
			slot := i
			if len(x.ArgOrder) > 0 {
				slot = x.ArgOrder[i]
			}
			args[slot] = rt.Clone(v)
		}
		if builtin {
			if name == "std.http.serve" {
				return m.serveHTTP(args, x.Pos)
			}
			if name == "std.time.sleep" {
				d, e := stdlib.SleepDuration(args[0])
				if e != nil {
					return nil, e
				}
				return nil, m.tasks.clock.Wait(m.ctx, d)
			}
			return m.library.Invoke(m.ctx, m.out, name, args)
		}
		if x.Left.Kind == "member" {
			switch value := receiver.(type) {
			case *task:
				return m.taskMethod(value, x.Left.Text, x.Pos)
			case string:
				return stdlib.StringMethod(value, x.Left.Text, args)
			case *rt.Array:
				return stdlib.ArrayMethod(value, x.Left.Text, args)
			default:
				return stdlib.ValueMethod(receiver, x.Left.Text, args)
			}
		}
		if st := m.structs[name]; st != nil {
			v := &rt.Struct{Name: name, Fields: map[string]any{}}
			for i, f := range st.Fields {
				v.Order = append(v.Order, f.Name)
				v.Fields[f.Name] = rt.Clone(args[i])
			}
			return v, nil
		}
		return m.call(m.functions[name], args, x.Pos)
	}
	return nil, fmt.Errorf("unsupported expression %s", x.Kind)
}
func (m *machine) index(x *ast.Expr, env *environment) (*rt.Array, int, error) {
	v, e := m.expression(x.Left, env)
	if e != nil {
		return nil, 0, e
	}
	index, e := m.expression(x.Right, env)
	if e != nil {
		return nil, 0, e
	}
	a := v.(*rt.Array)
	i := index.(int64)
	if i < 0 || i >= int64(len(a.Items)) {
		return nil, 0, fmt.Errorf("array index %d out of bounds for length %d", i, len(a.Items))
	}
	return a, int(i), nil
}
