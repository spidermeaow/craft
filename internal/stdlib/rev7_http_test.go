package stdlib

import (
	"bufio"
	"context"
	rt "craft/internal/runtime"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRev7RequestDeadlineDoesNotUseReadDeadline(t *testing.T) {
	causes := make(chan error, 1)
	address, client := testTransport(t, func(s *httpSettings) {
		s.read = 2 * time.Second
		s.write = 2 * time.Second
		s.request = 40 * time.Millisecond
	}, func(ctx context.Context, _ *rt.Struct) (*rt.Struct, error) {
		<-ctx.Done()
		causes <- context.Cause(ctx)
		return nil, ctx.Err()
	})
	response, e := client.Get(address + "/request-deadline")
	if e != nil {
		t.Fatal(e)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusGatewayTimeout || !strings.Contains(string(body), "request timeout") {
		t.Fatalf("status=%d body=%q", response.StatusCode, body)
	}
	select {
	case cause := <-causes:
		if cause != errRequestTimeout {
			t.Fatalf("cause=%v", cause)
		}
	case <-time.After(time.Second):
		t.Fatal("handler did not observe request deadline")
	}
}

func TestRev7BodyDeadlineSkipsDispatch(t *testing.T) {
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	settings := httpSettings{concurrent: 1, connections: 2, body: 128, headers: 4096, read: 40 * time.Millisecond, write: time.Second, idle: time.Second, request: time.Second, shutdown: time.Second}
	var calls atomic.Int64
	done := make(chan error, 1)
	go func() {
		done <- serveListener(ctx, settings, listener, func(context.Context, *rt.Struct) (*rt.Struct, error) {
			calls.Add(1)
			return nativeStruct("HttpResponse", int64(200), rt.NewMap(), Bytes{"unexpected"}), nil
		}, io.Discard)
	}()
	connection, e := net.Dial("tcp", listener.Addr().String())
	if e != nil {
		t.Fatal(e)
	}
	defer connection.Close()
	if _, e = io.WriteString(connection, "POST /body HTTP/1.1\r\nHost: test\r\nContent-Length: 4\r\n\r\nx"); e != nil {
		t.Fatal(e)
	}
	_ = connection.SetReadDeadline(time.Now().Add(time.Second))
	response, e := http.ReadResponse(bufio.NewReader(connection), nil)
	if e != nil {
		t.Fatal(e)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusRequestTimeout {
		t.Fatalf("status=%d", response.StatusCode)
	}
	if calls.Load() != 0 {
		t.Fatal("body timeout dispatched to application")
	}
	stop()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("server did not stop")
	}
}

func TestRev7CompletedBodyDoesNotExpireHandler(t *testing.T) {
	address, client := testTransport(t, func(s *httpSettings) {
		s.read = 25 * time.Millisecond
		s.request = 500 * time.Millisecond
		s.write = time.Second
	}, func(ctx context.Context, _ *rt.Struct) (*rt.Struct, error) {
		time.Sleep(75 * time.Millisecond)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nativeStruct("HttpResponse", int64(200), rt.NewMap(), Bytes{"ok"}), nil
	})
	response, e := client.Post(address+"/complete", "text/plain", strings.NewReader("done"))
	if e != nil {
		t.Fatal(e)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("completed body request status=%d", response.StatusCode)
	}
}
