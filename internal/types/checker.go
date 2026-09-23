package types

import (
	"craft/internal/ast"
	"craft/internal/diagnostics"
	"craft/internal/resolver"
	"craft/internal/stdlib"
	"strconv"
)

type exits uint8

const (
	next exits = 1 << iota
	returned
	thrown
	broken
	continued
)

type checker struct {
	structs          map[string]*ast.Struct
	functions        map[string]*ast.Function
	result           ast.Type
	loops, depth     int
	deferred, inTest bool
}

func (c *checker) valid(t ast.Type, void bool) bool {
	if result, params, ok := t.Function(); ok {
		if !c.valid(result, true) {
			return false
		}
		for _, p := range params {
			if !c.valid(p, false) {
				return false
			}
		}
		return true
	}
	if t.IsOptional() {
		return c.valid(t.Base(), false)
	}
	if t.IsArray() {
		return c.valid(t.Element(), false)
	}
	if t.IsMap() {
		return c.valid(t.MapElement(), false)
	}
	return stdlib.IsDBType(t) || t.Primitive() || t == ast.Exception || t == ast.JsonValue || t == ast.DateTime || t == ast.Duration || t == ast.Random || t == ast.Task || t == ast.Timer || t == ast.Bytes || t == ast.Context || c.structs[string(t)] != nil || (void && t == ast.Void)
}
func Check(p *ast.Program) error { return CheckMode(p, true) }
func CheckMode(p *ast.Program, requireMain bool) error {
	c := &checker{functions: map[string]*ast.Function{}, structs: map[string]*ast.Struct{}}
	for _, s := range stdlib.NativeStructs() {
		c.structs[s.Name] = s
	}
	if e := c.registerStructs(p); e != nil {
		return e
	}
	for _, f := range p.Functions {
		if _, ok := stdlib.Lookup(f.Name); ok || f.Name == "std" || c.structs[f.Name] != nil {
			return diagnostics.New(f.Pos, "type error: %s is a reserved built-in name", f.Name)
		}
		if _, ok := c.functions[f.Name]; ok {
			return diagnostics.New(f.Pos, "type error: function %s is declared more than once", f.Name)
		}
		c.functions[f.Name] = f
		if !c.valid(f.Return, true) {
			return diagnostics.New(f.Pos, "type error: unknown return type %s", f.Return)
		}
		for _, a := range f.Params {
			if !c.valid(a.Type, false) {
				return diagnostics.New(a.Pos, "type error: invalid parameter type %s", a.Type)
			}
		}
	}
	main, ok := c.functions["main"]
	if !ok && requireMain {
		return diagnostics.New(diagnostics.Position{File: "<project>", Line: 1, Column: 1}, "entry point error: add exactly one func main()")
	}
	if ok && (len(main.Params) != 0 || main.Return != ast.Void) {
		return diagnostics.New(main.Pos, "entry point error: main must have no parameters and return Void")
	}
	for _, f := range p.Functions {
		env := resolver.New(nil)
		for _, a := range f.Params {
			if e := declare(env, a.Name, a.Type, false, a.Pos); e != nil {
				return e
			}
		}
		c.result = f.Return
		c.loops = 0
		c.deferred = false
		c.inTest = false
		flow, e := c.block(f.Body, env)
		if e != nil {
			return e
		}
		if f.Return != ast.Void && flow&next != 0 {
			return diagnostics.New(f.Pos, "type error: function %s must return %s on every path (missing return value on some execution paths)", f.Name, f.Return)
		}
	}
	seen := map[string]bool{}
	for _, test := range p.Tests {
		if seen[test.Name] {
			return diagnostics.New(test.Pos, "type error: duplicate test name %q", test.Name)
		}
		seen[test.Name] = true
		c.result = ast.Void
		c.loops = 0
		c.deferred = false
		c.inTest = true
		if _, e := c.statement(test.Body, resolver.New(nil)); e != nil {
			return e
		}
	}
	return nil
}
func declare(env *resolver.Scope, name string, t ast.Type, mutable bool, p diagnostics.Position) error {
	if name == "std" {
		return diagnostics.New(p, "type error: std is a reserved library namespace")
	}
	if _, ok := env.Values[name]; ok {
		return diagnostics.New(p, "type error: variable or parameter %s is already declared in this scope (duplicate parameter or variable)", name)
	}
	env.Values[name] = resolver.Symbol{Type: t, Mutable: mutable}
	return nil
}
func (c *checker) block(s *ast.Stmt, env *resolver.Scope) (exits, error) {
	flow := next
	for _, st := range s.Statements {
		out, e := c.statement(st, env)
		if e != nil {
			return 0, e
		}
		if flow&next != 0 {
			flow = (flow &^ next) | out
		}
	}
	return flow, nil
}
func (c *checker) statement(s *ast.Stmt, env *resolver.Scope) (exits, error) {
	switch s.Kind {
	case "block":
		return c.block(s, resolver.New(env))
	case "let", "var":
		if !c.valid(s.Type, false) {
			return 0, diagnostics.New(s.Pos, "type error: invalid variable type %s", s.Type)
		}
		if _, ok := env.Values[s.Name]; ok {
			return 0, diagnostics.New(s.Pos, "type error: variable %s is already declared in this scope", s.Name)
		}
		t, e := c.expected(s.Expr, env, s.Type)
		if e != nil {
			return 0, e
		}
		if t != s.Type {
			return 0, diagnostics.New(s.Pos, "type error: cannot assign %s to %s; use a value of type %s", t, s.Type, s.Type)
		}
		return next, declare(env, s.Name, s.Type, s.Kind == "var", s.Pos)
	case "expression":
		_, e := c.expression(s.Expr, env)
		return next, e
	case "iflet":
		t, e := c.expression(s.Expr, env)
		if e != nil {
			return 0, e
		}
		if !t.IsOptional() {
			return 0, diagnostics.New(s.Pos, "type error: if let requires Optional, got %s", t)
		}
		child := resolver.New(env)
		if e = declare(child, s.Name, t.Base(), false, s.Pos); e != nil {
			return 0, e
		}
		a, e := c.block(s.Body, child)
		if e != nil {
			return 0, e
		}
		b := next
		if s.Else != nil {
			b, e = c.statement(s.Else, env)
		}
		return a | b, e
	case "if", "while":
		t, e := c.expression(s.Expr, env)
		if e != nil {
			return 0, e
		}
		if t != ast.Bool {
			return 0, diagnostics.New(s.Expr.Pos, "type error: condition must be Bool, got %s", t)
		}
		if s.Kind == "while" {
			c.loops++
		}
		a, e := c.statement(s.Body, env)
		if s.Kind == "while" {
			c.loops--
		}
		if e != nil {
			return 0, e
		}
		if s.Kind == "while" {
			return next | (a & (returned | thrown)), nil
		}
		b := next
		if s.Else != nil {
			b, e = c.statement(s.Else, env)
		}
		return a | b, e
	case "for":
		t, e := c.expression(s.Expr, env)
		if e != nil {
			return 0, e
		}
		item := t.Element()
		if s.End != nil {
			end, e := c.expression(s.End, env)
			if e != nil {
				return 0, e
			}
			if t != ast.Int || end != ast.Int {
				return 0, diagnostics.New(s.Pos, "type error: range bounds must be Int")
			}
			item = ast.Int
		} else if !t.IsArray() {
			return 0, diagnostics.New(s.Pos, "type error: for requires an Array or Int range")
		}
		local := resolver.New(env)
		if e = declare(local, s.Name, item, false, s.Pos); e != nil {
			return 0, e
		}
		c.loops++
		flow, e := c.statement(s.Body, local)
		c.loops--
		return next | (flow & (returned | thrown)), e
	case "return":
		if c.deferred || c.inTest {
			return 0, diagnostics.New(s.Pos, "type error: return cannot exit a defer or test block")
		}
		t := ast.Void
		var e error
		if s.Expr != nil {
			t, e = c.expected(s.Expr, env, c.result)
			if e != nil {
				return 0, e
			}
		}
		if t != c.result {
			return 0, diagnostics.New(s.Pos, "type error: return must be %s, got %s", c.result, t)
		}
		return returned, nil
	case "break", "continue":
		if c.loops == 0 {
			return 0, diagnostics.New(s.Pos, "type error: %s is only valid inside a loop within the current function/defer", s.Kind)
		}
		if s.Kind == "break" {
			return broken, nil
		}
		return continued, nil
	case "throw":
		t, e := c.expression(s.Expr, env)
		if e != nil {
			return 0, e
		}
		if t != ast.Exception {
			return 0, diagnostics.New(s.Pos, "type error: throw requires Exception, got %s", t)
		}
		return thrown, nil
	case "assert":
		t, e := c.expression(s.Expr, env)
		if e != nil {
			return 0, e
		}
		if t != ast.Bool {
			return 0, diagnostics.New(s.Pos, "type error: assert requires Bool, got %s", t)
		}
	case "try":
		a, e := c.statement(s.Body, env)
		if e != nil {
			return 0, e
		}
		local := resolver.New(env)
		if e = declare(local, s.Name, ast.Exception, false, s.Pos); e != nil {
			return 0, e
		}
		b, e := c.block(s.Else, local)
		return (a &^ thrown) | b, e
	case "defer":
		if c.deferred {
			return 0, diagnostics.New(s.Pos, "type error: nested defer registration is not supported")
		}
		saved := c.loops
		c.loops = 0
		c.deferred = true
		_, e := c.statement(s.Body, env)
		c.deferred = false
		c.loops = saved
		return next, e
	default:
		return 0, diagnostics.New(s.Pos, "type error: unsupported statement %s", s.Kind)
	}
	return next, nil
}
func (c *checker) expected(x *ast.Expr, env *resolver.Scope, want ast.Type) (ast.Type, error) {
	if want.IsOptional() {
		if x.Kind == "literal" && x.Type == ast.Null {
			x.Kind = "none"
			x.Type = want
			return want, nil
		}
		t, e := c.expected(x, env, want.Base())
		if e != nil {
			return "", e
		}
		if t == want.Base() {
			child := *x
			*x = ast.Expr{Kind: "some", Pos: child.Pos, Type: want, Left: &child}
			return want, nil
		}
		return t, nil
	}
	if (x.Kind == "array" && want.IsArray()) || (x.Kind == "map" && want.IsMap()) {
		x.Type = want
	}
	return c.expression(x, env)
}

