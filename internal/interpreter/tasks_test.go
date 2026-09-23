package interpreter

import (
	"bytes"
	"context"
	"craft/internal/diagnostics"
	"craft/internal/project"
	"errors"
	"strings"
	"testing"
	"time"
)

type clockRequest struct {
	delay   time.Duration
	release chan struct{}
}
type manualClock struct{ requests chan clockRequest }

func (c *manualClock) Wait(ctx context.Context, d time.Duration) error {
	r := clockRequest{d, make(chan struct{})}
	select {
	case c.requests <- r:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case <-r.release:
		return ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}
func request(t *testing.T, c *manualClock, want time.Duration) clockRequest {
	t.Helper()
	select {
	case r := <-c.requests:
		if r.delay != want {
			t.Fatalf("delay %v, want %v", r.delay, want)
		}
		return r
	case <-time.After(3 * time.Second):
		t.Fatal("scheduler did not reach wait")
		return clockRequest{}
	}
}
func awaitRun(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case e := <-done:
		return e
	case <-time.After(3 * time.Second):
		t.Fatal("run did not finish")
		return nil
	}
}
func checkedMachine(t *testing.T, ctx context.Context, source string, out *bytes.Buffer) *machine {
	t.Helper()
	p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: source}}}
	program, _, e := p.Compile()
	if e != nil {
		t.Fatal(e)
	}
	return newMachine(ctx, program, out)
}
func launch(m *machine) <-chan error {
	done := make(chan error, 1)
	go func() { f := m.functions["main"]; _, e := m.call(f, nil, f.Pos); done <- e }()
	return done
}

func TestTasksWaitIndependently(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out bytes.Buffer
	m := checkedMachine(t, ctx, `func worker(n:Int){std.time.sleep(std.time.durationMilliseconds(10));print(n)} func main(){let a:Task=std.task.spawn(worker,1);let b:Task=std.task.spawn(worker,2);a.join();b.join()}`, &out)
	c := &manualClock{make(chan clockRequest)}
	m.tasks.clock = c
	done := launch(m)
	a := request(t, c, 10*time.Millisecond)
	b := request(t, c, 10*time.Millisecond) // Both tasks reach sleep before either is released.
	close(b.release)
	close(a.release)
	if e := awaitRun(t, done); e != nil {
		t.Fatal(e)
	}
	if s := out.String(); s != "1\n2\n" && s != "2\n1\n" {
		t.Fatalf("interleaved output: %q", s)
	}
}

func TestTimerFixedDelayAndCancellationCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out bytes.Buffer
	m := checkedMachine(t, ctx, `func tick(){defer{print("clean")};std.time.sleep(std.time.durationMilliseconds(9));print("tick")} func main(){let t:Timer=std.timer.every(std.time.durationMilliseconds(5),tick);t.join()}`, &out)
	c := &manualClock{make(chan clockRequest)}
	m.tasks.clock = c
	done := launch(m)
	first := request(t, c, 5*time.Millisecond)
	close(first.release)
	callback := request(t, c, 9*time.Millisecond)
	close(callback.release)
	second := request(t, c, 5*time.Millisecond)
	close(second.release)
	request(t, c, 9*time.Millisecond)
	cancel()
	if e := awaitRun(t, done); !errors.Is(e, context.Canceled) {
		t.Fatalf("got %v", e)
	}
	if s := out.String(); s != "tick\nclean\nclean\n" {
		t.Fatalf("cleanup or callback overlap: %q", s)
	}
	if m.tasks.active != 0 {
		t.Fatal("tasks leaked")
	}
}

func TestUnobservedChildFailureAndOwnedLimit(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{`func fail(){throw Exception("lost failure")} func main(){let t:Task=std.task.spawn(fail);while t.status()!="failed"{}}`, "lost failure"},
		{`func noop(){} func main(){for i in 0..257{std.timer.after(std.time.durationMilliseconds(60000),noop)}}`, "task limit"},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		var out bytes.Buffer
		m := checkedMachine(t, ctx, tc.source, &out)
		e := awaitRun(t, launch(m))
		cancel()
		if e == nil || !strings.Contains(e.Error(), tc.want) {
			t.Fatalf("%v, want %s", e, tc.want)
		}
		if m.tasks.active != 0 {
			t.Fatal("tasks leaked")
		}
	}
}

func TestJoinRejectsCyclesAndAncestors(t *testing.T) {
	g := newTaskGroup()
	a, b := &task{group: g}, &task{group: g}
	m := &machine{ctx: context.Background(), tasks: g, self: a}
	g.waiting[b] = a
	if _, e := m.taskMethod(b, "join", diagnostics.Position{}); e == nil {
		t.Fatal("cycle accepted")
	}
	delete(g.waiting, b)
	m.self = b
	b.parent = a
	if _, e := m.taskMethod(a, "join", diagnostics.Position{}); e == nil {
		t.Fatal("ancestor join accepted")
	}
	if _, e := m.taskMethod(b, "join", diagnostics.Position{}); e == nil {
		t.Fatal("self join accepted")
	}
}

func TestZeroTimerRunsOnceAndArgumentsAreCopied(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out bytes.Buffer
	m := checkedMachine(t, ctx, `func once(a:Int[]){print(a)} func main(){var a:Int[]=[1];let t:Timer=std.timer.after(std.time.durationMilliseconds(0),once,a);a.append(2);t.join();t.join();print(a,t.status())}`, &out)
	c := &manualClock{make(chan clockRequest)}
	m.tasks.clock = c
	done := launch(m)
	r := request(t, c, 0)
	close(r.release)
	if e := awaitRun(t, done); e != nil {
		t.Fatal(e)
	}
	if s := out.String(); s != "[1]\n[1, 2] completed\n" {
		t.Fatalf("%q", s)
	}
}

func TestFrameCleanupCancelsPendingTimer(t *testing.T) {
	var out bytes.Buffer
	m := checkedMachine(t, context.Background(), `func fail(){throw Exception("should not run")} func owner():Timer{return std.timer.after(std.time.durationMilliseconds(60000),fail)} func main(){let t:Timer=owner();assert t.status()=="cancelled";try{t.join()}catch e{assert e.type=="CancellationError"}}`, &out)
	if e := awaitRun(t, launch(m)); e != nil {
		t.Fatal(e)
	}
}

func TestRepeatingTimerStopsOnFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out bytes.Buffer
	m := checkedMachine(t, ctx, `func fail(){throw Exception("timer failed")} func main(){let t:Timer=std.timer.every(std.time.durationMilliseconds(5),fail);try{t.join()}catch e{print(e.message)};assert t.status()=="failed";try{t.join()}catch e{print(e.message)}}`, &out)
	c := &manualClock{make(chan clockRequest)}
	m.tasks.clock = c
	done := launch(m)
	r := request(t, c, 5*time.Millisecond)
	close(r.release)
	if e := awaitRun(t, done); e != nil {
		t.Fatal(e)
	}
	if s := out.String(); s != "timer failed\ntimer failed\n" {
		t.Fatalf("%q", s)
	}
	if m.tasks.active != 0 {
		t.Fatal("failed timer remained active")
	}
}
