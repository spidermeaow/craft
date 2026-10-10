package stdlib

import (
	"bytes"
	"context"
	"craft/internal/ast"
	rt "craft/internal/runtime"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type Process struct {
	mu        sync.Mutex
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	done      chan struct{}
	exitCode  int64
	hasExit   bool
	output    *processOutput
	waitError error
}

func (*Process) String() string { return "Process" }

type processScopeKey struct{}
type processScope struct {
	mu        sync.Mutex
	closed    bool
	processes map[*Process]bool
}

func BeginProcessScope(ctx context.Context) (context.Context, func()) {
	scope := &processScope{processes: map[*Process]bool{}}
	return context.WithValue(ctx, processScopeKey{}, scope), func() {
		scope.mu.Lock()
		scope.closed = true
		active := make([]*Process, 0, len(scope.processes))
		for p := range scope.processes {
			active = append(active, p)
		}
		scope.mu.Unlock()
		for _, p := range active {
			select {
			case <-p.done:
			default:
				_ = p.cmd.Process.Kill()
				<-p.done
			}
		}
	}
}

type processOutput struct {
	mu       sync.Mutex
	stdout   bytes.Buffer
	stderr   bytes.Buffer
	limit    int
	used     int
	exceeded bool
}

type processWriter struct {
	owner  *processOutput
	stderr bool
}

func (w processWriter) Write(data []byte) (int, error) {
	o := w.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	remaining := o.limit - o.used
	if remaining < len(data) {
		o.exceeded = true
		if remaining < 0 {
			remaining = 0
		}
	} else {
		remaining = len(data)
	}
	if remaining > 0 {
		if w.stderr {
			_, _ = o.stderr.Write(data[:remaining])
		} else {
			_, _ = o.stdout.Write(data[:remaining])
		}
		o.used += remaining
	}
	// Continue draining both pipes even after the collection limit is reached.
	return len(data), nil
}

func init() {
	signatures["std.process.run"] = Signature{Params: []ast.Type{ast.String, ast.String.Array(), ast.Int, ast.Int}, Result: "ProcessResult"}
	signatures["std.process.start"] = Signature{Params: []ast.Type{ast.String, ast.String.Array(), ast.Int}, Result: "Process"}
	signatures["std.process.wait"] = Signature{Params: []ast.Type{"Process", ast.Int}, Result: "ProcessResult"}
	signatures["std.process.kill"] = Signature{Params: []ast.Type{"Process"}, Result: ast.Void}
	signatures["std.process.exitCode"] = Signature{Params: []ast.Type{"Process"}, Result: ast.Int.Optional()}
}

func rev43ProcessStruct() *ast.Struct {
	s := &ast.Struct{Name: "ProcessResult"}
	for _, field := range []struct {
		name string
		typ  ast.Type
	}{
		{"exitCode", ast.Int}, {"stdout", ast.Bytes}, {"stderr", ast.Bytes},
	} {
		s.Fields = append(s.Fields, ast.Field{Name: field.name, Type: field.typ})
	}
	return s
}

func processArgs(value *rt.Array) ([]string, error) {
	if len(value.Items) > 256 {
		return nil, toolingError("ProcessLimitError", "too many process arguments")
	}
	result := make([]string, len(value.Items))
	total := 0
	for i, item := range value.Items {
		part := item.(string)
		total += len(part)
		if !utf8.ValidString(part) || strings.ContainsRune(part, 0) || total > 32768 {
			return nil, toolingError("ProcessArgumentError", "invalid or oversized process arguments")
		}
		result[i] = part
	}
	return result, nil
}

func (s *Session) startProcess(ctx context.Context, executable string, arguments *rt.Array, maxOutput int64) (*Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	scope, _ := ctx.Value(processScopeKey{}).(*processScope)
	if scope == nil {
		return nil, toolingError("ProcessStartError", "process start requires an execution scope")
	}
	if executable == "" || !utf8.ValidString(executable) || strings.ContainsRune(executable, 0) || len(executable) > 32768 {
		return nil, toolingError("ProcessArgumentError", "invalid executable path")
	}
	ext := strings.ToLower(filepath.Ext(executable))
	if ext == ".bat" || ext == ".cmd" {
		return nil, toolingError("ProcessArgumentError", "batch scripts are not supported by std.process")
	}
	if maxOutput < 0 || maxOutput > MaxBodyBytes {
		return nil, toolingError("ProcessLimitError", "maxOutputBytes must be 0..16777216")
	}
	args, err := processArgs(arguments)
	if err != nil {
		return nil, err
	}
	childCtx, cancel := context.WithCancel(ctx)
	command := exec.CommandContext(childCtx, executable, args...)
	resolvedExt := strings.ToLower(filepath.Ext(command.Path))
	if resolvedExt == ".bat" || resolvedExt == ".cmd" {
		cancel()
		return nil, toolingError("ProcessArgumentError", "batch scripts are not supported by std.process")
	}
	limit := &processOutput{limit: int(maxOutput)}
	command.Stdout = processWriter{owner: limit}
	command.Stderr = processWriter{owner: limit, stderr: true}
	configureChildProcess(command)
	if err := command.Start(); err != nil {
		cancel()
		return nil, toolingError("ProcessStartError", "could not start executable")
	}
	p := &Process{cmd: command, cancel: cancel, done: make(chan struct{}), output: limit}
	scope.mu.Lock()
	if scope.closed {
		scope.mu.Unlock()
		_ = command.Process.Kill()
		_ = command.Wait()
		return nil, toolingError("ProcessStartError", "process owner has exited")
	}
	scope.processes[p] = true
	scope.mu.Unlock()
	go func() {
		waitErr := command.Wait()
		p.mu.Lock()
		p.waitError = waitErr
		if command.ProcessState != nil {
			p.exitCode = int64(command.ProcessState.ExitCode())
			p.hasExit = true
		}
		p.mu.Unlock()
		close(p.done)
		cancel()
		scope.mu.Lock()
		delete(scope.processes, p)
		scope.mu.Unlock()
	}()
	return p, nil
}

func (s *Session) processResult(ctx context.Context, p *Process, timeoutMs int64) (any, error) {
	if timeoutMs < 1 || timeoutMs > 300000 {
		return nil, toolingError("ProcessLimitError", "timeoutMs must be 1..300000")
	}
	timer := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-p.done:
	case <-timer.C:
		_ = p.cmd.Process.Kill()
		<-p.done
		return nil, toolingError("ProcessTimeoutError", "child process timed out")
	case <-ctx.Done():
		_ = p.cmd.Process.Kill()
		<-p.done
		return nil, ctx.Err()
	}
	p.output.mu.Lock()
	stdout := p.output.stdout.String()
	stderr := p.output.stderr.String()
	exceeded := p.output.exceeded
	p.output.mu.Unlock()
	if exceeded {
		return nil, toolingError("ProcessLimitError", "child output exceeded maxOutputBytes")
	}
	p.mu.Lock()
	code, ok, waitErr := p.exitCode, p.hasExit, p.waitError
	p.mu.Unlock()
	if !ok {
		return nil, toolingError("ProcessWaitError", "child process has no exit code")
	}
	var exitErr *exec.ExitError
	if waitErr != nil && !errors.As(waitErr, &exitErr) {
		return nil, toolingError("ProcessWaitError", "could not collect child process result")
	}
	return nativeStruct("ProcessResult", code, Bytes{data: stdout}, Bytes{data: stderr}), nil
}

