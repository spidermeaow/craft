package stdlib

import (
	"context"
	rt "craft/internal/runtime"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testTransport(t *testing.T, change func(*httpSettings), dispatch HTTPDispatch) (string, *http.Client) {
	t.Helper()
	s := httpSettings{address: "127.0.0.1:0", concurrent: 2, connections: 8, body: 128, headers: 4096, read: time.Second, write: 2 * time.Second, idle: time.Second, request: time.Second, shutdown: 100 * time.Millisecond}
	if change != nil {
		change(&s)
	}
	ln, e := net.Listen("tcp", s.address)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serveListener(ctx, s, ln, dispatch, io.Discard) }()
	transport := &http.Transport{}
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
	t.Cleanup(func() {
		transport.CloseIdleConnections()
		cancel()
		select {
		case e := <-done:
			if e != nil && !errors.Is(e, context.Canceled) {
				t.Error(e)
			}
		case <-time.After(5 * time.Second):
			t.Error("server failed to shut down")
		}
	})
	return "http://" + ln.Addr().String(), client
}
func TestHTTPValuesLimitsAndErrors(t *testing.T) {
	var saved atomic.Pointer[Context]
	address, client := testTransport(t, nil, func(ctx context.Context, r *rt.Struct) (*rt.Struct, error) {
		path := r.Fields["path"].(string)
		switch path {
		case "/fail":
			return nil, errors.New("secret stack")
		case "/bad":
			return nativeStruct("HttpResponse", int64(200), &rt.Map{Keys: []string{"x-test"}, Values: map[string]any{"x-test": &rt.Array{Items: []any{"bad\r\nheader"}}}}, Bytes{}), nil
		case "/echo":
			if r.Fields["query"].(*rt.Map).Values["q"].(*rt.Array).Items[1] != "b" {
				return nil, errors.New("query lost duplicate")
			}
			saved.Store(r.Fields["context"].(*Context))
			return nativeStruct("HttpResponse", int64(200), multiMap(map[string][]string{"set-cookie": {"a=1", "b=2"}}, true), r.Fields["body"]), nil
		}
		return nativeStruct("HttpResponse", int64(200), rt.NewMap(), Bytes{"ok"}), nil
	})
	for _, tc := range []struct {
		path, body string
		status     int
	}{{"/fail", "", 500}, {"/bad", "", 500}, {"/?q=%xx", "", 400}, {"/a%2fb", "", 400}, {"/", "too large" + strings.Repeat("x", 128), 413}, {"/", "", 200}} {
		resp, e := client.Post(address+tc.path, "text/plain", strings.NewReader(tc.body))
		if e != nil {
			t.Fatal(e)
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != tc.status || strings.Contains(string(b), "secret stack") {
			t.Fatalf("%s: %d %q", tc.path, resp.StatusCode, b)
		}
	}
	resp, e := client.Post(address+"/echo?q=a&q=b", "application/octet-stream", strings.NewReader("\x00\xff"))
	if e != nil {
		t.Fatal(e)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(b) != "\x00\xff" || len(resp.Header.Values("Set-Cookie")) != 2 {
		t.Fatalf("body/headers lost: %q %v", b, resp.Header)
	}
	// Expiration happens as the handler returns; wait via the atomic flag.
	deadline := time.Now().Add(time.Second)
	for !saved.Load().expired.Load() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if _, e := rev4ValueMethod(saved.Load(), "reason"); e == nil {
		t.Fatal("expired context accepted")
	}
}
func TestHTTPConcurrentLimitTimeoutAndDisconnect(t *testing.T) {
	started := make(chan struct{}, 4)
	ended := make(chan error, 4)
	var active atomic.Int64
	address, client := testTransport(t, func(s *httpSettings) { s.concurrent = 1; s.request = 150 * time.Millisecond }, func(ctx context.Context, r *rt.Struct) (*rt.Struct, error) {
		active.Add(1)
		started <- struct{}{}
		<-ctx.Done()
		active.Add(-1)
		ended <- context.Cause(ctx)
		return nil, ctx.Err()
	})
	first := make(chan int, 1)
	go func() {
		r, e := client.Get(address + "/slow")
		if e != nil {
			first <- 0
			return
		}
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
		first <- r.StatusCode
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler not started")
	}
	r, e := client.Get(address + "/busy")
	if e != nil {
		t.Fatal(e)
	}
	io.Copy(io.Discard, r.Body)
	r.Body.Close()
	if r.StatusCode != 503 {
		t.Fatalf("overload %d", r.StatusCode)
	}
	select {
	case code := <-first:
		if code != 504 {
			t.Fatalf("timeout %d", code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout missing")
	}
	if e := <-ended; !errors.Is(e, errRequestTimeout) {
		t.Fatalf("reason %v", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", address+"/disconnect", nil)
	done := make(chan struct{})
	go func() {
		resp, _ := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
		close(done)
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("second handler not started")
	}
	cancel()
	<-done
	select {
	case <-ended:
	case <-time.After(time.Second):
		t.Fatal("disconnect did not cancel")
	}
	if active.Load() != 0 {
		t.Fatal("handler remains active")
	}
}

func TestHTTPShutdownCancelsActiveRequestAfterDrain(t *testing.T) {
	ln, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	defer ln.Close()
	address := ln.Addr().String()
	s := httpSettings{concurrent: 1, connections: 2, body: 128, headers: 4096, read: time.Second, write: 3 * time.Second, idle: time.Second, request: 2 * time.Second, shutdown: 20 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started, ended := make(chan struct{}), make(chan error, 1)
	done := make(chan error, 1)
	go func() {
		done <- serveListener(ctx, s, ln, func(request context.Context, _ *rt.Struct) (*rt.Struct, error) {
			close(started)
			<-request.Done()
			ended <- context.Cause(request)
			return nil, request.Err()
		}, io.Discard)
	}()
	transport := &http.Transport{}
	defer transport.CloseIdleConnections()
	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
		if r, e := client.Get("http://" + address + "/slow"); e == nil {
			r.Body.Close()
		}
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	cancel()
	select {
	case cause := <-ended:
		if !errors.Is(cause, errServerShutdown) {
			t.Fatalf("shutdown cause: %v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("drain did not cancel request")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("server did not stop")
	}
	select {
	case <-clientDone:
	case <-time.After(4 * time.Second):
		t.Fatal("client did not finish")
	}
	rebound, e := net.Listen("tcp", address)
	if e != nil {
		t.Fatalf("listener not released: %v", e)
	}
	rebound.Close()
}
