package cli

import (
	"context"
	"craft/internal/diagnostics"
	"craft/internal/formatter"
	"craft/internal/interpreter"
	"craft/internal/project"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const Version = "0.1.11"
const help = `Craft 0.1.11 — Phase 1 Rev.11

Usage:
  craft version                Show toolchain version
  craft new <directory>        Create a project
  craft init                   Initialize the current directory
  craft install [github.com/owner/repo@v1.2.3] [--alias name] [--offline] [--refresh]
  craft install --recover      Recover an interrupted package transaction
  craft run [bundle] [-- args...]           Check and interpret a project or bundle
  craft check                  Check src/ and tests/ without running
  craft build [--release]      Create a source bundle in dist/
  craft package check          Validate dependency graph and lock (offline)
  craft package list           List resolved local and GitHub packages
  craft package lock           Write craft.lock for the resolved graph
  craft clean                  Remove this project's generated bundle
  craft fmt [--check] [file]   Format project or one .craft file
  craft test                   Run test blocks in src/ and tests/
  craft help                   Show this help

Formatting examples: craft fmt --check | craft fmt path/to/file.craft

let is immutable; use var for changing values. Arrays have value semantics.
Bundles require Craft; native compilation is not available.
Exit codes: 0 success, 1 runtime/test failure, 2 syntax, 3 type, 4 project, 5 usage.
`

func Run(ctx context.Context, args []string, cwd string, out, errOut io.Writer) int {
	return RunWithInput(ctx, args, cwd, os.Stdin, out, errOut)
}

func RunWithInput(ctx context.Context, args []string, cwd string, in io.Reader, out, errOut io.Writer) int {
	sources := map[string]string{}
	fail := func(err error) int {
		var d *diagnostics.Error
		if !errors.As(err, &d) {
			err = &diagnostics.Error{Code: "E3001", Exit: 1, Message: err.Error()}
		}
		fmt.Fprintln(errOut, diagnostics.Format(err, sources))
		return diagnostics.ExitCode(err)
	}
	usage := func(message string) int { return fail(&diagnostics.Error{Code: "E5001", Exit: 5, Message: message}) }
	projectError := func(err error) int {
		var d *diagnostics.Error
		if errors.As(err, &d) {
			return fail(err)
		}
		return fail(&diagnostics.Error{Code: "E4002", Exit: 4, Message: err.Error()})
	}
	if len(args) == 0 {
		fmt.Fprint(out, help)
		return 0
	}
	command := args[0]
	switch command {
	case "help", "--help", "-h":
		if len(args) != 1 {
			return usage("usage: craft help")
		}
		fmt.Fprint(out, help)
		return 0
	case "version", "--version":
		if len(args) != 1 {
			return usage("usage: craft version")
		}
		fmt.Fprintf(out, "Craft %s\nLanguage: Phase 1 Rev.11\nTarget: %s-%s\nCompiler: Go\nExecution: Interpreter\n", Version, runtime.GOOS, runtime.GOARCH)
		return 0
	case "new":
		if len(args) != 2 {
			return usage("usage: craft new <directory>")
		}
		if e := project.Init(resolve(cwd, args[1]), true); e != nil {
			return projectError(e)
		}
		fmt.Fprintf(out, "Created %s\nNext: cd %s, then craft run\n", args[1], args[1])
		return 0
	case "install":
		o := project.InstallOptions{}
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--offline":
				o.Offline = true
			case "--refresh":
				o.Refresh = true
			case "--recover":
				o.Recover = true
			case "--alias":
				i++
				if i >= len(args) || o.Alias != "" {
					return usage("--alias requires one identifier")
				}
				o.Alias = args[i]
			default:
				if strings.HasPrefix(args[i], "-") || o.Spec != "" {
					return usage("usage: craft install [github.com/owner/repository@v1.2.3] [--alias name] [--offline] [--refresh]")
				}
				o.Spec = args[i]
			}
		}
		messages, e := project.Install(ctx, cwd, o)
		if e != nil {
			return projectError(e)
		}
		for _, m := range messages {
			fmt.Fprintln(out, m)
		}
		return 0
	case "init":
		if len(args) != 1 {
			return usage("usage: craft init")
		}
		if e := project.Init(cwd, false); e != nil {
			return projectError(e)
		}
		fmt.Fprintln(out, "Initialized Craft project. Run craft run to get started.")
		return 0
	case "clean":
		if len(args) != 1 {
			return usage("usage: craft clean")
		}
		if e := project.Clean(cwd); e != nil {
			return projectError(e)
		}
		fmt.Fprintln(out, "Cleaned generated Craft bundle.")
		return 0
	case "fmt":
		check := false
		file := ""
		for _, a := range args[1:] {
			if a == "--check" && !check {
				check = true
			} else if strings.HasPrefix(a, "-") || file != "" {
				return usage("usage: craft fmt [--check] [file.craft]")
			} else {
				file = a
			}
		}
		var entries []project.Source
		root := cwd
		if file != "" {
			path := resolve(cwd, file)
			info, e := os.Lstat(path)
			if e != nil {
				if file == "check" && os.IsNotExist(e) {
					return projectError(fmt.Errorf("%w; hint: use craft fmt --check to check formatting", e))
				}
				return projectError(e)
			}
			if !info.Mode().IsRegular() || filepath.Ext(path) != ".craft" || info.Size() > project.MaxSourceSize {
				return projectError(fmt.Errorf("fmt requires a regular .craft file up to 2 MiB"))
			}
			b, e := os.ReadFile(path)
			if e != nil {
				return projectError(e)
			}
			entries = []project.Source{{Path: path, Text: string(b)}}
		} else {
			p, e := project.LoadTests(cwd)
			if e != nil {
				return projectError(e)
			}
			entries = p.Sources
			root = p.Root
		}
		type change struct{ path, text string }
		changes := []change{}
		for _, s := range entries {
			sources[s.Path] = s.Text
			formatted, e := formatter.Format(s.Path, s.Text)
			if e != nil {
				return fail(e)
			}
			if formatted != s.Text {
				changes = append(changes, change{resolve(root, s.Path), formatted})
			}
		}
		// Parse and format every file before writing any of them.
		if check {
			for _, c := range changes {
				fmt.Fprintf(out, "Needs formatting: %s\n", c.path)
			}
			if len(changes) > 0 {
				return 1
			}
			fmt.Fprintln(out, "Formatting is up to date.")
			return 0
		}
		for _, c := range changes {
			if e := replaceFile(c.path, []byte(c.text)); e != nil {
				return projectError(e)
			}
			fmt.Fprintf(out, "Formatted %s\n", c.path)
		}
		fmt.Fprintf(out, "Formatted %d file(s).\n", len(changes))
		return 0
	case "package":
		if len(args) != 2 || (args[1] != "check" && args[1] != "list" && args[1] != "lock") {
			return usage("usage: craft package <check|list|lock>")
		}
		if args[1] == "lock" {
			path, count, e := project.WritePackageLock(cwd)
			if e != nil {
				return projectError(e)
			}
			fmt.Fprintf(out, "Locked %d package(s) in %s\n", count, path)
			return 0
		}
		p, locked, e := project.LoadPackageGraph(cwd)
		if e != nil {
			return projectError(e)
		}
		if args[1] == "list" {
			for _, line := range project.PackageLines(p) {
				fmt.Fprintln(out, line)
			}
			return 0
		}
		_, loaded, e := p.CompileMode(p.Manifest.Entry != "library")
		sources = loaded
		if e != nil {
			return fail(e)
		}
		for _, message := range project.PackageCheckMessages(p, locked) {
			fmt.Fprintln(out, message)
		}
		return 0
	case "run", "check", "build", "test":
		programArgs := []string{}
		if command == "run" {
			for i, a := range args[1:] {
				if a == "--" {
					programArgs = append(programArgs, args[i+2:]...)
					args = args[:i+1]
					break
				}
			}
		}
		if (command == "check" || command == "test") && len(args) != 1 {
			return usage("usage: craft " + command)
		}
		if command == "run" && (len(args) > 2 || (len(args) == 2 && strings.HasPrefix(args[1], "-"))) {
			return usage("usage: craft run [file.craftbundle] [-- args...]")
		}
		if command == "build" && (len(args) > 2 || (len(args) == 2 && args[1] != "--release")) {
			return usage("usage: craft build [--release]")
		}
		var p *project.Project
		var e error
		if command == "run" && len(args) == 2 {
			p, e = project.LoadBundle(resolve(cwd, args[1]))
		} else if command == "test" || command == "check" {
			p, e = project.LoadTests(cwd)
		} else {
			p, e = project.Load(cwd)
		}
		if e != nil {
			return projectError(e)
		}
		if command == "run" && p.Manifest.Entry == "library" {
			return usage("cannot run a library; import it from an application")
		}
		program, loaded, e := p.CompileMode(command != "test" && p.Manifest.Entry != "library")
		sources = loaded
		if e != nil {
			return fail(e)
		}
		switch command {
		case "check":
			fmt.Fprintln(out, "Check succeeded.")
		case "run":
			if e = interpreter.RunWithWriters(ctx, program, in, out, errOut, programArgs); e != nil {
				return fail(e)
			}
		case "build":
			path, e := p.Build()
			if e != nil {
				return projectError(e)
			}
			fmt.Fprintf(out, "Built %s\nSource bundle for Craft %s: craft run \"%s\"\n", path, Version, path)
			if len(args) == 2 {
				fmt.Fprintln(out, "--release uses the same source bundle; native compilation is not yet available.")
			}
		case "test":
			passed, failed := 0, 0
			for _, test := range program.Tests {
				if e = ctx.Err(); e != nil {
					return fail(e)
				}
				e = interpreter.RunTestWithWriters(ctx, program, test, out, errOut)
				if e != nil {
					failed++
					fmt.Fprintf(out, "FAIL %s\n", test.Name)
					fmt.Fprintln(errOut, diagnostics.Format(e, sources))
				} else {
					passed++
					fmt.Fprintf(out, "PASS %s\n", test.Name)
				}
			}
			fmt.Fprintf(out, "%d passed, %d failed, %d total\n", passed, failed, len(program.Tests))
			if failed > 0 {
				return 1
			}
		}
		return 0
	default:
		return usage(fmt.Sprintf("unknown command %q; use craft help", command))
	}
}
func replaceFile(path string, data []byte) error {
	info, e := os.Lstat(path)
	if e != nil {
		return e
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refusing to replace non-regular file: %s", path)
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".craft-fmt-*")
	if e != nil {
		return e
	}
	temp := f.Name()
	defer os.Remove(temp)
	if e = f.Chmod(info.Mode().Perm()); e != nil {
		f.Close()
		return e
	}
	if _, e = f.Write(data); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return os.Rename(temp, path)
}
func Main(ctx context.Context) int {
	cwd, e := os.Getwd()
	if e != nil {
		fmt.Fprintln(os.Stderr, "E4002", e)
		return 4
	}
	return Run(ctx, os.Args[1:], cwd, os.Stdout, os.Stderr)
}
