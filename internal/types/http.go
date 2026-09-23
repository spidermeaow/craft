package types

import (
	"craft/internal/ast"
	"craft/internal/diagnostics"
	"craft/internal/resolver"
)

func (c *checker) httpCall(x *ast.Expr, env *resolver.Scope) (ast.Type, error) {
	if len(x.Args) != 3 {
		return "", diagnostics.New(x.Pos, "type error: serve expects config, dispatch and state")
	}
	for _, n := range x.ArgNames {
		if n != "" {
			return "", diagnostics.New(x.Pos, "type error: serve uses positional arguments")
		}
	}
	cfg, e := c.expression(x.Args[0], env)
	if e != nil {
		return "", e
	}
	if cfg != "HttpConfig" {
		return "", diagnostics.New(x.Pos, "type error: serve requires HttpConfig")
	}
	fn, e := c.expression(x.Args[1], env)
	if e != nil {
		return "", e
	}
	state, e := c.expression(x.Args[2], env)
	if e != nil {
		return "", e
	}
	r, p, ok := fn.Function()
	if !ok || r != "HttpResponse" || len(p) != 2 || p[0] != "HttpRequest" || p[1] != state || !c.valid(state, false) {
		return "", diagnostics.New(x.Pos, "type error: dispatch must have type Fn<HttpResponse,HttpRequest,State> matching state")
	}
	return ast.Void, nil
}
