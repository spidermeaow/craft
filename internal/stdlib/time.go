package stdlib

import (
	"craft/internal/ast"
	"fmt"
	"math"
	"sort"
	"time"
)

func sortedKeys(m map[string]any) []string {
	k := []string{}
	for s := range m {
		k = append(k, s)
	}
	sort.Strings(k)
	return k
}

type DateTime struct{ value time.Time }
type Duration struct{ value time.Duration }

// GoDuration exposes an immutable duration to the task runtime.
func (d Duration) GoDuration() time.Duration { return d.value }

// Keep the Rev.1/2 Int-millisecond overload and its original one-day limit.
func SleepDuration(v any) (time.Duration, error) {
	if d, ok := v.(Duration); ok {
		if d.value < 0 {
			return 0, fmt.Errorf("sleep duration must be nonnegative")
		}
		return d.value, nil
	}
	n, ok := v.(int64)
	if !ok || n < 0 || n > 86400000 {
		return 0, fmt.Errorf("sleep expects Duration or milliseconds between 0 and 86400000")
	}
	return time.Duration(n) * time.Millisecond, nil
}

func (d DateTime) String() string { return d.value.Format(time.RFC3339Nano) }
func (d Duration) String() string { return d.value.String() }
func registerTime(add func(string, ast.Type, ...ast.Type)) {
	add("std.time.parse", ast.DateTime, ast.String)
	add("std.time.fromUnixSeconds", ast.DateTime, ast.Int)
	add("std.time.durationMilliseconds", ast.Duration, ast.Int)
}
func timeMethodSignature(t ast.Type, n string) (Signature, bool) {
	s := func(r ast.Type, p ...ast.Type) (Signature, bool) { return Signature{Params: p, Result: r}, true }
	if t == ast.Duration {
		if n == "milliseconds" {
			return s(ast.Int)
		}
		return Signature{}, false
	}
	switch n {
	case "format":
		return s(ast.String)
	case "unixSeconds", "unixMilliseconds":
		return s(ast.Int)
	case "withOffset":
		return s(ast.DateTime, ast.Int)
	case "add", "subtract":
		return s(ast.DateTime, ast.Duration)
	case "difference":
		return s(ast.Duration, ast.DateTime)
	}
	return Signature{}, false
}
func validDate(t time.Time) (any, error) {
	if t.Year() < 1 || t.Year() > 9999 {
		return nil, fmt.Errorf("DateTime year must be 1..9999")
	}
	return DateTime{t}, nil
}
func timeInvoke(n string, a []any) (any, error) {
	switch n {
	case "std.time.parse":
		t, e := time.Parse(time.RFC3339Nano, a[0].(string))
		if e != nil {
			return nil, e
		}
		_, off := t.Zone()
		if off%60 != 0 || off <= -24*3600 || off >= 24*3600 {
			return nil, fmt.Errorf("invalid UTC offset")
		}
		return validDate(t)
	case "std.time.fromUnixSeconds":
		v := a[0].(int64)
		if v < -62135596800 || v > 253402300799 {
			return nil, fmt.Errorf("Unix seconds outside years 1..9999")
		}
		return validDate(time.Unix(v, 0).UTC())
	case "std.time.durationMilliseconds":
		v := a[0].(int64)
		if v > math.MaxInt64/int64(time.Millisecond) || v < math.MinInt64/int64(time.Millisecond) {
			return nil, fmt.Errorf("Duration overflow")
		}
		return Duration{time.Duration(v) * time.Millisecond}, nil
	}
	return nil, fmt.Errorf("unknown time function %s", n)
}
func dateMethod(d DateTime, n string, a []any) (any, error) {
	switch n {
	case "format":
		return d.String(), nil
	case "unixSeconds":
		return d.value.Unix(), nil
	case "unixMilliseconds":
		return d.value.UnixMilli(), nil
	case "withOffset":
		minutes := a[0].(int64)
		if minutes < -1439 || minutes > 1439 {
			return nil, fmt.Errorf("offset minutes must be -1439..1439")
		}
		return validDate(d.value.In(time.FixedZone("", int(minutes)*60)))
	case "add", "subtract":
		v := a[0].(Duration).value
		if n == "subtract" {
			if v == time.Duration(math.MinInt64) {
				return nil, fmt.Errorf("Duration overflow")
			}
			v = -v
		}
		return validDate(d.value.Add(v))
	case "difference":
		other := a[0].(DateTime).value
		v := d.value.Sub(other)
		if !other.Add(v).Equal(d.value) {
			return nil, fmt.Errorf("Duration overflow")
		}
		return Duration{v}, nil
	}
	return nil, fmt.Errorf("unknown DateTime method %s", n)
}
func durationMethod(d Duration, n string, a []any) (any, error) {
	if n == "milliseconds" {
		return d.value.Milliseconds(), nil
	}
	return nil, fmt.Errorf("unknown Duration method %s", n)
}
