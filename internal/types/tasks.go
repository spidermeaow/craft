package types

import (
	"craft/internal/ast"
	"craft/internal/diagnostics"
	"craft/internal/resolver"
	"craft/internal/stdlib"
)

// Task/timer callbacks share the general function-value type system in Rev.4.
func (c *checker) taskCall(x *ast.Expr, env *resolver.Scope, name string) (ast.Type, error) {
	offset, result := 0, ast.Task
	if name != "std.task.spawn" {
		offset, result = 1, ast.Timer
	}
	for _, n := range x.ArgNames {
		if n != "" {
			return "", diagnostics.New(x.Pos, "type error: task/timer calls use positional arguments")
		}
	}
	if len(x.Args) <= offset {
		return "", diagnostics.New(x.Pos, "type error: %s requires a named function", name)
	}
	if offset == 1 {
		t, e := c.expression(x.Args[0], env)
		if e != nil {
			return "", e
		}
		if t != ast.Duration {
			return "", diagnostics.New(x.Pos, "type error: timer interval must be Duration")
		}
	}
	ref := x.Args[offset]
	t, e := c.expression(ref, env)
	if e != nil {
		return "", e
	}
	ret, params, ok := t.Function()
	if !ok || ret != ast.Void {
		return "", diagnostics.New(ref.Pos, "type error: task entry must be a function returning Void")
	}
	args := x.Args[offset+1:]
	if len(args) != len(params) {
		return "", diagnostics.New(x.Pos, "type error: callback requires %d arguments, got %d", len(params), len(args))
	}
	for i, arg := range args {
		t, e := c.expected(arg, env, params[i])
		if e != nil {
			return "", e
		}
		if t != params[i] {
			return "", diagnostics.New(arg.Pos, "type error: callback argument %d must be %s, got %s", i+1, params[i], t)
		}
	}
	return result, nil
}

func (c *checker) callValue(x *ast.Expr, env *resolver.Scope, t ast.Type) (ast.Type, error) {
	r, p, ok := t.Function()
	if !ok {
		return "", diagnostics.New(x.Pos, "type error: value of type %s is not callable", t)
	}
	x.Type = t // Interpreter dispatch marker. Source syntax is unchanged.
	return c.arguments(x, env, stdlib.Signature{Params: p, Result: r})
}
