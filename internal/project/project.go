package project

import (
	"craft/internal/ast"
	"craft/internal/diagnostics"
	"craft/internal/lexer"
	"craft/internal/parser"
	"craft/internal/types"
	"craft/templates"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const MaxSourceSize = 2 * 1024 * 1024
const MaxProjectSize = 16 * 1024 * 1024

type Manifest struct {
	GitHub                        map[string]GitHubDependency
	Name, Version, Edition, Entry string
	Dependencies                  map[string]string
	Versions                      map[string]string
	DependencyPositions           map[string]diagnostics.Position
}
type Source struct {
	Path string `json:"path"`
	Text string `json:"text"`
}
type Bundle struct {
	Entry    string   `json:"entry,omitempty"`
	Modules  []Module `json:"modules,omitempty"`
	Format   string   `json:"format"`
	Version  int      `json:"version"`
	Name     string   `json:"name"`
	Language string   `json:"language"`
	Sources  []Source `json:"sources"`
}
type Project struct {
	Remotes  []RemotePin
	Modules  []Module
	Packages []ResolvedPackage
	Root     string
	Manifest Manifest
	Sources  []Source
}

var namePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

func validName(name string) bool {
	if !namePattern.MatchString(name) {
		return false
	}
	switch strings.ToUpper(name) {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return false
	}
	return len(name) <= 80
}
func Init(dir string, create bool) error {
	abs, e := filepath.Abs(dir)
	if e != nil {
		return e
	}
	name := filepath.Base(abs)
	if !validName(name) {
		return fmt.Errorf("project name must start with a letter, contain only letters, digits, '-' or '_', and not be a Windows reserved name")
	}
	if create {
		if _, e = os.Lstat(abs); !os.IsNotExist(e) {
			return fmt.Errorf("destination already exists or cannot be accessed: %s", abs)
		}
		if e = os.MkdirAll(abs, 0755); e != nil {
			return e
		}
	} else {
		info, e := os.Stat(abs)
		if e != nil {
			return e
		}
		if !info.IsDir() {
			return fmt.Errorf("project path must be a directory")
		}
	}
	for _, relative := range []string{"craft.toml", "src/main.craft", "tests/smoke.craft"} {
		if _, e := os.Lstat(filepath.Join(abs, relative)); !os.IsNotExist(e) {
			return fmt.Errorf("refusing to overwrite %s", relative)
		}
	}
	for _, relative := range []string{"src", "tests"} {
		path := filepath.Join(abs, relative)
		if info, e := os.Lstat(path); e == nil && (info.Mode()&os.ModeSymlink != 0 || !info.IsDir()) {
			return fmt.Errorf("%s must be a regular directory", relative)
		}
		if e := os.MkdirAll(path, 0755); e != nil {
			return e
		}
	}
	manifest := fmt.Sprintf("name = %q\nversion = \"0.1.0\"\nedition = \"2026\"\nentry = \"main\"\n", name)
	if e = writeNew(filepath.Join(abs, "craft.toml"), []byte(manifest)); e != nil {
		return e
	}
	if e = writeNew(filepath.Join(abs, "src", "main.craft"), []byte(templates.Main)); e != nil {
		return e
	}
	return writeNew(filepath.Join(abs, "tests", "smoke.craft"), []byte(templates.Test))
}
func writeNew(path string, data []byte) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if e != nil {
		return e
	}
	_, e = f.Write(data)
	closeErr := f.Close()
	if e != nil {
		return e
	}
	return closeErr
}
func Find(start string) (string, error) {
	dir, e := filepath.Abs(start)
	if e != nil {
		return "", e
	}
	for {
		if info, e := os.Stat(filepath.Join(dir, "craft.toml")); e == nil && !info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("craft.toml not found; run craft new <name> or craft init")
		}
		dir = parent
	}
}

