package tests

import (
	"context"
	"craft/internal/interpreter"
	"craft/internal/project"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

type serverOutput struct{ address chan string }

func (w *serverOutput) Write(p []byte) (int, error) {
	s := string(p)
	if strings.HasPrefix(s, "HTTP listening on ") {
		select {
		case w.address <- strings.TrimSpace(strings.TrimPrefix(s, "HTTP listening on ")):
		default:
		}
	}
	return len(p), nil
}

func runCraftHTTP(t *testing.T, count int) {
	t.Helper()
	t.Setenv("CRAFT_API_ADDRESS", "127.0.0.1:0")
	p, e := project.Load("../examples/api-server")
	if e != nil {
		t.Fatal(e)
	}
	program, _, e := p.Compile()
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	out := &serverOutput{make(chan string, 1)}
	go func() { done <- interpreter.Run(ctx, program, out) }()
	defer cancel()
	var address string
	select {
	case address = <-out.address:
	case e := <-done:
		t.Fatalf("server exited: %v", e)
	case <-time.After(5 * time.Second):
		t.Fatal("server startup timed out")
	}
	transport := &http.Transport{}
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
	defer transport.CloseIdleConnections()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("Craft server did not stop")
		}
		ln, e := net.Listen("tcp", address)
		if e != nil {
			t.Errorf("listener remains open: %v", e)
		} else {
			ln.Close()
		}
	})
	base := "http://" + address
	for _, tc := range []struct {
		method, path, body string
		code               int
	}{{"GET", "/health", "", 200}, {"GET", "/items/1", "", 200}, {"GET", "/items/2", "", 404}, {"POST", "/items", `{"name":"Craft"}`, 201}, {"POST", "/items", `{}`, 422}, {"POST", "/items", `{`, 400}, {"DELETE", "/health", "", 405}, {"GET", "/missing", "", 404}} {
		req, _ := http.NewRequest(tc.method, base+tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		r, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}
		b, _ := io.ReadAll(r.Body)
		r.Body.Close()
		if r.StatusCode != tc.code {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, r.StatusCode, b)
		}
	}
	warmup := 100
	var before, after runtime.MemStats
	for i := 0; i < count+warmup; i++ {
		if i == warmup {
			runtime.GC()
			runtime.ReadMemStats(&before)
		}
		r, e := client.Get(base + "/health")
		if e != nil {
			t.Fatal(e)
		}
		_, e = io.Copy(io.Discard, r.Body)
		r.Body.Close()
		if e != nil || r.StatusCode != 200 {
			t.Fatalf("request %d: %v status %d", i, e, r.StatusCode)
		}
	}
	runtime.GC()
	runtime.ReadMemStats(&after)
	t.Logf("%s/%s Go %s; requests=%d; heap before=%d after=%d; goroutines=%d", runtime.GOOS, runtime.GOARCH, runtime.Version(), count, before.HeapAlloc, after.HeapAlloc, runtime.NumGoroutine())
	// Deliberately report measured memory rather than inventing a platform-specific
	// acceptance threshold. See TRY-REV4 for repeated runs and process measurements.
	if count >= 10000 {
		t.Log(fmt.Sprintf("Completed %d real HTTP requests through the Craft framework", count))
	}
}
func TestRev4HTTPApplication(t *testing.T) { runCraftHTTP(t, 300) }
func TestRev4HTTPStress(t *testing.T) {
	if os.Getenv("CRAFT_STRESS") != "1" {
		t.Skip("set CRAFT_STRESS=1 for 10,000 HTTP requests")
	}
	runCraftHTTP(t, 10000)
}

func TestRev4HTTPRequestStateAndTaskScope(t *testing.T) {
	program := checked(t, `struct State{var count:Int}
func work(){}
func dispatch(request:HttpRequest,state:State):HttpResponse{
 var local:State=state
 local.count+=1
 let t:Task=std.task.spawn(work)
 t.join()
 return std.http.response(200,std.bytes.fromString(std.convert.toString(local.count)))
}
func main(){std.http.serve(std.http.config("127.0.0.1:0"),dispatch,State(count:0))}`)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	out := &serverOutput{make(chan string, 1)}
	go func() { done <- interpreter.Run(ctx, program, out) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("request scopes failed to drain")
		}
	})
	var address string
	select {
	case address = <-out.address:
	case e := <-done:
		t.Fatalf("startup: %v", e)
	case <-time.After(5 * time.Second):
		t.Fatal("startup timeout")
	}
	transport := &http.Transport{}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
	for i := 0; i < 300; i++ {
		r, e := client.Get("http://" + address + "/")
		if e != nil {
			t.Fatal(e)
		}
		body, e := io.ReadAll(r.Body)
		r.Body.Close()
		if e != nil || r.StatusCode != 200 || string(body) != "1" {
			t.Fatalf("request %d: %d %q %v", i, r.StatusCode, body, e)
		}
	}
}
