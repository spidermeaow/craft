package stdlib

import (
	"context"
	rt "craft/internal/runtime"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// A held physical connection is a deterministic barrier: WaitCount confirms
// that HTTP reached the pool wait before disconnect or shutdown is triggered.
func TestRev7HTTPPoolWaitCancellation(t *testing.T) {
	for _, driver := range []string{"mysql", "postgres", "sqlserver"} {
		for _, mode := range []string{"request-timeout", "db-timeout", "disconnect", "shutdown"} {
			t.Run(driver+"/"+mode, func(t *testing.T) {
				dsn := os.Getenv("CRAFT_" + strings.ToUpper(driver) + "_DSN")
				if dsn == "" {
					t.Skip("test DSN not loaded")
				}
				owner, finish := BeginDBScope(context.Background())
				defer finish()
				config, _ := invokeDB(owner, "std.db.config", nil)
				fields := config.(*rt.Struct).Fields
				fields["maxOpen"] = int64(1)
				fields["maxIdle"] = int64(1)
				fields["timeoutMs"] = int64(8000)
				if mode == "db-timeout" {
					fields["timeoutMs"] = int64(1500)
				}
				value, e := invokeDB(owner, "std.db.connect", []any{driver, dsn, config})
				if e != nil {
					t.Fatal(e)
				}
				db := value.(*DBConnection)
				holdCtx, release := context.WithTimeout(owner, 5*time.Second)
				defer release()
				held, e := db.db.Conn(holdCtx)
				if e != nil {
					t.Fatal("could not acquire pool barrier")
				}
				defer held.Close()
				ln, e := net.Listen("tcp", "127.0.0.1:0")
				if e != nil {
					t.Fatal(e)
				}
				address := ln.Addr().String()
				serverCtx, stopServer := context.WithCancel(context.Background())
				defer stopServer()
				// Keep read/write/client budgets longer than the deadline under test.
				// A GET has no slow body; request and DB timeout are isolated here.
				s := httpSettings{concurrent: 1, connections: 4, body: 128, headers: 4096, read: 10 * time.Second, write: 10 * time.Second, idle: time.Second, request: 8 * time.Second, shutdown: 50 * time.Millisecond}
				if mode == "request-timeout" {
					s.request = 1500 * time.Millisecond
				}
				done := make(chan error, 1)
				operations := make(chan error, 2)
				go func() {
					done <- serveListener(serverCtx, s, ln, func(ctx context.Context, _ *rt.Struct) (*rt.Struct, error) {
						ctx, cleanup := BeginDBScope(ctx)
						defer cleanup()
						_, e := invokeDB(ctx, "std.db.query", []any{db, "SELECT 1", &rt.Array{}})
						operations <- e
						if e != nil {
							return nil, e
						}
						return nativeStruct("HttpResponse", int64(200), rt.NewMap(), Bytes{"ok"}), nil
					}, io.Discard)
				}()
				consumed := false
				defer func() {
					stopServer()
					if !consumed {
						select {
						case <-done:
						case <-time.After(12 * time.Second):
							t.Error("HTTP server did not finish")
						}
					}
				}()
				transport := &http.Transport{}
				defer transport.CloseIdleConnections()
				client := &http.Client{Transport: transport, Timeout: 12 * time.Second}
				requestCtx, stopRequest := context.WithCancel(context.Background())
				defer stopRequest()
				req, _ := http.NewRequestWithContext(requestCtx, "GET", "http://"+address, nil)
				result := make(chan int, 1)
				go func() {
					r, e := client.Do(req)
					if e != nil {
						result <- 0
						return
					}
					defer r.Body.Close()
					io.Copy(io.Discard, r.Body)
					result <- r.StatusCode
				}()
				deadline := time.Now().Add(time.Second)
				for db.db.Stats().WaitCount == 0 {
					if time.Now().After(deadline) {
						t.Fatal("request did not reach pool wait")
					}
					time.Sleep(time.Millisecond)
				}
				// Admission rejection must not dispatch a second DB operation.
				busy, e := client.Get("http://" + address)
				if e != nil {
					t.Fatal(e)
				}
				io.Copy(io.Discard, busy.Body)
				busy.Body.Close()
				if busy.StatusCode != 503 || db.db.Stats().WaitCount != 1 {
					t.Fatal("overload reached DB or wrong HTTP status")
				}
				if mode == "disconnect" {
					stopRequest()
				}
				if mode == "shutdown" {
					stopServer()
				}
				select {
				case e := <-operations:
					var ex *rt.Exception
					if !errors.As(e, &ex) || (ex.Kind != "DbTimeoutError" && ex.Kind != "DbCancelledError") {
						t.Fatalf("cancellation failed: %v", e)
					}
				case <-time.After(10 * time.Second):
					t.Fatal("pool wait did not unwind")
				}
				select {
				case status := <-result:
					if mode == "request-timeout" && status != 504 {
						t.Fatalf("request timeout status=%d", status)
					}
					// DB-to-HTTP status mapping is application policy; raw transport returns 500.
					if mode == "db-timeout" && status != 500 {
						t.Fatalf("DB timeout status=%d", status)
					}
				case <-time.After(12 * time.Second):
					t.Fatal("client did not finish")
				}
				held.Close()
				if mode != "shutdown" {
					r, e := client.Get("http://" + address)
					if e != nil {
						t.Fatal(e)
					}
					io.Copy(io.Discard, r.Body)
					r.Body.Close()
					if r.StatusCode != 200 {
						t.Fatal("subsequent query failed")
					}
					<-operations
				}
				stopServer()
				select {
				case e := <-done:
					consumed = true
					if mode == "shutdown" && (e == nil || !strings.Contains(e.Error(), "drain deadline exceeded")) {
						t.Fatalf("shutdown deadline not reported: %v", e)
					}
				case <-time.After(12 * time.Second):
					t.Fatal("shutdown did not finish")
				}
				if db.db.Stats().InUse != 0 {
					t.Fatal("request leaked a connection")
				}
				finish()
				if db.db.Stats().OpenConnections != 0 {
					t.Fatal("owner did not close pool")
				}
				rebound, e := net.Listen("tcp", address)
				if e != nil {
					t.Fatal("port not released")
				}
				rebound.Close()
			})
		}
	}
}
