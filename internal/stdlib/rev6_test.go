package stdlib

import (
	"context"
	rt "craft/internal/runtime"
	"errors"
	"fmt"
	"strings"
	"syscall"
	"testing"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	mssql "github.com/microsoft/go-mssqldb"
)

func TestRev7ConnectionDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		cause error
		want  string
	}{
		{context.DeadlineExceeded, "timed out"}, {context.Canceled, "cancelled"},
		{fmt.Errorf("private address: %w", syscall.ECONNREFUSED), "refused"},
		{&mysql.MySQLError{Number: 1045, Message: "password secret"}, "authentication"},
		{&mysql.MySQLError{Number: 1049, Message: "private db"}, "does not exist"},
		{&pgconn.PgError{Code: "28P01", Message: "password secret"}, "authentication"},
		{&pgconn.PgError{Code: "3D000", Message: "private db"}, "does not exist"},
		{mssql.Error{Number: 18456, Message: "password secret"}, "authentication"},
		{mssql.Error{Number: 4060, Message: "private db"}, "unavailable or access"},
		{errors.New("password secret private"), "could not connect"},
	} {
		e := dbConnectError(tc.cause).(*rt.Exception)
		if e.Kind != "DbConnectionError" || !strings.Contains(e.Message, tc.want) || strings.Contains(e.Message, "secret") || strings.Contains(e.Message, "private") {
			t.Fatalf("unsafe/incorrect diagnostic: %v", e)
		}
	}
	ctx, closeScope := BeginDBScope(context.Background())
	defer closeScope()
	config, _ := invokeDB(ctx, "std.db.config", nil)
	for _, driver := range []string{"mysql", "postgres", "sqlserver"} {
		for _, dsn := range []string{"", " \t\r\n"} {
			_, e := invokeDB(ctx, "std.db.connect", []any{driver, dsn, config})
			var ex *rt.Exception
			if !errors.As(e, &ex) || ex.Kind != "DbConfigError" || !strings.Contains(ex.Message, "DSN is empty") {
				t.Fatalf("empty DSN: %v", e)
			}
		}
	}
}
