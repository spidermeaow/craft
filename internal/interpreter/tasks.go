package interpreter

import (
	"context"
	"craft/internal/ast"
	"craft/internal/diagnostics"
	rt "craft/internal/runtime"
	"craft/internal/stdlib"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
)

const taskLimit = 256

type clock interface {
	Wait(context.Context, time.Duration) error
}
type realClock struct{}

func (realClock) Wait(ctx context.Context, d time.Duration) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if d == 0 {
		runtime.Gosched()
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return ctx.Err()
	}
}

type lockedWriter struct {
	mu  sync.Mutex
	out io.Writer
}

func (w *lockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.out.Write(p)
}

// All handle state and wait edges are protected by the group's mutex. Each task
// gets a separate machine; checked ASTs, functions and structs are read-only.
type taskGroup struct {
	mu      sync.Mutex
	active  int
	waiting map[*task]*task
	clock   clock
}

func newTaskGroup() *taskGroup { return &taskGroup{waiting: map[*task]*task{}, clock: realClock{}} }
func (g *taskGroup) yield()    { runtime.Gosched() }

type task struct {
	group    *taskGroup
	parent   *task
	cancel   context.CancelFunc
	done     chan struct{}
	status   string
	err      error
	observed bool
	timer    bool
}

func (t *task) String() string {
	t.group.mu.Lock()
	defer t.group.mu.Unlock()
	name := "Task"
	if t.timer {
		name = "Timer"
	}
	return name + "(" + t.status + ")"
}

func (m *machine) startTask(x *ast.Expr, env *environment, name string) (any, error) {
	if m.cleaning {
		return nil, fmt.Errorf("cannot start tasks or timers during defer cleanup")
	}
	offset := 0
	var delay time.Duration
	if name != "std.task.spawn" {
		offset = 1
		v, e := m.expression(x.Args[0], env)
		if e != nil {
			return nil, e
		}
		delay = v.(stdlib.Duration).GoDuration()
		if delay < 0 || (name == "std.timer.every" && delay == 0) {
			return nil, fmt.Errorf("timer delay must be nonnegative; every interval must be positive")
		}
	}
	args := make([]any, len(x.Args)-offset-1)
	ref, e := m.expression(x.Args[offset], env)
	if e != nil {
		return nil, e
	}
	for i, a := range x.Args[offset+1:] {
		v, e := m.expression(a, env)
		if e != nil {
			return nil, e
		}
		args[i] = rt.Clone(v)
	}
	f := ref.(rt.Function).Declaration
	fr := m.frames[len(m.frames)-1]
	g := m.tasks
	g.mu.Lock()
	if g.active >= taskLimit || len(fr.children) >= taskLimit {
		// Reap terminal successful/observed tasks before enforcing the owner limit.
		live := fr.children[:0]
		for _, child := range fr.children {
			select {
			case <-child.done:
				if child.status == "failed" && !child.observed {
					live = append(live, child)
				}
			default:
				live = append(live, child)
			}
		}
		for i := len(live); i < len(fr.children); i++ {
			fr.children[i] = nil
		}
		fr.children = live
	}
	if g.active >= taskLimit || len(fr.children) >= taskLimit {
		g.mu.Unlock()
		return nil, fmt.Errorf("task limit is %d active per run or unobserved/active children per function call; join failed children to observe errors", taskLimit)
	}
	ctx, cancel := context.WithCancel(m.ctx)
	t := &task{group: g, parent: m.self, cancel: cancel, done: make(chan struct{}), status: "pending", timer: offset == 1}
	g.active++
	fr.children = append(fr.children, t)
	g.mu.Unlock()
	child := &machine{ctx: ctx, out: m.out, tasks: g, self: t, functions: m.functions, structs: m.structs, library: m.library, servingRequest: m.servingRequest}
	go func() {
		var err error
		defer func() {
			cancel()
			g.mu.Lock()
			t.err = err
			t.status = "completed"
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				t.status = "cancelled"
			} else if err != nil {
				t.status = "failed"
			}
			g.active--
			close(t.done)
			g.mu.Unlock()
		}()
		for {
			if t.timer {
				err = g.clock.Wait(ctx, delay)
				if err != nil {
					return
				}
			}
			// cancel and callback dispatch have one serialized decision point.
			g.mu.Lock()
			err = ctx.Err()
			if err == nil {
				t.status = "running"
			}
			g.mu.Unlock()
			if err != nil {
				return
			}
			_, err = child.call(f, args, x.Pos)
			if err == nil {
				err = ctx.Err()
			}
			if err != nil || name != "std.timer.every" {
				return
			}
			g.mu.Lock()
			t.status = "pending"
			g.mu.Unlock()
		}
	}()
	return t, nil
}

func (m *machine) taskMethod(t *task, name string, pos diagnostics.Position) (any, error) {
	g := m.tasks
	g.mu.Lock()
	switch name {
	case "status":
		s := t.status
		g.mu.Unlock()
		return s, nil
	case "cancel":
		t.cancel()
		g.mu.Unlock()
		return nil, nil
	}
	// An ancestor implicitly waits for descendants during structured cleanup.
	for p := m.self; p != nil; p = p.parent {
		if p == t {
			g.mu.Unlock()
			return nil, fmt.Errorf("cannot join self or an ancestor task")
		}
	}
	if m.self != nil {
		for p := t; p != nil; p = g.waiting[p] {
			if p == m.self {
				g.mu.Unlock()
				return nil, fmt.Errorf("task join cycle detected")
			}
		}
		g.waiting[m.self] = t
	}
	g.mu.Unlock()
	defer func() { g.mu.Lock(); delete(g.waiting, m.self); g.mu.Unlock() }()
	select {
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	case <-t.done:
	}
	g.mu.Lock()
	t.observed = true
	err, status := t.err, t.status
	g.mu.Unlock()
	if status == "cancelled" {
		return nil, m.exception("CancellationError", "task was cancelled", "E3004", pos)
	}
	return nil, err
}

func (m *machine) finishChildren(fr *frame) error {
	g := m.tasks
	g.mu.Lock()
	for _, t := range fr.children {
		t.cancel()
	}
	g.mu.Unlock()
	var first error
	for _, t := range fr.children {
		<-t.done
		g.mu.Lock()
		if first == nil && t.status == "failed" && !t.observed {
			first = t.err
		}
		g.mu.Unlock()
	}
	return first
}
