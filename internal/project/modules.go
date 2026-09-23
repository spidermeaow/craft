package project

import (
	"context"
	"craft/internal/ast"
	"craft/internal/diagnostics"
	"craft/internal/lexer"
	"craft/internal/parser"
	"craft/internal/stdlib"
	"craft/internal/types"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

// One manifest is one module. Imports are file-local, exports are module-wide.
type Module struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
	Sources      []Source          `json:"sources"`
}

var aliasPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func validAlias(a string) bool { return aliasPattern.MatchString(a) && a != "std" }
func loadGraph(start string, tests bool) (*Project, error) {
	return loadGraphMode(start, tests, true)
}

func loadGraphMode(start string, tests, verifyLock bool) (*Project, error) {
	root, e := load(start, tests)
	if e != nil {
		return nil, e
	}
	if e = checkInstallJournal(root.Root); e != nil {
		return nil, e
	}
	lock, e := optionalLock(root.Root)
	if e != nil {
		return nil, e
	}
	r, e := newRemoteResolver(context.Background(), lock, false)
	if e != nil {
		return nil, e
	}
	return resolveGraph(root, verifyLock, r)
}

type remoteLocation struct {
	root string
	pin  RemotePin
}

func resolveGraph(root *Project, verifyLock bool, r *remoteResolver) (*Project, error) {
	var e error
	rootID := root.Manifest.Name + "@" + root.Manifest.Version
	mods := map[string]Module{}
	paths := map[string]string{}
	packagePaths := map[string]string{}
	visiting := map[string]bool{}
	identities := map[string]string{}
	locations := map[string]remoteLocation{}
	remoteSources := map[string]string{}
	totalBytes, totalFiles := 0, 0
	var visit func(*Project, []string) error
	visit = func(p *Project, chain []string) error {
		abs, e := filepath.EvalSymlinks(p.Root)
		if e != nil {
			return e
		}
		abs, e = filepath.Abs(abs)
		if e != nil {
			return e
		}
		key := abs
		if runtime.GOOS == "windows" {
			key = strings.ToLower(abs)
		}
		id := p.Manifest.Name + "@" + p.Manifest.Version
		if old, ok := identities[p.Manifest.Name]; ok && old != id {
			return fmt.Errorf("version conflict along %s: %s conflicts with %s", strings.Join(append(chain, id), " -> "), id, old)
		}
		identities[p.Manifest.Name] = id
		if visiting[key] {
			return fmt.Errorf("dependency cycle: %s", strings.Join(append(chain, id), " -> "))
		}
		if old, ok := paths[id]; ok {
			if old != key {
				return fmt.Errorf("module identity %s resolves to multiple paths", id)
			}
			return nil
		}
		if len(paths) >= 64 {
			return fmt.Errorf("dependency graph exceeds 64 modules")
		}
		visiting[key] = true
		paths[id] = key
		packagePaths[id] = abs
		if loc, ok := locations[p.Root]; ok {
			rel, err := filepath.Rel(loc.root, p.Root)
			if err != nil {
				return err
			}
			remoteSources[id] = "github:" + loc.pin.Repository + "@" + loc.pin.Commit
			if rel != "." {
				remoteSources[id] += "/" + filepath.ToSlash(rel)
			}
		}
		for _, source := range p.Sources {
			totalBytes += len(source.Text)
			totalFiles++
		}
		if totalBytes > MaxProjectSize || totalFiles > 1024 {
			return fmt.Errorf("module graph exceeds 16 MiB/1024 files")
		}
		m := Module{ID: id, Name: p.Manifest.Name, Version: p.Manifest.Version, Dependencies: map[string]string{}, Sources: p.Sources}
		aliases := []string{}
		for a := range p.Manifest.Dependencies {
			aliases = append(aliases, a)
		}
		sort.Strings(aliases)
		for _, a := range aliases {
			dir := filepath.Join(p.Root, p.Manifest.Dependencies[a])
			if filepath.IsAbs(p.Manifest.Dependencies[a]) {
				dir = p.Manifest.Dependencies[a]
			}
			var location remoteLocation
			remote := false
			if dep, ok := p.Manifest.GitHub[a]; ok {
				var pin RemotePin
				dir, pin, e = r.ensure(dep)
				if e != nil {
					return dependencyError(p.Manifest, a, "%s", e)
				}
				location = remoteLocation{dir, pin}
				remote = true
			} else if parent, ok := locations[p.Root]; ok {
				resolved, err := filepath.EvalSymlinks(dir)
				if err != nil {
					return dependencyError(p.Manifest, a, "cannot resolve local package in remote tree")
				}
				base, err := filepath.EvalSymlinks(parent.root)
				if err != nil {
					return err
				}
				rel, err := filepath.Rel(base, resolved)
				if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
					return dependencyError(p.Manifest, a, "local dependency escapes GitHub source tree")
				}
				location = parent
				remote = true
			}
			if _, err := os.Stat(filepath.Join(dir, "craft.toml")); err != nil {
				return dependencyError(p.Manifest, a, "missing package manifest in %q", dir)
			}
			d, e := load(dir, false)
			if e != nil {
				return dependencyError(p.Manifest, a, "cannot load path %q: %s", filepath.Clean(dir), e)
			}
			resolved, e := filepath.EvalSymlinks(dir)
			if e != nil {
				return dependencyError(p.Manifest, a, "cannot resolve path %q: %s", filepath.Clean(dir), e)
			}
			actual, e := filepath.EvalSymlinks(d.Root)
			if e != nil {
				return dependencyError(p.Manifest, a, "cannot resolve package root: %s", e)
			}
			if !strings.EqualFold(resolved, actual) {
				return dependencyError(p.Manifest, a, "path %q must point to a package root", filepath.Clean(dir))
			}
			if d.Manifest.Version != p.Manifest.Versions[a] {
				return dependencyError(p.Manifest, a, "version mismatch: want %s, got %s", p.Manifest.Versions[a], d.Manifest.Version)
			}
			if remote {
				locations[d.Root] = location
			}
			m.Dependencies[a] = d.Manifest.Name + "@" + d.Manifest.Version
			if e = visit(d, append(chain, id)); e != nil {
				var diagnostic *diagnostics.Error
				if errors.As(e, &diagnostic) {
					return e
				}
				return dependencyError(p.Manifest, a, "%s", e)
			}
		}
		visiting[key] = false
		mods[id] = m
		return nil
	}
	if e = visit(root, nil); e != nil {
		return nil, e
	}
	root.Modules = []Module{mods[rootID]}
	ids := []string{}
	for id := range mods {
		if id != rootID {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		root.Modules = append(root.Modules, mods[id])
	}
	if e = validateModules(root.Modules, true); e != nil {
		return nil, e
	}
	root.Packages, e = resolvedPackages(root.Root, root.Modules, packagePaths)
	if e != nil {
		return nil, e
	}
	for i := range root.Packages {
		if s, ok := remoteSources[root.Packages[i].ID]; ok {
			root.Packages[i].Source = s
		}
	}
	root.Remotes = r.usedPins()
	if verifyLock {
		if _, e = verifyProjectLock(root); e != nil {
			return nil, e
		}
	}
	return root, nil
}

