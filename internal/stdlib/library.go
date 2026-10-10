// Package stdlib provides typed built-ins without coupling them to the AST interpreter.
package stdlib

import (
	"context"
	"craft/internal/ast"
	rt "craft/internal/runtime"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

type Signature struct {
	Params   []ast.Type
	Result   ast.Type
	Variadic bool
}

var signatures = map[string]Signature{
	"print":                   {Result: ast.Void, Variadic: true},
	"Exception":               {Params: []ast.Type{ast.String}, Result: ast.Exception},
	"std.math.abs":            {Params: []ast.Type{ast.Float}, Result: ast.Float},
	"std.math.sqrt":           {Params: []ast.Type{ast.Float}, Result: ast.Float},
	"std.math.min":            {Params: []ast.Type{ast.Float, ast.Float}, Result: ast.Float},
	"std.math.max":            {Params: []ast.Type{ast.Float, ast.Float}, Result: ast.Float},
	"std.fs.readText":         {Params: []ast.Type{ast.String}, Result: ast.String},
	"std.fs.writeText":        {Params: []ast.Type{ast.String, ast.String}, Result: ast.Void},
	"std.fs.writeTextAtomic":  {Params: []ast.Type{ast.String, ast.String}, Result: ast.Void},
	"std.fs.readBytes":        {Params: []ast.Type{ast.String}, Result: ast.Bytes},
	"std.fs.writeBytesAtomic": {Params: []ast.Type{ast.String, ast.Bytes}, Result: ast.Void},
	"std.fs.exists":           {Params: []ast.Type{ast.String}, Result: ast.Bool},
	"std.time.now":            {Result: ast.String},
	"std.time.sleep":          {Params: []ast.Type{ast.Int}, Result: ast.Void},
	"std.json.isValid":        {Params: []ast.Type{ast.String}, Result: ast.Bool},
	"std.json.quote":          {Params: []ast.Type{ast.String}, Result: ast.String},
	"std.env.get":             {Params: []ast.Type{ast.String}, Result: ast.String},
}

func init() {
	signatures["std.ui.openWindow"] = Signature{Params: []ast.Type{ast.String, ast.Int, ast.Int}, Result: ast.UiWindow}
	signatures["std.ui.setTitle"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String}, Result: ast.Void}
	signatures["std.ui.setWindowIcon"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String}, Result: ast.Void}
	signatures["std.ui.maximizeWindow"] = Signature{Params: []ast.Type{ast.UiWindow}, Result: ast.Void}
	signatures["std.ui.inputStyle"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.setBackground"] = Signature{Params: []ast.Type{ast.UiWindow, ast.Int, ast.Int, ast.Int}, Result: ast.Void}
	signatures["std.ui.closeWindow"] = Signature{Params: []ast.Type{ast.UiWindow}, Result: ast.Void}
	signatures["std.ui.nextEvent"] = Signature{Params: []ast.Type{ast.Int}, Result: ast.Type("UiEvent").Optional()}
	signatures["std.ui.row"] = Signature{Params: []ast.Type{ast.String, ast.UiNode.Array()}, Result: ast.UiNode}
	signatures["std.ui.column"] = Signature{Params: []ast.Type{ast.String, ast.UiNode.Array()}, Result: ast.UiNode}
	signatures["std.ui.panel"] = Signature{Params: []ast.Type{ast.String, ast.UiNode.Array()}, Result: ast.UiNode}
	signatures["std.ui.grid"] = Signature{Params: []ast.Type{ast.String, ast.UiNode.Array(), ast.Int, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.size"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.maxSize"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.align"] = Signature{Params: []ast.Type{ast.UiNode, ast.String, ast.String}, Result: ast.UiNode}
	signatures["std.ui.spacing"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.splitView"] = Signature{Params: []ast.Type{ast.String, ast.UiNode, ast.UiNode, ast.Int, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.scrollView"] = Signature{Params: []ast.Type{ast.String, ast.UiNode}, Result: ast.UiNode}
	signatures["std.ui.virtualList"] = Signature{Params: []ast.Type{ast.String, ast.Int, ast.Int, ast.Int, ast.UiNode.Array()}, Result: ast.UiNode}
	signatures["std.ui.virtualGrid"] = Signature{Params: []ast.Type{ast.String, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.UiNode.Array()}, Result: ast.UiNode}
	signatures["std.ui.gridSelection"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.gridRangeSelection"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.gridRangeStyle"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.clipboardReadText"] = Signature{Params: []ast.Type{ast.UiWindow}, Result: ast.String}
	signatures["std.ui.clipboardWriteText"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String}, Result: ast.Void}
	signatures["std.ui.gridEditor"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int, ast.UiNode}, Result: ast.UiNode}
	signatures["std.ui.text"] = Signature{Params: []ast.Type{ast.String, ast.String}, Result: ast.UiNode}
	signatures["std.ui.button"] = Signature{Params: []ast.Type{ast.String, ast.String}, Result: ast.UiNode}
	signatures["std.ui.textStyle"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.textWrap"] = Signature{Params: []ast.Type{ast.UiNode}, Result: ast.UiNode}
	signatures["std.ui.focusPolicy"] = Signature{Params: []ast.Type{ast.UiNode, ast.Bool, ast.Bool}, Result: ast.UiNode}
	signatures["std.ui.accessible"] = Signature{Params: []ast.Type{ast.UiNode, ast.String, ast.String}, Result: ast.UiNode}
	signatures["std.ui.focus"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String}, Result: ast.Void}
	signatures["std.ui.pushFocusScope"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String}, Result: ast.Void}
	signatures["std.ui.popFocusScope"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String}, Result: ast.Void}
	signatures["std.ui.bindScopedShortcut"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String, ast.String, ast.Int, ast.Int}, Result: ast.Void}
	signatures["std.ui.unbindScopedShortcut"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String, ast.String}, Result: ast.Void}
	signatures["std.ui.bindShortcut"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String, ast.Int, ast.Int}, Result: ast.Void}
	signatures["std.ui.unbindShortcut"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String}, Result: ast.Void}
	signatures["std.ui.surface"] = Signature{Params: []ast.Type{ast.UiNode, ast.Int, ast.Int, ast.Int}, Result: ast.UiNode}
	signatures["std.ui.textInput"] = Signature{Params: []ast.Type{ast.String, ast.String, ast.String}, Result: ast.UiNode}
	signatures["std.ui.setText"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String, ast.String}, Result: ast.Void}
	signatures["std.ui.mount"] = Signature{Params: []ast.Type{ast.UiWindow, ast.UiNode}, Result: ast.Void}
	signatures["std.ui.update"] = Signature{Params: []ast.Type{ast.UiWindow, ast.UiNode}, Result: ast.Void}
	signatures["std.ui.renderer"] = Signature{Params: []ast.Type{ast.UiWindow}, Result: ast.String}
	signatures["std.ui.viewportWidth"] = Signature{Params: []ast.Type{ast.UiWindow}, Result: ast.Int}
	signatures["std.ui.viewportHeight"] = Signature{Params: []ast.Type{ast.UiWindow}, Result: ast.Int}
	signatures["std.ui.measureTextWidth"] = Signature{Params: []ast.Type{ast.UiWindow, ast.String, ast.Int, ast.Int}, Result: ast.Int}
	signatures["std.ui.color"] = Signature{Params: []ast.Type{ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.Int}
	signatures["std.ui.fillRect"] = Signature{Params: []ast.Type{ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.strokeRect"] = Signature{Params: []ast.Type{ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.drawLine"] = Signature{Params: []ast.Type{ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.pathMoveTo"] = Signature{Params: []ast.Type{ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.pathLineTo"] = Signature{Params: []ast.Type{ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.pathArcTo"] = Signature{Params: []ast.Type{ast.Int, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.pathQuadraticTo"] = Signature{Params: []ast.Type{ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.pathCubicTo"] = Signature{Params: []ast.Type{ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.pathClose"] = Signature{Params: []ast.Type{}, Result: ast.UiDraw}
	signatures["std.ui.fillPath"] = Signature{Params: []ast.Type{ast.UiDraw.Array(), ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.strokePath"] = Signature{Params: []ast.Type{ast.UiDraw.Array(), ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.drawText"] = Signature{Params: []ast.Type{ast.String, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.drawTextStyled"] = Signature{Params: []ast.Type{ast.String, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.String}, Result: ast.UiDraw}
	signatures["std.ui.hitRegion"] = Signature{Params: []ast.Type{ast.String, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.targetPolicy"] = Signature{Params: []ast.Type{ast.UiDraw, ast.Bool, ast.Bool, ast.String}, Result: ast.UiDraw}
	signatures["std.ui.drawImage"] = Signature{Params: []ast.Type{ast.String, ast.Int, ast.Int, ast.Int, ast.Int}, Result: ast.UiDraw}
	signatures["std.ui.canvas"] = Signature{Params: []ast.Type{ast.String, ast.Int, ast.Int, ast.UiDraw.Array()}, Result: ast.UiNode}
}

func Lookup(name string) (Signature, bool) { s, ok := signatures[name]; return s, ok }
func Method(receiver ast.Type, name string) (Signature, bool) {
	if s, ok := rev4Method(receiver, name); ok {
		return s, true
	}
	if receiver == ast.Task || receiver == ast.Timer {
		switch name {
		case "join", "cancel":
			return Signature{Result: ast.Void}, true
		case "status":
			return Signature{Result: ast.String}, true
		}
	}
	if receiver == ast.String {
		switch name {
		case "upper", "lower", "trim":
			return Signature{Result: ast.String}, true
		case "contains", "startsWith", "endsWith":
			return Signature{Params: []ast.Type{ast.String}, Result: ast.Bool}, true
		case "split":
			return Signature{Params: []ast.Type{ast.String}, Result: ast.String.Array()}, true
		}
	}
	if receiver.IsArray() {
		switch name {
		case "append":
			return Signature{Params: []ast.Type{receiver.Element()}, Result: ast.Void}, true
		case "remove":
			return Signature{Params: []ast.Type{ast.Int}, Result: receiver.Element()}, true
		}
	}
	return methodRev2(receiver, name)
}
func Invoke(ctx context.Context, out io.Writer, name string, args []any) (any, error) {
	if strings.HasPrefix(name, "std.encoding.") || strings.HasPrefix(name, "std.crypto.") || strings.HasPrefix(name, "std.url.") || strings.HasPrefix(name, "std.regex.") {
		return rev43DataInvoke(ctx, name, args)
	}
	if strings.HasPrefix(name, "std.os.") {
		return invokeOS(ctx, name, args)
	}
	if name == "std.fs.writeTextAtomic" || name == "std.fs.writeBytesAtomic" {
		var data string
		if name == "std.fs.writeTextAtomic" {
			data = args[1].(string)
			if !utf8.ValidString(data) {
				return nil, toolingError("FsEncodingError", "atomic text write requires UTF-8")
			}
		} else {
			data = args[1].(Bytes).data
		}
		return nil, writeAtomicContext(ctx, args[0].(string), []byte(data))
	}
	if strings.HasPrefix(name, "std.config.") {
		return configInvoke(name, args)
	}
	if strings.HasPrefix(name, "std.fs.") {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if strings.HasPrefix(name, "std.db.") {
		return invokeDB(ctx, name, args)
	}
	if strings.HasPrefix(name, "std.bytes.") || strings.HasPrefix(name, "std.context.") || strings.HasPrefix(name, "std.http.") {
		return invokeRev4(ctx, name, args)
	}
	switch name {
	case "print":
		values := append([]any(nil), args...)
		for i, v := range values {
			if v == nil {
				values[i] = "null"
			}
		}
		_, e := fmt.Fprintln(out, values...)
		return nil, e
	case "Exception":
		return &rt.Exception{Message: args[0].(string), Kind: "Exception"}, nil
	case "std.math.abs":
		return math.Abs(args[0].(float64)), nil
	case "std.math.sqrt":
		v := args[0].(float64)
		if v < 0 {
			return nil, fmt.Errorf("sqrt requires a nonnegative value")
		}
		return math.Sqrt(v), nil
	case "std.math.min":
		return math.Min(args[0].(float64), args[1].(float64)), nil
	case "std.math.max":
		return math.Max(args[0].(float64), args[1].(float64)), nil
	case "std.fs.readText":
		path := args[0].(string)
		b, e := readFile(path)
		if e != nil {
			return nil, e
		}
		if !utf8.Valid(b) {
			return nil, toolingError("FsEncodingError", fmt.Sprintf("filesystem read %q failed: requires UTF-8 text", path))
		}
		return string(b), nil
	case "std.fs.writeText":
		path := args[0].(string)
		if e := os.WriteFile(path, []byte(args[1].(string)), 0644); e != nil {
			return nil, fsFailure("write", path, e)
		}
		return nil, nil
	case "std.fs.exists":
		_, e := os.Stat(args[0].(string))
		if os.IsNotExist(e) {
			return false, nil
		}
		if e != nil {
			return nil, fsFailure("inspect", args[0].(string), e)
		}
		return true, nil
	case "std.time.now":
		return time.Now().UTC().Format(time.RFC3339Nano), nil
	case "std.time.sleep":
		d, e := SleepDuration(args[0])
		if e != nil {
			return nil, e
		}
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, nil
		}
	case "std.json.isValid":
		return json.Valid([]byte(args[0].(string))), nil
	case "std.json.quote":
		b, e := json.Marshal(args[0].(string))
		return string(b), e
	case "std.env.get":
		return os.Getenv(args[0].(string)), nil
	}
	return invokeRev2(out, name, args)
}
func StringMethod(s, name string, args []any) (any, error) {
	switch name {
	case "replace":
		return strings.ReplaceAll(s, args[0].(string), args[1].(string)), nil
	case "substring":
		r := []rune(s)
		a, b, e := sliceBounds(args, len(r))
		if e != nil {
			return nil, e
		}
		return string(r[a:b]), nil
	case "upper":
		return strings.ToUpper(s), nil
	case "lower":
		return strings.ToLower(s), nil
	case "trim":
		return strings.TrimSpace(s), nil
	case "contains":
		return strings.Contains(s, args[0].(string)), nil
	case "startsWith":
		return strings.HasPrefix(s, args[0].(string)), nil
	case "endsWith":
		return strings.HasSuffix(s, args[0].(string)), nil
	case "split":
		parts := strings.Split(s, args[0].(string))
		a := &rt.Array{}
		for _, p := range parts {
			a.Items = append(a.Items, p)
		}
		return a, nil
	}
	return nil, fmt.Errorf("unknown String method %s", name)
}

func ArrayMethod(value *rt.Array, name string, args []any) (any, error) {
	switch name {
	case "append":
		value.Items = append(value.Items, rt.Clone(args[0]))
		return nil, nil
	case "remove":
		i := args[0].(int64)
		if i < 0 || i >= int64(len(value.Items)) {
			return nil, fmt.Errorf("array index %d out of bounds for length %d", i, len(value.Items))
		}
		removed := value.Items[i]
		copy(value.Items[i:], value.Items[i+1:])
		value.Items[len(value.Items)-1] = nil
		value.Items = value.Items[:len(value.Items)-1]
		return rt.Clone(removed), nil
	}
	return arrayRev2(value, name, args)
}

// QualifiedName recognizes built-in namespaces, not user modules.
func QualifiedName(x *ast.Expr) string {
	parts := []string{}
	for x != nil && x.Kind == "member" {
		if len(parts) >= 512 {
			return ""
		}
		parts = append(parts, x.Text)
		x = x.Left
	}
	if x == nil || x.Kind != "identifier" {
		return ""
	}
	parts = append(parts, x.Text)
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".")
}