// ReadManifest implements scalar metadata, the Rev.8 [dependencies] table and
// the legacy Rev.4 dotted dependency keys. Unsupported TOML features are
// rejected explicitly instead of being ignored.
func ReadManifest(root string) (Manifest, error) {
	path := filepath.Join(root, "craft.toml")
	data, e := readLimited(path, 64*1024)
	if e != nil {
		return Manifest{}, e
	}
	return readManifestData(root, data)
}
func readManifestData(root string, data []byte) (Manifest, error) {
	path := filepath.Join(root, "craft.toml")
	github := map[string]GitHubDependency{}
	values := map[string]string{}
	valuePositions := map[string]diagnostics.Position{}
	dependencies := map[string]string{}
	versions := map[string]string{}
	positions := map[string]diagnostics.Position{}
	canonicalDependencies := map[string]bool{}
	section := ""
	for i, line := range strings.Split(strings.TrimPrefix(string(data), "\ufeff"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := diagnostics.Position{File: path, Line: i + 1, Column: 1}
		if strings.HasPrefix(line, "[") {
			table := strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
			if table != "[dependencies]" {
				return Manifest{}, diagnostics.New(p, "manifest error: unsupported table %s", line)
			}
			if section != "" {
				return Manifest{}, diagnostics.New(p, "manifest error: duplicate table [dependencies]")
			}
			section = "dependencies"
			continue
		}
		k, raw, ok := strings.Cut(line, "=")
		k = strings.TrimSpace(k)
		if !ok {
			return Manifest{}, diagnostics.New(p, "manifest error: expected key = \"value\"")
		}
		if section == "dependencies" {
			if !validAlias(k) {
				return Manifest{}, diagnostics.New(p, "manifest error: invalid dependency alias %q", k)
			}
			if _, exists := dependencies[k]; exists {
				return Manifest{}, diagnostics.New(p, "manifest error: duplicate dependency %s", k)
			}
			fields, e := manifestDependencyFields(strings.TrimSpace(raw))
			if e != nil {
				return Manifest{}, diagnostics.New(p, "manifest error: dependency %s: %s", k, e)
			}
			depPath, version := fields["path"], fields["version"]
			if fields["github"] != "" {
				remote := GitHubDependency{Repository: strings.ToLower(fields["github"]), Tag: fields["tag"], Version: version}
				if e := remote.validate(); e != nil {
					return Manifest{}, diagnostics.New(p, "manifest error: dependency %s: %s", k, e)
				}
				github[k] = remote
			}
			dependencies[k], versions[k], positions[k] = depPath, version, p
			canonicalDependencies[k] = true
			continue
		}
		switch k {
		case "name", "version", "edition", "entry":
		default:
			if !strings.HasPrefix(k, "dependency.") && !strings.HasPrefix(k, "dependency-version.") {
				return Manifest{}, diagnostics.New(p, "manifest error: unsupported key %q", k)
			}
		}
		if _, ok := values[k]; ok {
			return Manifest{}, diagnostics.New(p, "manifest error: duplicate key %s", k)
		}
		value, e := manifestString(strings.TrimSpace(raw))
		if e != nil {
			return Manifest{}, diagnostics.New(p, "manifest error: %s", e)
		}
		values[k] = value
		valuePositions[k] = p
	}
	m := Manifest{Name: values["name"], Version: values["version"], Edition: values["edition"], Entry: values["entry"]}
	m.Dependencies = dependencies
	m.GitHub = github
	m.Versions = versions
	m.DependencyPositions = positions
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := values[k]
		if a, ok := strings.CutPrefix(k, "dependency."); ok {
			if !validAlias(a) || v == "" {
				return m, diagnostics.New(valuePositions[k], "manifest error: invalid dependency %s", a)
			}
			if _, exists := m.Dependencies[a]; exists {
				return m, diagnostics.New(valuePositions[k], "manifest error: duplicate dependency %s", a)
			}
			m.Dependencies[a] = v
			m.DependencyPositions[a] = valuePositions[k]
		}
	}
	for _, k := range keys {
		v := values[k]
		if a, ok := strings.CutPrefix(k, "dependency-version."); ok {
			if !validAlias(a) || v == "" {
				return m, diagnostics.New(valuePositions[k], "manifest error: invalid dependency version %s", a)
			}
			if canonicalDependencies[a] {
				return m, diagnostics.New(valuePositions[k], "manifest error: duplicate dependency version %s", a)
			}
			m.Versions[a] = v
		}
	}
	for a := range m.Dependencies {
		if m.Versions[a] == "" {
			return m, fmt.Errorf("dependency-version.%s is required", a)
		}
	}
	for a := range m.Versions {
		if _, ok := m.Dependencies[a]; !ok {
			return m, fmt.Errorf("dependency-version without dependency: %s", a)
		}
	}
	if !validName(m.Name) {
		return m, fmt.Errorf("manifest error: invalid or missing project name")
	}
	if m.Version == "" {
		return m, fmt.Errorf("manifest error: version is required")
	}
	if len(m.Version) > 80 || strings.ContainsAny(m.Version, "@/\\:") {
		return m, fmt.Errorf("manifest error: invalid version")
	}
	if m.Edition != "2026" {
		return m, fmt.Errorf("manifest error: edition must be \"2026\"")
	}
	if m.Entry == "" {
		m.Entry = "main"
	}
	if m.Entry != "main" && m.Entry != "library" {
		return m, fmt.Errorf("manifest error: entry must be main or library")
	}
	return m, nil
}