func (c *checker) writable(x *ast.Expr, env *resolver.Scope) error {
	root := x
	for root.Kind == "index" || root.Kind == "member" {
		if root.Kind == "member" {
			t, e := c.expression(root.Left, env)
			if e != nil {
				return e
			}
			f, ok := c.field(t, root.Text)
			if !ok || !f.Mutable {
				return diagnostics.New(root.Pos, "type error: cannot assign to immutable field '%s'", root.Text)
			}
		}
		root = root.Left
	}
	if root.Kind != "identifier" {
		return diagnostics.New(x.Pos, "type error: assignment target must be a variable or its array element")
	}
	symbol, ok := env.Find(root.Text)
	if !ok {
		return diagnostics.New(root.Pos, "type error: unknown variable %s", root.Text)
	}
	if !symbol.Mutable {
		return diagnostics.New(root.Pos, "type error: cannot assign to immutable variable '%s'; declare it with var", root.Text)
	}
	return nil
}
func (c *checker) expression(x *ast.Expr, env *resolver.Scope) (ast.Type, error) {
	c.depth++
	defer func() { c.depth-- }()
	if c.depth > 512 {
		return "", diagnostics.New(x.Pos, "type error: expression exceeds 512 levels")
	}
	switch x.Kind {
	case "some", "none":
		return x.Type, nil
	case "literal":
		if x.Type == ast.Int {
			if _, e := strconv.ParseInt(x.Text, 10, 64); e != nil {
				return "", diagnostics.New(x.Pos, "type error: Int literal is outside signed 64-bit range")
			}
		}
		if x.Type == ast.Float {
			if _, e := strconv.ParseFloat(x.Text, 64); e != nil {
				return "", diagnostics.New(x.Pos, "type error: Float literal is outside 64-bit range")
			}
		}
		return x.Type, nil
	case "identifier":
		if s, ok := env.Find(x.Text); ok {
			return s.Type, nil
		}
		if f := c.functions[x.Text]; f != nil {
			return f.Type(), nil
		}
		return "", diagnostics.New(x.Pos, "type error: unknown variable %s; declare it with let or var", x.Text)
	case "map":
		element := ast.Type("")
		if x.Type.IsMap() {
			element = x.Type.MapElement()
		}
		seen := map[string]bool{}
		for i, a := range x.Args {
			if seen[x.ArgNames[i]] {
				return "", diagnostics.New(a.Pos, "type error: duplicate Map key %s", x.ArgNames[i])
			}
			seen[x.ArgNames[i]] = true
			t, e := c.expected(a, env, element)
			if e != nil {
				return "", e
			}
			if element == "" {
				element = t
			}
			if t != element || !c.valid(t, false) {
				return "", diagnostics.New(a.Pos, "type error: Map values must be %s, got %s", element, t)
			}
		}
		if element == "" {
			return "", diagnostics.New(x.Pos, "type error: empty Map requires a declared Map type")
		}
		x.Type = element.Map()
		return x.Type, nil
	case "array":
		element := ast.Type("")
		if x.Type.IsArray() {
			element = x.Type.Element()
		}
		for _, a := range x.Args {
			t, e := c.expected(a, env, element)
			if e != nil {
				return "", e
			}
			if element == "" {
				element = t
			}
			if t != element || !c.valid(t, false) {
				return "", diagnostics.New(a.Pos, "type error: array elements must be %s, got %s", element, t)
			}
		}
		if element == "" {
			return "", diagnostics.New(x.Pos, "type error: empty array requires a declared array type")
		}
		x.Type = element.Array()
		return x.Type, nil
	case "index":
		t, e := c.expression(x.Left, env)
		if e != nil {
			return "", e
		}
		idx, e := c.expression(x.Right, env)
		if e != nil {
			return "", e
		}
		if !t.IsArray() || idx != ast.Int {
			return "", diagnostics.New(x.Pos, "type error: indexing requires an Array and Int index")
		}
		return t.Element(), nil
	case "member":
		t, e := c.expression(x.Left, env)
		if e != nil {
			return "", e
		}
		if x.Text == "length" && (t.IsArray() || t.IsMap() || t == ast.String) {
			return ast.Int, nil
		}
		if f, ok := c.field(t, x.Text); ok {
			return f.Type, nil
		}
		if t == ast.Exception {
			switch x.Text {
			case "message", "type", "code", "file":
				return ast.String, nil
			case "line", "column":
				return ast.Int, nil
			}
		}
		return "", diagnostics.New(x.Pos, "type error: %s has no property %s", t, x.Text)
	case "unary":
		if x.Text == "-" && x.Right.Kind == "literal" && x.Right.Type == ast.Int && x.Right.Text == "9223372036854775808" {
			return ast.Int, nil
		}
		t, e := c.expression(x.Right, env)
		if e != nil {
			return "", e
		}
		if x.Text == "!" && t == ast.Bool {
			return t, nil
		}
		if (x.Text == "-" || x.Text == "+") && (t == ast.Int || t == ast.Float) {
			return t, nil
		}
		return "", diagnostics.New(x.Pos, "type error: operator %s cannot be applied to %s", x.Text, t)
	case "assignment":
		if e := c.writable(x.Left, env); e != nil {
			return "", e
		}
		l, e := c.expression(x.Left, env)
		if e != nil {
			return "", e
		}
		r, e := c.expected(x.Right, env, l)
		if e != nil {
			return "", e
		}
		if l != r {
			return "", diagnostics.New(x.Pos, "type error: cannot assign %s to %s", r, l)
		}
		if x.Text != "=" {
			if _, e = binary(x.Text[:1], l, r, x.Pos); e != nil {
				return "", e
			}
		}
		return l, nil
	case "binary":
		l, e := c.expression(x.Left, env)
		if e != nil {
			return "", e
		}
		r, e := c.expression(x.Right, env)
		if e != nil {
			return "", e
		}
		return binary(x.Text, l, r, x.Pos)
	case "call":
		return c.call(x, env)
	}
	return "", diagnostics.New(x.Pos, "type error: unsupported expression %s", x.Kind)
}
func (c *checker) call(x *ast.Expr, env *resolver.Scope) (ast.Type, error) {
	name := stdlib.QualifiedName(x.Left)
	if name == "std.task.spawn" || name == "std.timer.after" || name == "std.timer.every" {
		return c.taskCall(x, env, name)
	}
	if name == "std.http.serve" {
		return c.httpCall(x, env)
	}
	if x.Left.Kind == "identifier" {
		if s, ok := env.Find(name); ok {
			return c.callValue(x, env, s.Type)
		}
	}
	if sig, ok := stdlib.Lookup(name); ok {
		if name == "std.time.sleep" && len(x.Args) == 1 {
			t, e := c.expression(x.Args[0], env)
			if e != nil {
				return "", e
			}
			if t != ast.Int && t != ast.Duration {
				return "", diagnostics.New(x.Pos, "type error: sleep requires Duration or Int milliseconds")
			}
			sig.Params = []ast.Type{t}
		}
		if name == "std.convert.toString" && len(x.Args) == 1 {
			t, e := c.expression(x.Args[0], env)
			if e != nil {
				return "", e
			}
			if !t.Primitive() {
				return "", diagnostics.New(x.Pos, "type error: toString requires a primitive value")
			}
			sig.Params = []ast.Type{t}
		}
		return c.arguments(x, env, sig)
	}
	if x.Left.Kind == "member" {
		t, e := c.expression(x.Left.Left, env)
		if e != nil {
			return "", e
		}
		if field, ok := c.field(t, x.Left.Text); ok {
			return c.callValue(x, env, field.Type)
		}
		sig, ok := stdlib.Method(t, x.Left.Text)
		if !ok {
			return "", diagnostics.New(x.Pos, "type error: %s has no method %s", t, x.Left.Text)
		}
		if stdlib.Mutates(t, x.Left.Text) {
			if e = c.writable(x.Left.Left, env); e != nil {
				return "", e
			}
		}
		return c.arguments(x, env, sig)
	}
	if x.Left.Kind != "identifier" {
		t, e := c.expression(x.Left, env)
		if e != nil {
			return "", e
		}
		return c.callValue(x, env, t)
	}
	f, ok := c.functions[name]
	if st := c.structs[name]; st != nil {
		f = &ast.Function{Name: name, Return: ast.Type(name)}
		for _, field := range st.Fields {
			f.Params = append(f.Params, ast.Parameter{Name: field.Name, Type: field.Type, Pos: field.Pos})
		}
		for _, n := range x.ArgNames {
			if n == "" {
				return "", diagnostics.New(x.Pos, "type error: struct constructors require named arguments")
			}
		}
		ok = true
	}
	if !ok {
		return "", diagnostics.New(x.Pos, "type error: unknown function %s", name)
	}
	if len(x.Args) != len(f.Params) {
		return "", diagnostics.New(x.Pos, "type error: %s expects %d arguments, got %d", name, len(f.Params), len(x.Args))
	}
	x.ArgOrder = make([]int, len(x.Args))
	named := len(x.ArgNames) > 0 && x.ArgNames[0] != ""
	seen := map[int]bool{}
	for i, a := range x.Args {
		argName := ""
		if i < len(x.ArgNames) {
			argName = x.ArgNames[i]
		}
		if (argName != "") != named {
			return "", diagnostics.New(a.Pos, "type error: use either all named or all positional arguments")
		}
		index := i
		if named {
			index = -1
			for j, p := range f.Params {
				if p.Name == argName {
					index = j
					break
				}
			}
			if index < 0 {
				return "", diagnostics.New(a.Pos, "type error: unknown parameter %s for %s", argName, name)
			}
		}
		if seen[index] {
			return "", diagnostics.New(a.Pos, "type error: duplicate argument %s", argName)
		}
		seen[index] = true
		x.ArgOrder[i] = index
		t, e := c.expected(a, env, f.Params[index].Type)
		if e != nil {
			return "", e
		}
		if t != f.Params[index].Type {
			return "", diagnostics.New(a.Pos, "type error: argument %d of %s must be %s, got %s", index+1, name, f.Params[index].Type, t)
		}
	}
	return f.Return, nil
}
func (c *checker) arguments(x *ast.Expr, env *resolver.Scope, sig stdlib.Signature) (ast.Type, error) {
	for _, name := range x.ArgNames {
		if name != "" {
			return "", diagnostics.New(x.Pos, "type error: built-ins and methods use positional arguments")
		}
	}
	if !sig.Variadic && len(x.Args) != len(sig.Params) {
		return "", diagnostics.New(x.Pos, "type error: call expects %d arguments, got %d", len(sig.Params), len(x.Args))
	}
	for i, a := range x.Args {
		want := ast.Type("")
		if !sig.Variadic {
			want = sig.Params[i]
		}
		t, e := c.expected(a, env, want)
		if e != nil {
			return "", e
		}
		if t == ast.Void {
			return "", diagnostics.New(a.Pos, "type error: print or function cannot accept Void")
		}
		if !sig.Variadic && t != want {
			return "", diagnostics.New(a.Pos, "type error: argument %d must be %s, got %s", i+1, want, t)
		}
	}
	return sig.Result, nil
}
func binary(op string, l, r ast.Type, p diagnostics.Position) (ast.Type, error) {
	if (op == "==" || op == "!=") && ((l.IsOptional() && r == ast.Null) || (r.IsOptional() && l == ast.Null) || (l == ast.Null && r == ast.Null)) {
		return ast.Bool, nil
	}
	if l == r && l.Primitive() {
		switch op {
		case "==", "!=":
			return ast.Bool, nil
		case "&&", "||":
			if l == ast.Bool {
				return ast.Bool, nil
			}
		case "<", ">", "<=", ">=":
			if l == ast.Int || l == ast.Float || l == ast.String {
				return ast.Bool, nil
			}
		case "+":
			if l == ast.String || l == ast.Int || l == ast.Float {
				return l, nil
			}
		case "-", "*", "/":
			if l == ast.Int || l == ast.Float {
				return l, nil
			}
		case "%":
			if l == ast.Int {
				return l, nil
			}
		}
	}
	return "", diagnostics.New(p, "type error: operator %s cannot combine %s and %s; operands must have compatible types", op, l, r)
}
