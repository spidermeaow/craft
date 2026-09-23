package interpreter

import (
	"context"
	"craft/internal/diagnostics"
	rt "craft/internal/runtime"
	"craft/internal/stdlib"
	"fmt"
)

func (m *machine) serveHTTP(args []any, pos diagnostics.Position) (any, error) {
	if m.cleaning || m.servingRequest {
		return nil, fmt.Errorf("cannot start HTTP server from request or defer cleanup")
	}
	f := args[1].(rt.Function).Declaration
	if !snapshotSafe(args[2]) {
		return nil, fmt.Errorf("HTTP state must not contain Task, Timer, Context or DbTransaction handles")
	}
	state := rt.Clone(args[2]) // Frozen private snapshot, only read to clone per request.
	err := stdlib.ServeHTTPTransport(m.ctx, args[0].(*rt.Struct), func(ctx context.Context, request *rt.Struct) (*rt.Struct, error) {
		child := &machine{ctx: ctx, out: m.out, tasks: m.tasks, functions: m.functions, structs: m.structs, library: m.library.Fork(), servingRequest: true}
		v, e := child.call(f, []any{request, rt.Clone(state)}, pos)
		if e != nil {
			return nil, e
		}
		return v.(*rt.Struct), nil
	}, m.out)
	return nil, err
}

func snapshotSafe(v any) bool {
	switch x := v.(type) {
	case *task, *stdlib.Context, *stdlib.DBTransaction:
		return false
	case rt.Optional:
		return !x.Valid || snapshotSafe(x.Value)
	case *rt.Struct:
		for _, v := range x.Fields {
			if !snapshotSafe(v) {
				return false
			}
		}
	case *rt.Map:
		for _, v := range x.Values {
			if !snapshotSafe(v) {
				return false
			}
		}
	case *rt.Array:
		for _, v := range x.Items {
			if !snapshotSafe(v) {
				return false
			}
		}
	}
	return true
}
