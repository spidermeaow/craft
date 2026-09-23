package interpreter

import (
	"craft/internal/diagnostics"
	rt "craft/internal/runtime"
	"math"
)

func calculate(op string, l, r any, p diagnostics.Position) (any, error) {
	fail := func(message string) (any, error) { return nil, diagnostics.New(p, "runtime error: %s", message) }
	if opt, ok := l.(rt.Optional); ok && r == nil {
		l = opt.Valid
		r = false
	}
	if opt, ok := r.(rt.Optional); ok && l == nil {
		r = opt.Valid
		l = false
	}
	if op == "==" {
		return l == r, nil
	}
	if op == "!=" {
		return l != r, nil
	}
	switch a := l.(type) {
	case bool:
		b := r.(bool)
		if op == "&&" {
			return a && b, nil
		}
		if op == "||" {
			return a || b, nil
		}
	case string:
		b := r.(string)
		switch op {
		case "+":
			return a + b, nil
		case "<":
			return a < b, nil
		case ">":
			return a > b, nil
		case "<=":
			return a <= b, nil
		case ">=":
			return a >= b, nil
		}
	case int64:
		b := r.(int64)
		switch op {
		case "+":
			v := a + b
			if (b > 0 && v < a) || (b < 0 && v > a) {
				return fail("Int overflow")
			}
			return v, nil
		case "-":
			v := a - b
			if (b < 0 && v < a) || (b > 0 && v > a) {
				return fail("Int overflow")
			}
			return v, nil
		case "*":
			if (a == math.MinInt64 && b == -1) || (b == math.MinInt64 && a == -1) {
				return fail("Int overflow")
			}
			v := a * b
			if b != 0 && v/b != a {
				return fail("Int overflow")
			}
			return v, nil
		case "/", "%":
			if b == 0 {
				return fail("division by zero; use a nonzero divisor")
			}
			if a == math.MinInt64 && b == -1 {
				if op == "%" {
					return int64(0), nil
				}
				return fail("Int overflow")
			}
			if op == "/" {
				return a / b, nil
			}
			return a % b, nil
		case "<":
			return a < b, nil
		case ">":
			return a > b, nil
		case "<=":
			return a <= b, nil
		case ">=":
			return a >= b, nil
		}
	case float64:
		b := r.(float64)
		var v float64
		switch op {
		case "+":
			v = a + b
		case "-":
			v = a - b
		case "*":
			v = a * b
		case "/":
			if b == 0 {
				return fail("division by zero; use a nonzero divisor")
			}
			v = a / b
		case "<":
			return a < b, nil
		case ">":
			return a > b, nil
		case "<=":
			return a <= b, nil
		case ">=":
			return a >= b, nil
		default:
			return fail("unsupported operator " + op)
		}
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return fail("Float overflow")
		}
		return v, nil
	}
	return fail("unsupported operator " + op)
}
