//go:build !windows

package stdlib

import (
	"context"
	"craft/internal/ast"
)

func init() {
	signatures["std.os.version"] = Signature{Result: ast.String}
	signatures["std.os.architecture"] = Signature{Result: ast.String}
	signatures["std.os.processId"] = Signature{Result: ast.Int}
	signatures["std.os.open"] = Signature{Params: []ast.Type{ast.String}, Result: ast.Void}
}

func invokeOS(ctx context.Context, name string, a []any) (any, error) {
	return nil, toolingError("OsUnsupportedError", "std.os requires Windows")
}