func manifestDependencyFields(raw string) (map[string]string, error) {
	if !strings.HasPrefix(raw, "{") {
		return nil, fmt.Errorf("expected inline dependency table")
	}
	end, quote, escaped := -1, byte(0), false
	for i := 1; i < len(raw); i++ {
		c := raw[i]
		if quote != 0 {
			if quote == '"' && escaped {
				escaped = false
				continue
			}
			if quote == '"' && c == '\\' {
				escaped = true
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			continue
		}
		if c == '}' {
			end = i
			break
		}
	}
	if end < 0 || quote != 0 {
		return nil, fmt.Errorf("unterminated dependency table")
	}
	tail := strings.TrimSpace(raw[end+1:])
	if tail != "" && !strings.HasPrefix(tail, "#") {
		return nil, fmt.Errorf("unexpected text after dependency")
	}
	fields, e := splitManifestFields(raw[1:end])
	if e != nil {
		return nil, e
	}
	values := map[string]string{}
	for _, field := range fields {
		k, value, ok := strings.Cut(field, "=")
		k = strings.TrimSpace(k)
		if !ok || (k != "path" && k != "version" && k != "github" && k != "tag") {
			return nil, fmt.Errorf("expected path/version or github/tag/version fields")
		}
		if _, exists := values[k]; exists {
			return nil, fmt.Errorf("duplicate field %s", k)
		}
		values[k], e = manifestString(strings.TrimSpace(value))
		if e != nil {
			return nil, fmt.Errorf("%s: %s", k, e)
		}
	}
	if values["version"] == "" {
		return nil, fmt.Errorf("version is required")
	}
	_, local := values["path"]
	if local {
		if len(values) != 2 || values["path"] == "" {
			return nil, fmt.Errorf("local dependency requires only path and version")
		}
	} else if len(values) != 3 || values["github"] == "" || values["tag"] == "" {
		return nil, fmt.Errorf("remote dependency requires github, tag and version")
	}
	return values, nil
}

func splitManifestFields(raw string) ([]string, error) {
	fields, start, quote, escaped := []string{}, 0, byte(0), false
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if quote != 0 {
			if quote == '"' && escaped {
				escaped = false
				continue
			}
			if quote == '"' && c == '\\' {
				escaped = true
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
		} else if c == ',' {
			if field := strings.TrimSpace(raw[start:i]); field != "" {
				fields = append(fields, field)
			}
			start = i + 1
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated string")
	}
	if field := strings.TrimSpace(raw[start:]); field != "" {
		fields = append(fields, field)
	}
	return fields, nil
}
func manifestString(raw string) (string, error) {
	if len(raw) < 2 || (raw[0] != '"' && raw[0] != '\'') {
		return "", fmt.Errorf("expected a quoted string")
	}
	quote := raw[0]
	end := -1
	for i := 1; i < len(raw); i++ {
		if quote == '"' && raw[i] == '\\' {
			i++
			continue
		}
		if raw[i] == quote {
			end = i
			break
		}
	}
	if end < 0 {
		return "", fmt.Errorf("unterminated string")
	}
	tail := strings.TrimSpace(raw[end+1:])
	if tail != "" && !strings.HasPrefix(tail, "#") {
		return "", fmt.Errorf("unexpected text after string")
	}
	if quote == '\'' {
		return raw[1:end], nil
	}
	v, e := strconv.Unquote(raw[:end+1])
	if e != nil {
		return "", fmt.Errorf("invalid string escape")
	}
	return v, nil
}
func readLimited(path string, limit int64) ([]byte, error) {
	info, e := os.Stat(path)
	if e != nil {
		return nil, e
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("expected regular file: %s", path)
	}
	if info.Size() > limit {
		return nil, fmt.Errorf("file exceeds size limit: %s", path)
	}
	data, e := os.ReadFile(path)
	if e == nil && int64(len(data)) > limit {
		return nil, fmt.Errorf("file exceeds size limit: %s", path)
	}
	return data, e
}
func Load(start string) (*Project, error)      { return loadGraph(start, false) }
func LoadTests(start string) (*Project, error) { return loadGraph(start, true) }
func load(start string, includeTests bool) (*Project, error) {
	root, e := Find(start)
	if e != nil {
		return nil, e
	}
	manifest, e := ReadManifest(root)
	if e != nil {
		return nil, e
	}
	p := &Project{Root: root, Manifest: manifest}
	total := 0
	roots := []string{"src"}
	if includeTests {
		roots = append(roots, "tests")
	}
	for _, relative := range roots {
		sourceRoot := filepath.Join(root, relative)
		if info, err := os.Stat(sourceRoot); err == nil && !info.IsDir() {
			return nil, fmt.Errorf("%s must be a directory", relative)
		} else if os.IsNotExist(err) && includeTests {
			continue
		}
		e = filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("source symlinks are not supported: %s", path)
			}
			if d.IsDir() && path != sourceRoot {
				switch d.Name() {
				case "dist", "build", "node_modules", ".git":
					return filepath.SkipDir
				}
			}
			if d.IsDir() || filepath.Ext(path) != ".craft" {
				return nil
			}
			data, e := readLimited(path, MaxSourceSize)
			if e != nil {
				return e
			}
			total += len(data)
			if total > MaxProjectSize || len(p.Sources) >= 1024 {
				return fmt.Errorf("project exceeds Phase 1 limits (16 MiB or 1024 source files)")
			}
			rel, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			p.Sources = append(p.Sources, Source{Path: filepath.ToSlash(rel), Text: string(data)})
			return nil
		})
		if e != nil {
			return nil, e
		}
	}
	if len(p.Sources) == 0 {
		return nil, fmt.Errorf("no .craft source files found in the selected source directories")
	}
	sort.Slice(p.Sources, func(i, j int) bool { return p.Sources[i].Path < p.Sources[j].Path })
	return p, nil
}
func (p *Project) Compile() (*ast.Program, map[string]string, error) {
	return p.CompileMode(p.Manifest.Entry != "library")
}
func (p *Project) CompileMode(requireMain bool) (*ast.Program, map[string]string, error) {
	if len(p.Modules) > 0 {
		return p.compileModules(requireMain)
	}
	program := &ast.Program{}
	sources := map[string]string{}
	for _, s := range p.Sources {
		sources[s.Path] = s.Text
	}
	for _, s := range p.Sources {
		tokens, e := lexer.Scan(s.Path, s.Text)
		if e != nil {
			return nil, sources, e
		}
		unit, e := parser.Parse(tokens)
		if e != nil {
			return nil, sources, e
		}
		if len(unit.Imports) != 0 {
			return nil, sources, fmt.Errorf("imports require a Rev.4 module graph; load the project manifest or rebuild the bundle")
		}
		program.Structs = append(program.Structs, unit.Structs...)
		program.Functions = append(program.Functions, unit.Functions...)
		program.Tests = append(program.Tests, unit.Tests...)
	}
	if e := types.CheckMode(program, requireMain); e != nil {
		return nil, sources, e
	}
	return program, sources, nil
}
func (p *Project) Build() (string, error) {
	dist := filepath.Join(p.Root, "dist")
	if info, e := os.Lstat(dist); e == nil && (info.Mode()&os.ModeSymlink != 0 || !info.IsDir()) {
		return "", fmt.Errorf("dist must be a regular directory")
	}
	if e := os.MkdirAll(dist, 0755); e != nil {
		return "", e
	}
	path := filepath.Join(dist, p.Manifest.Name+".craftbundle")
	if info, e := os.Lstat(path); e == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("refusing to overwrite non-regular artifact")
		}
		if e := recognizedArtifact(path); e != nil {
			return "", fmt.Errorf("refusing to overwrite unrecognized artifact: %w", e)
		}
	}
	data, e := json.MarshalIndent(Bundle{Format: "craft-source-bundle", Version: 3, Language: "0.1.11", Name: p.Manifest.Name, Entry: p.Manifest.Entry, Sources: p.Sources, Modules: p.Modules}, "", "  ")
	if e != nil {
		return "", e
	}
	f, e := os.CreateTemp(dist, ".craft-build-*")
	if e != nil {
		return "", e
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, e = f.Write(append(data, '\n')); e != nil {
		f.Close()
		return "", e
	}
	if e = f.Close(); e != nil {
		return "", e
	}
	if e = os.Rename(temp, path); e != nil {
		return "", e
	}
	return path, nil
}
func LoadBundle(path string) (*Project, error) {
	data, e := readLimited(path, MaxProjectSize*8)
	if e != nil {
		return nil, e
	}
	var b Bundle
	if e = json.Unmarshal(data, &b); e != nil {
		return nil, fmt.Errorf("invalid Craft bundle: %w", e)
	}
	if b.Version == 3 {
		return loadModuleBundle(b)
	}
	if b.Format != "craft-source-bundle" || b.Version != 2 || (b.Language != "0.1.1" && b.Language != "0.1.2" && b.Language != "0.1.3") || !validName(b.Name) {
		return nil, fmt.Errorf("unsupported Craft bundle format or language; migrate mutable let declarations to var and rebuild with Craft 0.1.7")
	}
	if len(b.Sources) == 0 || len(b.Sources) > 1024 {
		return nil, fmt.Errorf("invalid bundle source count")
	}
	seen := map[string]bool{}
	total := 0
	for _, s := range b.Sources {
		clean := filepath.ToSlash(filepath.Clean(s.Path))
		if clean != s.Path || !strings.HasPrefix(clean, "src/") || strings.Contains(clean, "\\") || filepath.Ext(clean) != ".craft" || seen[clean] {
			return nil, fmt.Errorf("invalid or duplicate bundle source path: %s", s.Path)
		}
		seen[clean] = true
		total += len(s.Text)
		if len(s.Text) > MaxSourceSize || total > MaxProjectSize {
			return nil, fmt.Errorf("bundle exceeds source size limits")
		}
	}
	return &Project{Manifest: Manifest{Name: b.Name, Version: "0.1.0", Edition: "2026", Entry: "main"}, Sources: b.Sources}, nil
}
func Clean(start string) error {
	root, e := Find(start)
	if e != nil {
		return e
	}
	m, e := ReadManifest(root)
	if e != nil {
		return e
	}
	dist := filepath.Join(root, "dist")
	info, e := os.Lstat(dist)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("dist must be a regular directory")
	}
	path := filepath.Join(dist, m.Name+".craftbundle")
	info, e = os.Lstat(path)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refusing to clean a non-regular artifact")
	}
	if e = recognizedArtifact(path); e != nil {
		return fmt.Errorf("refusing to clean an unrecognized artifact: %w", e)
	}
	return os.Remove(path)
}

// Legacy bundles may be replaced/cleaned, but must be rebuilt before execution.
func recognizedArtifact(path string) error {
	data, e := readLimited(path, MaxProjectSize*8)
	if e != nil {
		return e
	}
	var b Bundle
	if e = json.Unmarshal(data, &b); e != nil {
		return e
	}
	if b.Format != "craft-source-bundle" || (b.Version != 1 && b.Version != 2 && b.Version != 3) || !validName(b.Name) {
		return fmt.Errorf("unrecognized Craft artifact")
	}
	return nil
}
