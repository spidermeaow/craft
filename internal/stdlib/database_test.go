package stdlib

import (
	"context"
	rt "craft/internal/runtime"
	"errors"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

func TestRev5DBConversionAndRedaction(t *testing.T) {
	for _, test := range []struct {
		value         any
		sqlType, kind string
	}{
		{nil, "INT", "null"}, {int64(-1), "BIGINT", "int"}, {[]byte("18446744073709551615"), "UNSIGNED BIGINT", "error"},
		{[]byte("9007199254740993.1200"), "DECIMAL", "decimal"}, {float64(1.1), "DECIMAL", "error"},
		{[]byte{0, 255}, "BYTEA", "bytes"}, {[]byte{255}, "VARCHAR", "error"}, {math.Inf(1), "FLOAT", "error"},
	} {
		v, _, e := dbCell(test.value, test.sqlType)
		if test.kind == "error" {
			if e == nil {
				t.Fatal("accepted unsafe value", test.sqlType)
			}
			continue
		}
		if e != nil || v.kind != test.kind {
			t.Fatalf("%s: kind=%s error=%v", test.sqlType, v.kind, e)
		}
	}
	for _, e := range []error{errors.New("password=secret sql=private"), &mysql.MySQLError{Number: 1062, Message: "secret duplicate value"}, context.DeadlineExceeded, context.Canceled} {
		safe := dbError(e)
		if strings.Contains(safe.Error(), "secret") || strings.Contains(safe.Error(), "private") {
			t.Fatal("driver error leaked")
		}
	}
	_, e := dbParams(&rt.Array{Items: []any{DBValue{"text", strings.Repeat("a", 100)}}}, 32)
	if e == nil {
		t.Fatal("parameter budget not enforced")
	}
}

type countedDBResource struct{ count int }

func (r *countedDBResource) close() error { r.count++; return nil }
func TestRev5DBScopeOwnership(t *testing.T) {
	ctx, closeScope := BeginDBScope(context.Background())
	s := ctx.Value(dbScopeKey{}).(*dbScope)
	for i := 0; i < 1000; i++ {
		r := &countedDBResource{}
		if e := s.add(r); e != nil {
			t.Fatal(e)
		}
		s.remove(r)
	}
	r := &countedDBResource{}
	if e := s.add(r); e != nil {
		t.Fatal(e)
	}
	closeScope()
	closeScope()
	if r.count != 1 {
		t.Fatal("cleanup must happen once")
	}
	if e := s.add(&countedDBResource{}); e == nil {
		t.Fatal("closed scope accepted resource")
	}
}

func TestRev5DBPoolWaitCancellation(t *testing.T) {
	for _, driver := range []string{"mysql", "postgres", "sqlserver"} {
		t.Run(driver, func(t *testing.T) {
			key := map[string]string{"mysql": "CRAFT_MYSQL_DSN", "postgres": "CRAFT_POSTGRES_DSN", "sqlserver": "CRAFT_SQLSERVER_DSN"}[driver]
			dsn := os.Getenv(key)
			if dsn == "" {
				t.Skip("dedicated DB credentials not loaded")
			}
			base, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			ctx, finish := BeginDBScope(base)
			defer finish()
			config, _ := invokeDB(ctx, "std.db.config", nil)
			config.(*rt.Struct).Fields["maxOpen"] = int64(1)
			config.(*rt.Struct).Fields["maxIdle"] = int64(1)
			value, e := invokeDB(ctx, "std.db.connect", []any{driver, dsn, config})
			if e != nil {
				t.Fatal(e)
			}
			db := value.(*DBConnection)
			transaction, e := invokeDB(ctx, "std.db.begin", []any{db})
			if e != nil {
				t.Fatal(e)
			}
			short, stop := context.WithTimeout(ctx, 40*time.Millisecond)
			defer stop()
			_, e = invokeDB(short, "std.db.query", []any{db, "SELECT 1", &rt.Array{}})
			var ex *rt.Exception
			if !errors.As(e, &ex) || ex.Kind != "DbTimeoutError" {
				t.Fatalf("pool wait did not time out: %v", e)
			}
			if db.db.Stats().OpenConnections > 1 {
				t.Fatal("pool exceeded limit")
			}
			if e = transaction.(*DBTransaction).close(); e != nil {
				t.Fatal(e)
			}
			cancelled, stopCancelled := context.WithCancel(ctx)
			stopCancelled()
			_, e = invokeDB(cancelled, "std.db.query", []any{db, "SELECT 1", &rt.Array{}})
			if !errors.As(e, &ex) || ex.Kind != "DbCancelledError" {
				t.Fatalf("cancelled operation accepted: %v", e)
			}
			if _, e = invokeDB(ctx, "std.db.query", []any{db, "SELECT 1", &rt.Array{}}); e != nil {
				t.Fatal(e)
			}
			finish()
			if db.db.Stats().OpenConnections != 0 {
				t.Fatal("owner exit did not close pool")
			}
		})
	}
}