func dependencyError(m Manifest, alias, format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	if p, ok := m.DependencyPositions[alias]; ok {
		return diagnostics.New(p, "package error: dependency %s: %s", alias, message)
	}
	return fmt.Errorf("package error: dependency %s: %s", alias, message)
}

func validateModules(mods []Module, tests bool) error {
	if len(mods) == 0 || len(mods) > 64 {
		return fmt.Errorf("invalid module count")
	}
	known := map[string]bool{}
	count, total := 0, 0
	for i, m := range mods {
		if !validName(m.Name) || m.Version == "" || len(m.Version) > 80 || strings.ContainsAny(m.Version, "@/\\:") || m.ID != m.Name+"@"+m.Version || known[m.ID] {
			return fmt.Errorf("invalid/duplicate module identity %s", m.ID)
		}
		known[m.ID] = true
		seen := map[string]bool{}
		if len(m.Sources) == 0 {
			return fmt.Errorf("empty module %s", m.ID)
		}
		for _, s := range m.Sources {
			clean := filepath.ToSlash(filepath.Clean(s.Path))
			allowed := strings.HasPrefix(s.Path, "src/") || (i == 0 && tests && strings.HasPrefix(s.Path, "tests/"))
			if clean != s.Path || strings.Contains(s.Path, "\\") || !allowed || filepath.Ext(s.Path) != ".craft" || seen[s.Path] {
				return fmt.Errorf("invalid module source %s", s.Path)
			}
			seen[s.Path] = true
			count++
			total += len(s.Text)
			if len(s.Text) > MaxSourceSize || total > MaxProjectSize || count > 1024 {
				return fmt.Errorf("module graph exceeds 16 MiB/1024 files")
			}
		}
	}
	state := map[string]int{}
	byID := map[string]Module{}
	for _, m := range mods {
		byID[m.ID] = m
	}
	var visit func(string, []string) error
	visit = func(id string, chain []string) error {
		if state[id] == 1 {
			return fmt.Errorf("dependency cycle: %s", strings.Join(append(chain, id), " -> "))
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for a, dep := range byID[id].Dependencies {
			if !validAlias(a) || !known[dep] {
				return fmt.Errorf("invalid dependency %s in %s", a, id)
			}
			if e := visit(dep, append(chain, id)); e != nil {
				return e
			}
		}
		state[id] = 2
		return nil
	}
	for _, m := range mods {
		if e := visit(m.ID, nil); e != nil {
			return e
		}
	}
	return nil
}
func loadModuleBundle(b Bundle) (*Project, error) {
	if b.Entry == "" {
		b.Entry = "main"
	}
	if b.Entry != "main" && b.Entry != "library" {
		return nil, fmt.Errorf("invalid bundle entry")
	}
	if b.Format != "craft-source-bundle" || (b.Language != "0.1.4" && b.Language != "0.1.5" && b.Language != "0.1.6" && b.Language != "0.1.7" && b.Language != "0.1.10" && b.Language != "0.1.11") || !validName(b.Name) {
		return nil, fmt.Errorf("unsupported Rev.4 bundle")
	}
	if len(b.Modules) == 0 {
		b.Modules = []Module{{ID: b.Name + "@0.1.0", Name: b.Name, Version: "0.1.0", Sources: b.Sources}}
	}
	if e := validateModules(b.Modules, false); e != nil {
		return nil, e
	}
	if b.Modules[0].Name != b.Name {
		return nil, fmt.Errorf("bundle root name mismatch")
	}
	return &Project{Manifest: Manifest{Name: b.Name, Version: b.Modules[0].Version, Edition: "2026", Entry: b.Entry}, Sources: b.Modules[0].Sources, Modules: b.Modules}, nil
}

type symbol struct {
	name      string
	exported  bool
	structure bool
}

func (p *Project) compileModules(requireMain bool) (*ast.Program, map[string]string, error) {
	texts := map[string]string{}
	units := map[string][]*ast.Program{}
	symbols := map[string]map[string]symbol{}
	mods := map[string]Module{}
	if e := validateModules(p.Modules, true); e != nil {
		return nil, texts, e
	}
	root := p.Modules[0].ID
	for _, m := range p.Modules {
		mods[m.ID] = m
		symbols[m.ID] = map[string]symbol{}
		for _, s := range m.Sources {
			file := s.Path
			if m.ID != root {
				file = "packages/" + m.ID + "/" + file
			}
			texts[file] = s.Text
			t, e := lexer.Scan(file, s.Text)
			if e != nil {
				return nil, texts, e
			}
			u, e := parser.Parse(t)
			if e != nil {
				return nil, texts, e
			}
			units[m.ID] = append(units[m.ID], u)
			add := func(name string, exported, structure bool, pos diagnostics.Position) error {
				_, builtin := stdlib.Lookup(name)
				reserved := name == "std" || builtin || name == "HttpRequest" || name == "HttpResponse" || name == "HttpConfig"
				for _, native := range stdlib.NativeStructs() {
					if native.Name == name {
						reserved = true
					}
				}
				if structure {
					switch name {
					case "Int", "Float", "Bool", "String", "Void", "Exception", "Map", "Fn", "JsonValue", "DateTime", "Duration", "Random", "Task", "Timer", "Bytes", "Context", "DbConnection", "DbTransaction", "DbValue":
						reserved = true
					}
				}
				if reserved {
					return diagnostics.New(pos, "type error: reserved declaration name %s", name)
				}
				if _, ok := symbols[m.ID][name]; ok {
					return diagnostics.New(pos, "type error: duplicate declaration %s", name)
				}
				n := name
				if m.ID != root {
					n = m.ID + "::" + name
				}
				symbols[m.ID][name] = symbol{n, exported, structure}
				return nil
			}
			for _, s := range u.Structs {
				if e = add(s.Name, s.Exported, true, s.Pos); e != nil {
					return nil, texts, e
				}
			}
			for _, f := range u.Functions {
				if e = add(f.Name, f.Exported, false, f.Pos); e != nil {
					return nil, texts, e
				}
			}
		}
	}
	result := &ast.Program{}
	for _, m := range p.Modules {
		for _, u := range units[m.ID] {
			aliases := map[string]string{}
			imported := map[string]bool{}
			for _, imp := range u.Imports {
				dep := m.Dependencies[imp.Dependency]
				if !validAlias(imp.Alias) || dep == "" || aliases[imp.Alias] != "" || imported[dep] {
					return nil, texts, fmt.Errorf("invalid/duplicate import %s in %s", imp.Alias, m.ID)
				}
				if _, ok := symbols[m.ID][imp.Alias]; ok {
					return nil, texts, fmt.Errorf("import alias conflicts with declaration %s", imp.Alias)
				}
				aliases[imp.Alias] = dep
				imported[dep] = true
			}
			lookup := func(n string) (string, error) {
				if a, b, ok := strings.Cut(n, "."); ok {
					dep := aliases[a]
					s, exists := symbols[dep][b]
					if dep == "" || !exists || !s.exported {
						return "", fmt.Errorf("unknown or private imported symbol %s in %s", n, m.ID)
					}
					return s.name, nil
				}
				if s, ok := symbols[m.ID][n]; ok {
					return s.name, nil
				}
				return n, nil
			}
			var typ func(ast.Type) (ast.Type, error)
			typ = func(t ast.Type) (ast.Type, error) {
				if t.IsOptional() {
					v, e := typ(t.Base())
					return v.Optional(), e
				}
				if t.IsArray() {
					v, e := typ(t.Element())
					return v.Array(), e
				}
				if t.IsMap() {
					v, e := typ(t.MapElement())
					return v.Map(), e
				}
				if r, ps, ok := t.Function(); ok {
					r, e := typ(r)
					if e != nil {
						return "", e
					}
					for i := range ps {
						ps[i], e = typ(ps[i])
						if e != nil {
							return "", e
						}
					}
					return ast.FunctionType(r, ps), nil
				}
				n, e := lookup(string(t))
				return ast.Type(n), e
			}
			clone := func(s map[string]bool) map[string]bool {
				r := map[string]bool{}
				for k, v := range s {
					r[k] = v
				}
				return r
			}
			var expr func(*ast.Expr, map[string]bool) error
			expr = func(x *ast.Expr, locals map[string]bool) error {
				if x == nil {
					return nil
				}
				var e error
				if x.Kind == "member" && x.Left.Kind == "identifier" && aliases[x.Left.Text] != "" && !locals[x.Left.Text] {
					n, e := lookup(x.Left.Text + "." + x.Text)
					if e != nil {
						return diagnostics.New(x.Pos, "type error: %s", e)
					}
					x.Kind = "identifier"
					x.Text = n
					x.Left = nil
					return nil
				}
				if x.Kind == "identifier" && !locals[x.Text] {
					x.Text, e = lookup(x.Text)
					if e != nil {
						return e
					}
				}
				x.Type, e = typ(x.Type)
				if e != nil {
					return e
				}
				if e = expr(x.Left, locals); e != nil {
					return e
				}
				if e = expr(x.Right, locals); e != nil {
					return e
				}
				for _, a := range x.Args {
					if e = expr(a, locals); e != nil {
						return e
					}
				}
				return nil
			}
			var stmt func(*ast.Stmt, map[string]bool) error
			stmt = func(s *ast.Stmt, locals map[string]bool) error {
				if s == nil {
					return nil
				}
				var e error
				s.Type, e = typ(s.Type)
				if e != nil {
					return e
				}
				if e = expr(s.Expr, locals); e != nil {
					return e
				}
				if e = expr(s.End, locals); e != nil {
					return e
				}
				if s.Kind == "block" {
					scope := clone(locals)
					for _, st := range s.Statements {
						if e = stmt(st, scope); e != nil {
							return e
						}
					}
					return nil
				}
				if s.Kind == "let" || s.Kind == "var" {
					if aliases[s.Name] != "" {
						return diagnostics.New(s.Pos, "type error: local shadows import alias %s", s.Name)
					}
					locals[s.Name] = true
				}
				body := clone(locals)
				if s.Kind == "for" || s.Kind == "iflet" {
					if aliases[s.Name] != "" {
						return diagnostics.New(s.Pos, "type error: local shadows import alias %s", s.Name)
					}
					body[s.Name] = true
				}
				if e = stmt(s.Body, body); e != nil {
					return e
				}
				other := clone(locals)
				if s.Kind == "try" {
					if aliases[s.Name] != "" {
						return diagnostics.New(s.Pos, "type error: local shadows import alias %s", s.Name)
					}
					other[s.Name] = true
				}
				return stmt(s.Else, other)
			}
			for _, s := range u.Structs {
				s.Name = symbols[m.ID][s.Name].name
				for i := range s.Fields {
					v, e := typ(s.Fields[i].Type)
					if e != nil {
						return nil, texts, e
					}
					s.Fields[i].Type = v
				}
				result.Structs = append(result.Structs, s)
			}
			for _, f := range u.Functions {
				f.Name = symbols[m.ID][f.Name].name
				v, e := typ(f.Return)
				if e != nil {
					return nil, texts, e
				}
				f.Return = v
				locals := map[string]bool{}
				for i := range f.Params {
					a := &f.Params[i]
					a.Type, e = typ(a.Type)
					if e != nil {
						return nil, texts, e
					}
					if aliases[a.Name] != "" {
						return nil, texts, fmt.Errorf("parameter shadows import alias %s", a.Name)
					}
					locals[a.Name] = true
				}
				if e = stmt(f.Body, locals); e != nil {
					return nil, texts, e
				}
				result.Functions = append(result.Functions, f)
			}
			if m.ID == root {
				for _, t := range u.Tests {
					if e := stmt(t.Body, map[string]bool{}); e != nil {
						return nil, texts, e
					}
					result.Tests = append(result.Tests, t)
				}
			}
		}
	}
	if e := types.CheckMode(result, requireMain); e != nil {
		return nil, texts, e
	}
	return result, texts, nil
}
