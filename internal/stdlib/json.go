package stdlib

import (
	"craft/internal/ast"
	rt "craft/internal/runtime"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// JSON contains immutable private data. Accessors return independent Craft values.
type JSON struct{ value any }

func (j JSON) String() string { b, _ := json.Marshal(j.value); return string(b) }
func registerJSON(add func(string, ast.Type, ...ast.Type)) {
	add("std.json.parse", ast.JsonValue, ast.String)
	add("std.json.stringify", ast.String, ast.JsonValue)
	add("std.json.nullValue", ast.JsonValue)
	for n, t := range map[string]ast.Type{"fromInt": ast.Int, "fromFloat": ast.Float, "fromBool": ast.Bool, "fromString": ast.String, "fromArray": ast.JsonValue.Array(), "fromObject": ast.JsonValue.Map()} {
		add("std.json."+n, ast.JsonValue, t)
	}
}
func jsonMethodSignature(name string) (Signature, bool) {
	s := Signature{}
	switch name {
	case "kind":
		s.Result = ast.String
	case "asInt":
		s.Result = ast.Int
	case "asFloat":
		s.Result = ast.Float
	case "asBool":
		s.Result = ast.Bool
	case "asString":
		s.Result = ast.String
	case "asArray":
		s.Result = ast.JsonValue.Array()
	case "asObject":
		s.Result = ast.JsonValue.Map()
	default:
		return s, false
	}
	return s, true
}

// Token parsing detects duplicate keys instead of silently losing data.
func parseJSON(d *json.Decoder, depth int) (any, error) {
	if depth > 128 {
		return nil, fmt.Errorf("JSON nesting exceeds 128")
	}
	t, e := d.Token()
	if e != nil {
		return nil, e
	}
	switch v := t.(type) {
	case json.Delim:
		if v == '[' {
			a := []any{}
			for d.More() {
				x, e := parseJSON(d, depth+1)
				if e != nil {
					return nil, e
				}
				a = append(a, x)
			}
			_, e = d.Token()
			return a, e
		}
		if v == '{' {
			m := map[string]any{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return nil, e
				}
				key, ok := k.(string)
				if !ok {
					return nil, fmt.Errorf("JSON object key must be String")
				}
				if _, ok = m[key]; ok {
					return nil, fmt.Errorf("duplicate JSON key %q", key)
				}
				x, e := parseJSON(d, depth+1)
				if e != nil {
					return nil, e
				}
				m[key] = x
			}
			_, e = d.Token()
			return m, e
		}
		return nil, fmt.Errorf("unexpected JSON delimiter")
	case json.Number:
		// Keep the exact decimal spelling; accessors apply their own numeric bounds.
		return v, nil
	default:
		return t, nil
	}
}
func jsonInvoke(name string, args []any) (any, error) {
	switch name {
	case "std.json.parse":
		text := args[0].(string)
		if len(text) > 16*1024*1024 || !utf8.ValidString(text) {
			return nil, fmt.Errorf("JSON requires UTF-8 text up to 16 MiB")
		}
		d := json.NewDecoder(strings.NewReader(text))
		d.UseNumber()
		v, e := parseJSON(d, 0)
		if e != nil {
			return nil, e
		}
		if _, e = d.Token(); e != io.EOF {
			return nil, fmt.Errorf("JSON must contain exactly one value")
		}
		return JSON{v}, nil
	case "std.json.stringify":
		b, e := json.Marshal(args[0].(JSON).value)
		if len(b) > 16*1024*1024 {
			return nil, fmt.Errorf("JSON output exceeds 16 MiB")
		}
		return string(b), e
	case "std.json.nullValue":
		return JSON{nil}, nil
	case "std.json.fromInt":
		return JSON{json.Number(strconv.FormatInt(args[0].(int64), 10))}, nil
	case "std.json.fromFloat":
		v := args[0].(float64)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("JSON requires finite numbers")
		}
		return JSON{json.Number(strconv.FormatFloat(v, 'g', -1, 64))}, nil
	case "std.json.fromString", "std.json.fromBool":
		return JSON{args[0]}, nil
	case "std.json.fromArray":
		a := []any{}
		for _, v := range args[0].(*rt.Array).Items {
			a = append(a, v.(JSON).value)
		}
		j := JSON{a}
		return boundedJSON(j)
	case "std.json.fromObject":
		m := map[string]any{}
		for k, v := range args[0].(*rt.Map).Values {
			m[k] = v.(JSON).value
		}
		j := JSON{m}
		return boundedJSON(j)
	}
	return nil, fmt.Errorf("unknown JSON function %s", name)
}
func boundedJSON(j JSON) (any, error) {
	// Constructors must preserve the same depth/size limits as the parser.
	var walk func(any, int) bool
	walk = func(v any, n int) bool {
		if n > 128 {
			return false
		}
		switch x := v.(type) {
		case []any:
			for _, a := range x {
				if !walk(a, n+1) {
					return false
				}
			}
		case map[string]any:
			for _, a := range x {
				if !walk(a, n+1) {
					return false
				}
			}
		}
		return true
	}
	if !walk(j.value, 0) {
		return nil, fmt.Errorf("JSON nesting exceeds 128")
	}
	b, e := json.Marshal(j.value)
	if e != nil {
		return nil, e
	}
	if len(b) > 16*1024*1024 {
		return nil, fmt.Errorf("JSON output exceeds 16 MiB")
	}
	return j, nil
}
func jsonMethod(j JSON, name string, args []any) (any, error) {
	kind := "null"
	switch j.value.(type) {
	case bool:
		kind = "bool"
	case string:
		kind = "string"
	case json.Number:
		kind = "number"
	case []any:
		kind = "array"
	case map[string]any:
		kind = "object"
	}
	if name == "kind" {
		return kind, nil
	}
	switch name {
	case "asInt":
		if v, ok := j.value.(json.Number); ok {
			return v.Int64()
		}
	case "asFloat":
		if v, ok := j.value.(json.Number); ok {
			n, e := v.Float64()
			if e != nil {
				return nil, e
			}
			return finite(n)
		}
	case "asBool":
		if v, ok := j.value.(bool); ok {
			return v, nil
		}
	case "asString":
		if v, ok := j.value.(string); ok {
			return v, nil
		}
	case "asArray":
		if v, ok := j.value.([]any); ok {
			a := &rt.Array{}
			for _, x := range v {
				a.Items = append(a.Items, JSON{x})
			}
			return a, nil
		}
	case "asObject":
		if v, ok := j.value.(map[string]any); ok {
			m := rt.NewMap()
			for _, k := range sortedKeys(v) {
				m.Set(k, JSON{v[k]})
			}
			return m, nil
		}
	}
	return nil, fmt.Errorf("JSON %s cannot use %s", kind, name)
}
