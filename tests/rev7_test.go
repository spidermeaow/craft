package tests

import (
	"context"
	"craft/internal/interpreter"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// rev7Trace keeps test evidence structured without recording a DSN, SQL text,
// payload or other secret-bearing configuration.
type rev7Trace struct {
	driver, mode, event, body, pool string
	status                          int
	elapsed                         time.Duration
}

func (trace rev7Trace) log(t *testing.T) {
	t.Helper()
	t.Logf("REV7 trace driver=%s mode=%s event=%s status=%d elapsed=%s body=%q pool=%q", trace.driver, trace.mode, trace.event, trace.status, trace.elapsed, trace.body, trace.pool)
}

const rev7DeadlineProgram = `
struct State {
    let db: DbConnection
    let driver: String
}
func reply(status: Int, text: String): HttpResponse { return std.http.response(status, std.bytes.fromString(text)) }
func slow(driver: String): String {
    if driver == "postgres" { return "SELECT pg_sleep(10)" }
    if driver == "sqlserver" { return "WAITFOR DELAY '00:00:10'; SELECT 1" }
    return "SELECT SLEEP(10)"
}
func dispatch(request: HttpRequest, state: State): HttpResponse {
    if request.path == "/health" { return reply(200, "ok") }
    if request.path == "/stats" {
        let s: DbPoolStats = std.db.stats(state.db)
        return reply(200, "open=" + std.convert.toString(s.open) + " inUse=" + std.convert.toString(s.inUse) + " idle=" + std.convert.toString(s.idle))
    }
    if request.path == "/ping" { std.db.ping(state.db); return reply(200, "ready") }
    if request.path == "/slow" {
        try { std.db.query(state.db, slow(state.driver), []); return reply(200, "unexpected completion") }
        catch error {
            print("REV7 TRACE db-error", error.type)
            if error.type == "DbTimeoutError" { return reply(504, "database timeout") }
            return reply(500, "database operation failed")
        }
    }
    return reply(404, "missing")
}
func main() {
    let driver: String = std.env.get("CRAFT_DB_DRIVER")
    let mode: String = std.env.get("CRAFT_REV7_TIMEOUT_MODE")
    var dbConfig: DbConfig = std.db.config()
    dbConfig.maxOpen = 2
    dbConfig.maxIdle = 2
    dbConfig.timeoutMs = 5000
    if mode == "db" { dbConfig.timeoutMs = 500 }
    let db: DbConnection = std.db.connect(driver, std.env.get("CRAFT_DB_DSN"), dbConfig)
    defer { std.db.close(db) }
    var http: HttpConfig = std.http.config("127.0.0.1:0")
    http.readTimeoutMs = 5000
    http.writeTimeoutMs = 5000
    http.requestTimeoutMs = 5000
    if mode == "request" { http.requestTimeoutMs = 1500 }
    std.http.serve(http, dispatch, State(db: db, driver: driver))
}
`

// This is a Craft integration test, not a direct Go-driver probe. It keeps
// read/write/client budgets longer than the one deadline selected by mode.
func TestRev7CraftDeadlineIsolation(t *testing.T) {
	for _, driver := range []string{"mysql", "postgres", "sqlserver"} {
		key := "CRAFT_" + strings.ToUpper(driver) + "_DSN"
		if os.Getenv(key) == "" {
			t.Run(driver, func(t *testing.T) { t.Skip("dedicated database credentials not loaded") })
			continue
		}
		for _, mode := range []string{"db", "request"} {
			t.Run(driver+"/"+mode, func(t *testing.T) {
				t.Setenv("CRAFT_DB_DRIVER", driver)
				t.Setenv("CRAFT_DB_DSN", os.Getenv(key))
				t.Setenv("CRAFT_REV7_TIMEOUT_MODE", mode)
				program := checked(t, rev7DeadlineProgram)
				ctx, cancel := context.WithCancel(context.Background())
				done := make(chan error, 1)
				out := &serverOutput{address: make(chan string, 1)}
				go func() { done <- interpreter.Run(ctx, program, out) }()
				var address string
				select {
				case address = <-out.address:
				case e := <-done:
					cancel()
					t.Fatalf("startup: %v", e)
				case <-time.After(15 * time.Second):
					cancel()
					t.Fatal("startup timeout")
				}
				t.Cleanup(func() {
					cancel()
					select {
					case <-done:
					case <-time.After(10 * time.Second):
						t.Error("server did not clean up")
					}
				})
				client := &http.Client{Timeout: 5 * time.Second}
				request := func(path string) (int, string) {
					t.Helper()
					response, e := client.Get("http://" + address + path)
					if e != nil {
						t.Fatal(e)
					}
					defer response.Body.Close()
					body, e := io.ReadAll(response.Body)
					if e != nil {
						t.Fatal(e)
					}
					return response.StatusCode, string(body)
				}
				if status, body := request("/health"); status != 200 || body != "ok" {
					t.Fatalf("health status=%d body=%q", status, body)
				}
				started := time.Now()
				status, body := request("/slow")
				elapsed := time.Since(started)
				trace := rev7Trace{driver: driver, mode: mode, event: "slow", status: status, body: body, elapsed: elapsed}
				trace.log(t)
				if status != 504 {
					t.Fatalf("mode=%s slow status=%d body=%q elapsed=%s", mode, status, body, elapsed)
				}
				minimum := 400 * time.Millisecond
				maximum := 3 * time.Second
				if mode == "request" {
					minimum = time.Second
				}
				if elapsed < minimum || elapsed > maximum {
					t.Fatalf("mode=%s deadline elapsed=%s", mode, elapsed)
				}
				if mode == "db" && body != "database timeout" {
					t.Fatalf("database timeout body=%q", body)
				}
				deadline := time.Now().Add(5 * time.Second)
				for {
					status, body = request("/stats")
					trace = rev7Trace{driver: driver, mode: mode, event: "pool", status: status, body: body}
					trace.log(t)
					if status == 200 && strings.Contains(body, "inUse=0") {
						break
					}
					if time.Now().After(deadline) {
						t.Fatalf("pool did not quiesce: %s", body)
					}
					time.Sleep(10 * time.Millisecond)
				}
				if status, body = request("/ping"); status != 200 || body != "ready" {
					t.Fatalf("subsequent ping status=%d body=%q", status, body)
				}
				rev7Trace{driver: driver, mode: mode, event: "subsequent-ping", status: status, body: body}.log(t)
			})
		}
	}
}

func TestRev7ProgramTimeoutContractCompiles(t *testing.T) {
	// Keep an explicit checker regression even when database credentials are absent.
	program := checked(t, rev7DeadlineProgram)
	if program == nil {
		t.Fatal("deadline fixture did not compile")
	}
}