func (s *Session) invokeProcess(ctx context.Context, name string, a []any) (any, error) {
	switch name {
	case "std.process.start":
		return s.startProcess(ctx, a[0].(string), a[1].(*rt.Array), a[2].(int64))
	case "std.process.run":
		timeout := a[2].(int64)
		if timeout < 1 || timeout > 300000 {
			return nil, toolingError("ProcessLimitError", "timeoutMs must be 1..300000")
		}
		p, err := s.startProcess(ctx, a[0].(string), a[1].(*rt.Array), a[3].(int64))
		if err != nil {
			return nil, err
		}
		return s.processResult(ctx, p, timeout)
	case "std.process.wait":
		return s.processResult(ctx, a[0].(*Process), a[1].(int64))
	case "std.process.kill":
		p := a[0].(*Process)
		select {
		case <-p.done:
			return nil, nil
		default:
		}
		if err := p.cmd.Process.Kill(); err != nil {
			select {
			case <-p.done:
				return nil, nil
			default:
				return nil, toolingError("ProcessWaitError", "could not kill child process")
			}
		}
		return nil, nil
	case "std.process.exitCode":
		p := a[0].(*Process)
		select {
		case <-p.done:
		default:
			return rt.Optional{}, nil
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.hasExit {
			return rt.Some(p.exitCode), nil
		}
		return rt.Optional{}, nil
	}
	return nil, toolingError("ProcessError", "unknown process operation")
}
