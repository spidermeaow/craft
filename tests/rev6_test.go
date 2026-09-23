package tests

import (
	"context"
	"craft/internal/interpreter"
	"craft/internal/project"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Extra endpoints are injected only into this compiled fixture, never shipped
// in the example application. Markers establish uncommitted-write barriers.
const rev6FixtureRoutes = `
    if request.path == "/_child" {
        let child: Task = std.task.spawn(rev6Child, state)
        child.join()
        return reply(200, "child finished")
    }
    if request.path == "/_slow" || request.path == "/_dbslow" || request.path == "/_inflight" {
        var sql: String = "SELECT SLEEP(10)"
        if state.driver == "postgres" { sql = "SELECT pg_sleep(10)" }
        if state.driver == "sqlserver" { sql = "WAITFOR DELAY '00:00:10'; SELECT 1" }
        sql = sql + " /* " + state.table + " */"
        if request.path == "/_dbslow" {
            var config: DbConfig = std.db.config()
            config.timeoutMs = 500
            let local: DbConnection = std.db.connect(state.driver, std.env.get("CRAFT_DB_DSN"), config)
            std.db.query(local, sql, [])
        } else { std.db.query(state.db, sql, []) }
        return reply(200, "unexpected completion")
    }
    if request.path == "/_stats" {
        let stats: DbPoolStats = std.db.stats(state.db)
        assert stats.open <= 2
        return reply(200, std.convert.toString(stats.inUse))
    }
    if request.path == "/_stop" { std.http.stop(request.context); return reply(200, "stopping") }
    if request.path == "/_rollback" || request.path == "/_cancel" || request.path == "/_commit" {
        let tx: DbTransaction = std.db.begin(state.db)
        std.db.executeTx(tx, "INSERT INTO " + state.table + " (id, value) VALUES (" + placeholder(state.driver, 1) + ", " + placeholder(state.driver, 2) + ")", [std.db.text("barrier"), std.db.text("private")])
        if request.path == "/_commit" { std.db.commit(tx) }
        print("REV6 BARRIER")
        if request.path == "/_rollback" { throw Exception("forced fixture failure") }
        while true { std.db.query(state.db, "SELECT 1", []) }
    }
`

const rev6ChildSource = `
func rev6Child(state: ServerState) {
    let tx: DbTransaction = std.db.begin(state.db)
    std.db.executeTx(tx, "INSERT INTO " + state.table + " (id, value) VALUES (" + placeholder(state.driver, 1) + ", " + placeholder(state.driver, 2) + ")", [std.db.text("child"), std.db.text("uncommitted")])
    print("REV6 BARRIER")
    while true { std.db.query(state.db, "SELECT 1", []) }
}
`

type rev6Output struct {
	serverOutput
	barrier chan struct{}
}

func (w *rev6Output) Write(p []byte) (int, error) {
	w.serverOutput.Write(p)
	if strings.Contains(string(p), "REV6 BARRIER") {
		select {
		case w.barrier <- struct{}{}:
		default:
		}
	}
	return len(p), nil
}

// Rev.7 keeps this end-to-end fixture separate from deadline-isolation tests.
// It uses an actual Craft program and an independent observer connection.
func TestRev7CraftDatabaseLifecycle(t *testing.T) {
	for _, driver := range []string{"mysql", "postgres", "sqlserver"} {
		t.Run(driver, func(t *testing.T) {
			dsn := os.Getenv("CRAFT_" + strings.ToUpper(driver) + "_DSN")
			if dsn == "" {
				t.Skip("dedicated test DSN not loaded")
			}
			id := fmt.Sprintf("r%d", time.Now().UnixNano())
			table := "craft_rev6_" + id
			t.Setenv("CRAFT_DB_DRIVER", driver)
			t.Setenv("CRAFT_DB_DSN", dsn)
			t.Setenv("CRAFT_DB_RUN_ID", id)
			t.Setenv("CRAFT_HTTP_ADDRESS", "127.0.0.1:0")
			p, e := project.Load("../examples/rev6-http-db")
			if e != nil {
				t.Fatal(e)
			}
			for i := range p.Sources {
				p.Sources[i].Text = strings.Replace(p.Sources[i].Text, "func route(request: HttpRequest, state: ServerState): HttpResponse {", "func route(request: HttpRequest, state: ServerState): HttpResponse {"+rev6FixtureRoutes, 1)
				if strings.Contains(p.Sources[i].Text, "func main()") {
					p.Sources[i].Text += rev6ChildSource
				}
			}
			program, _, e := p.Compile()
			if e != nil {
				t.Fatal(e)
			}
			sqlDriver := driver
			if driver == "postgres" {
				sqlDriver = "pgx"
			}
			observer, e := sql.Open(sqlDriver, dsn)
			if e != nil {
				t.Fatal("observer configuration rejected")
			}
			t.Cleanup(func() { observer.Close() })
			observer.SetMaxOpenConns(1)
			identity, version := "SELECT DATABASE()", "SELECT VERSION()"
			if driver == "postgres" {
				identity = "SELECT current_database()"
			}
			if driver == "sqlserver" {
				identity = "SELECT DB_NAME()"
				version = "SELECT @@VERSION"
			}
			queryText := func(query string) string {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var value string
				if e := observer.QueryRowContext(ctx, query).Scan(&value); e != nil {
					t.Fatal("independent database read failed")
				}
				return value
			}
			if queryText(identity) != "craft_test" {
				t.Fatal("refusing non-test database")
			}
			t.Logf("driver=%s server=%s runtime=%s/%s %s fixture=%s", driver, queryText(version), runtime.GOOS, runtime.GOARCH, runtime.Version(), table)
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			out := &rev6Output{serverOutput{make(chan string, 1)}, make(chan struct{}, 1)}
			go func() { done <- interpreter.Run(ctx, program, out) }()
			var address string
			stopped := false
			t.Cleanup(func() {
				cancel()
				if !stopped {
					select {
					case <-done:
					case <-time.After(15 * time.Second):
						t.Errorf("server cleanup timed out; inspect fixture %s", table)
						return
					}
				}
				// Never drop an unknown fixture: the Craft owner registered its own DROP.
				if address != "" {
					ln, e := net.Listen("tcp", address)
					if e != nil {
						t.Error("listener not released")
					} else {
						ln.Close()
					}
				}
				cleanCtx, release := context.WithTimeout(context.Background(), 5*time.Second)
				defer release()
				var count int
				e := observer.QueryRowContext(cleanCtx, "SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = '"+table+"'").Scan(&count)
				if e != nil || count != 0 {
					t.Errorf("fixture cleanup unconfirmed: %s", table)
				}
			})
			select {
			case address = <-out.address:
			case e := <-done:
				stopped = true
				t.Fatalf("startup: %v", e)
			case <-time.After(15 * time.Second):
				t.Fatal("startup timeout")
			}
			transport := &http.Transport{}
			defer transport.CloseIdleConnections()
			client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
			base := "http://" + address
			request := func(method, path, body string, want int) (string, error) {
				req, e := http.NewRequest(method, base+path, strings.NewReader(body))
				if e != nil {
					return "", e
				}
				resp, e := client.Do(req)
				if e != nil {
					return "", e
				}
				defer resp.Body.Close()
				b, e := io.ReadAll(resp.Body)
				if e != nil {
					return "", e
				}
				if resp.StatusCode != want {
					return "", fmt.Errorf("%s %s status=%d want=%d", method, path, resp.StatusCode, want)
				}
				return string(b), nil
			}
			must := func(method, path, body string, want int) string {
				t.Helper()
				v, e := request(method, path, body, want)
				if e != nil {
					t.Fatal(e)
				}
				return v
			}
			waitPool := func(event string) {
				t.Helper()
				deadline := time.Now().Add(5 * time.Second)
				for {
					inUse := must("GET", "/_stats", "", 200)
					t.Logf("REV7 trace driver=%s event=%s poolInUse=%s", driver, event, inUse)
					if inUse == "0" {
						return
					}
					if time.Now().After(deadline) {
						t.Fatalf("pool did not quiesce after %s; inUse=%s", event, inUse)
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
			must("GET", "/health", "", 200)
			must("GET", "/ready", "", 200)
			// Confirm actual server execution, then disconnect. Activity-view
			// permissions are an explicit fixture prerequisite, never auto-granted.
			activity := "SELECT COUNT(*) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE ID <> CONNECTION_ID() AND INFO LIKE '%" + table + "%'"
			if driver == "postgres" {
				activity = "SELECT COUNT(*) FROM pg_stat_activity WHERE pid <> pg_backend_pid() AND state = 'active' AND query LIKE '%" + table + "%'"
			}
			if driver == "sqlserver" {
				activity = "SELECT COUNT(*) FROM sys.dm_exec_requests r CROSS APPLY sys.dm_exec_sql_text(r.sql_handle) q WHERE r.session_id <> @@SPID AND q.text LIKE '%" + table + "%'"
			}
			inflightCtx, disconnect := context.WithCancel(context.Background())
			defer disconnect()
			inflightReq, _ := http.NewRequestWithContext(inflightCtx, "GET", base+"/_inflight", nil)
			inflightDone := make(chan struct{})
			go func() {
				defer close(inflightDone)
				r, _ := client.Do(inflightReq)
				if r != nil {
					r.Body.Close()
				}
			}()
			activityDeadline := time.Now().Add(2 * time.Second)
			for queryText(activity) == "0" {
				if time.Now().After(activityDeadline) {
					t.Fatal("server activity barrier not observed; check monitoring permissions")
				}
				time.Sleep(10 * time.Millisecond)
			}
			disconnect()
			disconnectedAt := time.Now()
			select {
			case <-inflightDone:
			case <-time.After(5 * time.Second):
				t.Fatal("in-flight disconnect did not return")
			}
			// PostgreSQL and SQL Server cancel the server operation promptly. The
			// MySQL driver cancels the caller and discards the connection, but the
			// server may retain SLEEP until its bounded statement duration ends.
			serverCleanupBudget := 5 * time.Second
			if driver == "mysql" {
				serverCleanupBudget = 12 * time.Second
			}
			activityDeadline = time.Now().Add(serverCleanupBudget)
			for queryText(activity) != "0" {
				if time.Now().After(activityDeadline) {
					t.Fatalf("cancelled SQL remained active on server after %s", serverCleanupBudget)
				}
				time.Sleep(10 * time.Millisecond)
			}
			t.Logf("REV7 trace driver=%s event=inflight-sql serverActive=0 cleanup=%s", driver, time.Since(disconnectedAt))
			waitPool("inflight-disconnect")
			payload := "ไทย O'Brien"
			must("POST", "/items?id=first", payload, 201)
			if got := queryText("SELECT value FROM " + table + " WHERE id='first'"); got != payload {
				t.Fatal("committed data differs")
			}
			if must("GET", "/items/first", "", 200) != payload {
				t.Fatal("HTTP read differs")
			}
			must("POST", "/items?id=first", payload, 409)
			must("PUT", "/items/first", "changed", 200)
			if queryText("SELECT value FROM "+table+" WHERE id='first'") != "changed" {
				t.Fatal("update missing")
			}
			must("DELETE", "/items/first", "", 204)
			must("GET", "/items/first", "", 404)
			must("PUT", "/items/missing", "", 404)
			// Warmup plus three equal batches, bounded concurrency below HTTP admission.
			for batch := -1; batch < 3; batch++ {
				var wg sync.WaitGroup
				slots := make(chan struct{}, 4)
				failures := make(chan error, 12)
				for i := 0; i < 12; i++ {
					slots <- struct{}{}
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						defer func() { <-slots }()
						id := fmt.Sprintf("b%d_r%d", batch+1, i)
						body := "value_" + id
						if _, e := request("POST", "/items?id="+id, body, 201); e != nil {
							failures <- e
							return
						}
						got, e := request("GET", "/items/"+id, "", 200)
						if e != nil {
							failures <- e
						} else if got != body {
							failures <- fmt.Errorf("request state mixed")
						}
					}(i)
				}
				wg.Wait()
				close(failures)
				for e := range failures {
					t.Error(e)
				}
				waitPool(fmt.Sprintf("concurrent-batch-%d", batch))
				var mem runtime.MemStats
				runtime.ReadMemStats(&mem)
				t.Logf("batch=%d requests=24 concurrency=4 heap=%d goroutines=%d", batch, mem.HeapAlloc, runtime.NumGoroutine())
			}
			if queryText("SELECT COUNT(*) FROM "+table) != "48" {
				t.Fatal("concurrent writes lost")
			}
			must("GET", "/_rollback", "", 500)
			select {
			case <-out.barrier:
			case <-time.After(5 * time.Second):
				t.Fatal("rollback barrier timeout")
			}
			if queryText("SELECT COUNT(*) FROM "+table+" WHERE id='barrier'") != "0" {
				t.Fatal("exception failed to roll back")
			}
			// Disconnect before and after commit: the barrier is emitted after INSERT.
			for _, path := range []string{"/_cancel", "/_commit"} {
				requestCtx, stopRequest := context.WithCancel(context.Background())
				req, _ := http.NewRequestWithContext(requestCtx, "GET", base+path, nil)
				returned := make(chan struct{})
				go func() {
					defer close(returned)
					r, _ := client.Do(req)
					if r != nil {
						r.Body.Close()
					}
				}()
				select {
				case <-out.barrier:
				case <-time.After(5 * time.Second):
					stopRequest()
					t.Fatal("write barrier timeout")
				}
				stopRequest()
				<-returned
				waitPool("transaction-" + path)
				want := "0"
				if path == "/_commit" {
					want = "1"
				}
				if queryText("SELECT COUNT(*) FROM "+table+" WHERE id='barrier'") != want {
					t.Fatal("disconnect transaction outcome incorrect")
				}
			}
			must("GET", "/health", "", 200)
			must("GET", "/ready", "", 200)
			// Server-side delay exceeds both deadlines; each driver must unwind and
			// leave the shared pool usable. Request-owned pools close on timeout too.
			for _, path := range []string{"/_slow", "/_dbslow"} {
				started := time.Now()
				must("GET", path, "", 504)
				limit := 6 * time.Second
				if path == "/_dbslow" {
					limit = 2500 * time.Millisecond
				}
				if time.Since(started) > limit {
					t.Fatal("timeout cleanup exceeded fixture budget")
				}
				waitPool("timeout-" + path)
				must("GET", "/ready", "", 200)
			}
			must("GET", "/_stop", "", 200)
			select {
			case e := <-done:
				stopped = true
				if e != nil {
					t.Fatal(e)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("explicit stop timeout")
			}
			// Restart on the same port and fixture name after owner cleanup, then
			// stop while a request holds an uncommitted transaction.
			t.Setenv("CRAFT_HTTP_ADDRESS", address)
			stopped = false
			go func() { done <- interpreter.Run(ctx, program, out) }()
			select {
			case <-out.address:
			case e := <-done:
				stopped = true
				t.Fatalf("restart: %v", e)
			case <-time.After(15 * time.Second):
				t.Fatal("restart timed out")
			}
			activeDone := make(chan struct{})
			go func() { defer close(activeDone); request("GET", "/_child", "", 200) }()
			select {
			case <-out.barrier:
			case <-time.After(5 * time.Second):
				t.Fatal("shutdown transaction barrier timed out")
			}
			must("GET", "/_stop", "", 200)
			select {
			case e := <-done:
				stopped = true
				if e == nil || !strings.Contains(e.Error(), "drain deadline exceeded") {
					t.Fatalf("missing forced drain diagnostic: %v", e)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("active transaction shutdown timed out")
			}
			select {
			case <-activeDone:
			case <-time.After(10 * time.Second):
				t.Fatal("active client did not finish")
			}
		})
	}
}
