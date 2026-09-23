package stdlib

import (
	"context"
	"craft/internal/ast"
	rt "craft/internal/runtime"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	mssql "github.com/microsoft/go-mssqldb"
	_ "github.com/microsoft/go-mssqldb/sharedmemory"
)

const dbConnectionType ast.Type = "DbConnection"
const dbTransactionType ast.Type = "DbTransaction"
const dbValueType ast.Type = "DbValue"

func IsDBType(t ast.Type) bool {
	return t == dbConnectionType || t == dbTransactionType || t == dbValueType
}

func dbStructs() []*ast.Struct {
	makeStruct := func(name string, names []string, types []ast.Type) *ast.Struct {
		s := &ast.Struct{Name: name}
		for i, n := range names {
			s.Fields = append(s.Fields, ast.Field{Name: n, Type: types[i], Mutable: true})
		}
		return s
	}
	return []*ast.Struct{
		makeStruct("DbConfig", []string{"maxOpen", "maxIdle", "idleTimeoutMs", "lifetimeMs", "timeoutMs", "maxRows", "maxBytes", "connectTimeoutMs"}, []ast.Type{ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int, ast.Int}),
		makeStruct("DbResult", []string{"affectedRows", "lastInsertId"}, []ast.Type{ast.Int.Optional(), ast.Int.Optional()}),
		makeStruct("DbRows", []string{"columns", "rows"}, []ast.Type{ast.String.Array(), dbValueType.Array().Array()}),
		makeStruct("DbPoolStats", []string{"open", "inUse", "idle", "waitCount"}, []ast.Type{ast.Int, ast.Int, ast.Int, ast.Int}),
	}
}

func init() {
	add := func(n string, r ast.Type, p ...ast.Type) { signatures["std.db."+n] = Signature{Params: p, Result: r} }
	add("config", "DbConfig")
	add("connect", dbConnectionType, ast.String, ast.String, "DbConfig")
	add("ping", ast.Void, dbConnectionType)
	add("close", ast.Void, dbConnectionType)
	add("stats", "DbPoolStats", dbConnectionType)
	add("execute", "DbResult", dbConnectionType, ast.String, dbValueType.Array())
	add("query", "DbRows", dbConnectionType, ast.String, dbValueType.Array())
	add("begin", dbTransactionType, dbConnectionType)
	add("commit", ast.Void, dbTransactionType)
	add("rollback", ast.Void, dbTransactionType)
	add("executeTx", "DbResult", dbTransactionType, ast.String, dbValueType.Array())
	add("queryTx", "DbRows", dbTransactionType, ast.String, dbValueType.Array())
	add("nullValue", dbValueType)
	add("isNull", ast.Bool, dbValueType)
	add("kind", ast.String, dbValueType)
	for _, v := range []struct {
		n string
		t ast.Type
	}{{"int", ast.Int}, {"float", ast.Float}, {"bool", ast.Bool}, {"text", ast.String}, {"decimal", ast.String}, {"bytes", ast.Bytes}, {"time", ast.DateTime}} {
		add(v.n, dbValueType, v.t)
		add("as"+strings.ToUpper(v.n[:1])+v.n[1:], v.t, dbValueType)
	}
}

// Scope ownership makes cleanup deterministic, including exception/cancellation.
// Closed resources remove themselves, so long-running scopes do not retain history.
type dbScopeKey struct{}
type dbResource interface{ close() error }
type dbScope struct {
	mu        sync.Mutex
	closed    bool
	resources map[dbResource]bool
}

func BeginDBScope(ctx context.Context) (context.Context, func()) {
	s := &dbScope{resources: map[dbResource]bool{}}
	return context.WithValue(ctx, dbScopeKey{}, s), func() {
		s.mu.Lock()
		s.closed = true
		list := make([]dbResource, 0, len(s.resources))
		for r := range s.resources {
			list = append(list, r)
		}
		s.resources = nil
		s.mu.Unlock()
		// Rollbacks before pools; resource close is idempotent.
		for _, r := range list {
			if _, ok := r.(*DBTransaction); ok {
				_ = r.close()
			}
		}
		for _, r := range list {
			if _, ok := r.(*DBTransaction); !ok {
				_ = r.close()
			}
		}
	}
}
func (s *dbScope) add(r dbResource) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || len(s.resources) >= 256 {
		return dbFailure("DbLimitError", "resource scope is closed or has 256 live database handles")
	}
	s.resources[r] = true
	return nil
}
func (s *dbScope) remove(r dbResource) { s.mu.Lock(); delete(s.resources, r); s.mu.Unlock() }

type dbSettings struct {
	maxOpen, maxIdle        int
	idle, lifetime, timeout time.Duration
	connectTimeout          time.Duration
	maxRows, maxBytes       int
}
type DBConnection struct {
	mu           sync.Mutex
	db           *sql.DB
	owner        *dbScope
	settings     dbSettings
	closed       bool
	transactions map[*DBTransaction]bool
}

func (*DBConnection) String() string { return "DbConnection" }
func (d *DBConnection) close() error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil
	}
	d.closed = true
	txs := make([]*DBTransaction, 0, len(d.transactions))
	for t := range d.transactions {
		txs = append(txs, t)
	}
	d.mu.Unlock()
	for _, t := range txs {
		_ = t.close()
	}
	e := d.db.Close()
	d.owner.remove(d)
	return dbError(e)
}
func (d *DBConnection) check() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return dbFailure("DbClosedError", "connection is closed or its owning function has returned")
	}
	return nil
}

type DBTransaction struct {
	mu         sync.Mutex
	tx         *sql.Tx
	connection *DBConnection
	owner      *dbScope
	cancel     context.CancelFunc
	terminal   bool
}

func (*DBTransaction) String() string { return "DbTransaction" }
func (t *DBTransaction) finish(commit bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.terminal {
		if commit {
			return dbFailure("DbClosedError", "transaction has ended")
		}
		return nil
	}
	t.terminal = true
	var e error
	if commit {
		e = t.tx.Commit()
	} else {
		e = t.tx.Rollback()
	}
	t.cancel()
	t.owner.remove(t)
	t.connection.mu.Lock()
	delete(t.connection.transactions, t)
	t.connection.mu.Unlock()
	if !commit && errors.Is(e, sql.ErrTxDone) {
		return nil
	}
	return dbError(e)
}
func (t *DBTransaction) close() error { return t.finish(false) }

// DbValue is an immutable tagged value; Clone preserves its value/immutable bytes.
type DBValue struct {
	kind  string
	value any
}

func (v DBValue) String() string { return "DbValue(" + v.kind + ")" }

var decimalPattern = regexp.MustCompile(`^[+-]?[0-9]+(\.[0-9]+)?$`)

func dbFailure(kind, message string) error { return &rt.Exception{Kind: kind, Message: message} }

// Connection failures preserve the Rev.5 exception type. Only typed causes
// select a diagnostic; raw driver text can contain credentials and SQL values.
func dbConnectError(e error) error {
	message := "could not connect to the configured database"
	var my *mysql.MySQLError
	var pg *pgconn.PgError
	var ms mssql.Error
	var network net.Error
	switch {
	case errors.Is(e, context.Canceled):
		message = "database connection cancelled"
	case errors.Is(e, context.DeadlineExceeded):
		message = "database connection timed out"
	case errors.Is(e, syscall.ECONNREFUSED):
		message = "database connection refused; check server and transport settings"
	case errors.As(e, &network) && network.Timeout():
		message = "database connection timed out"
	case errors.As(e, &my):
		switch my.Number {
		case 1045:
			message = "database authentication failed"
		case 1049:
			message = "configured database does not exist"
		}
	case errors.As(e, &pg):
		switch pg.Code {
		case "28P01", "28000":
			message = "database authentication failed"
		case "3D000":
			message = "configured database does not exist"
		}
	case errors.As(e, &ms):
		switch ms.Number {
		case 18456:
			message = "database authentication failed"
		case 4060:
			message = "configured database is unavailable or access was denied"
		}
	}
	return dbFailure("DbConnectionError", message)
}
func dbError(e error) error {
	if e == nil {
		return nil
	}
	// Never forward driver messages: they can contain DSNs, SQL or user data.
	kind, message := "DbQueryError", "database operation failed"
	if errors.Is(e, context.DeadlineExceeded) {
		return dbFailure("DbTimeoutError", "database operation timed out")
	}
	if errors.Is(e, context.Canceled) {
		return dbFailure("DbCancelledError", "database operation cancelled")
	}
	if errors.Is(e, sql.ErrTxDone) {
		return dbFailure("DbClosedError", "transaction has ended")
	}
	var my *mysql.MySQLError
	var pg *pgconn.PgError
	var ms mssql.Error
	switch {
	case errors.As(e, &my):
		message = fmt.Sprintf("database operation failed (MySQL %d)", my.Number)
		if my.Number == 1062 || my.Number == 1451 || my.Number == 1452 || my.Number == 1048 {
			kind = "DbConstraintError"
		}
	case errors.As(e, &pg):
		message = "database operation failed (PostgreSQL " + pg.Code + ")"
		if strings.HasPrefix(pg.Code, "23") {
			kind = "DbConstraintError"
		}
	case errors.As(e, &ms):
		message = fmt.Sprintf("database operation failed (SQL Server %d)", ms.Number)
		if ms.Number == 2601 || ms.Number == 2627 || ms.Number == 547 || ms.Number == 515 {
			kind = "DbConstraintError"
		}
	}
	return dbFailure(kind, message)
}
func databaseSettings(v *rt.Struct) (dbSettings, error) {
	get := func(n string) int64 { return v.Fields[n].(int64) }
	if get("connectTimeoutMs") < 1 || get("connectTimeoutMs") > 300000 || get("maxOpen") < 1 || get("maxOpen") > 128 || get("maxIdle") < 0 || get("maxIdle") > get("maxOpen") || get("timeoutMs") < 1 || get("timeoutMs") > 300000 || get("idleTimeoutMs") < 1 || get("idleTimeoutMs") > 86400000 || get("lifetimeMs") < 1 || get("lifetimeMs") > 86400000 || get("maxRows") < 1 || get("maxRows") > 10000 || get("maxBytes") < 1 || get("maxBytes") > 16*1024*1024 {
		return dbSettings{}, dbFailure("DbConfigError", "invalid database limits; see DbConfig contract")
	}
	return dbSettings{int(get("maxOpen")), int(get("maxIdle")), time.Duration(get("idleTimeoutMs")) * time.Millisecond, time.Duration(get("lifetimeMs")) * time.Millisecond, time.Duration(get("timeoutMs")) * time.Millisecond, time.Duration(get("connectTimeoutMs")) * time.Millisecond, int(get("maxRows")), int(get("maxBytes"))}, nil
}

func invokeDB(ctx context.Context, name string, a []any) (any, error) {
	n := strings.TrimPrefix(name, "std.db.")
	if n == "config" {
		return nativeStruct("DbConfig", int64(8), int64(2), int64(60000), int64(300000), int64(5000), int64(1000), int64(4*1024*1024), int64(5000)), nil
	}
	if n == "nullValue" {
		return DBValue{kind: "null"}, nil
	}
	if n == "kind" {
		return a[0].(DBValue).kind, nil
	}
	if n == "isNull" {
		return a[0].(DBValue).kind == "null", nil
	}
	if strings.HasPrefix(n, "as") {
		v := a[0].(DBValue)
		k := strings.ToLower(n[2:3]) + n[3:]
		if v.kind != k {
			return nil, dbFailure("DbTypeError", "database value is "+v.kind+", expected "+k)
		}
		return v.value, nil
	}
	switch n {
	case "int", "bool", "text", "bytes", "time", "float", "decimal":
		if n == "float" && (math.IsNaN(a[0].(float64)) || math.IsInf(a[0].(float64), 0)) {
			return nil, dbFailure("DbTypeError", "non-finite database float")
		}
		if n == "decimal" && (len(a[0].(string)) > 1024 || !decimalPattern.MatchString(a[0].(string))) {
			return nil, dbFailure("DbTypeError", "decimal requires up to 1024 characters of plain decimal notation")
		}
		return DBValue{n, a[0]}, nil
	case "connect":
		if strings.TrimSpace(a[1].(string)) == "" {
			return nil, dbFailure("DbConfigError", "database DSN is empty; set the connection string in the process environment or application configuration")
		}
		scope, _ := ctx.Value(dbScopeKey{}).(*dbScope)
		if scope == nil {
			return nil, dbFailure("DbConfigError", "database connection requires an execution scope")
		}
		driver := a[0].(string)
		switch driver {
		case "postgres":
			driver = "pgx"
		case "mysql", "sqlserver":
		default:
			return nil, dbFailure("DbConfigError", "driver must be mysql, postgres or sqlserver")
		}
		cfg, e := databaseSettings(a[2].(*rt.Struct))
		if e != nil {
			return nil, e
		}
		db, e := sql.Open(driver, a[1].(string))
		if e != nil {
			return nil, dbFailure("DbConnectionError", "invalid database connection configuration")
		}
		db.SetMaxOpenConns(cfg.maxOpen)
		db.SetMaxIdleConns(cfg.maxIdle)
		db.SetConnMaxIdleTime(cfg.idle)
		db.SetConnMaxLifetime(cfg.lifetime)
		pingCtx, cancel := context.WithTimeout(ctx, cfg.connectTimeout)
		defer cancel()
		if e = db.PingContext(pingCtx); e != nil {
			_ = db.Close()
			return nil, dbConnectError(e)
		}
		d := &DBConnection{db: db, owner: scope, settings: cfg, transactions: map[*DBTransaction]bool{}}
		if e = scope.add(d); e != nil {
			_ = db.Close()
			return nil, e
		}
		return d, nil
	case "close":
		return nil, a[0].(*DBConnection).close()
	case "commit":
		return nil, a[0].(*DBTransaction).finish(true)
	case "rollback":
		return nil, a[0].(*DBTransaction).finish(false)
	}
	var d *DBConnection
	var exec dbExecutor
	if n == "queryTx" || n == "executeTx" {
		t := a[0].(*DBTransaction)
		if !t.mu.TryLock() {
			return nil, dbFailure("DbBusyError", "transaction is already in use")
		}
		defer t.mu.Unlock()
		if t.terminal {
			return nil, dbFailure("DbClosedError", "transaction has ended")
		}
		d = t.connection
		exec = t.tx
	} else {
		d = a[0].(*DBConnection)
		exec = d.db
	}
	if e := d.check(); e != nil {
		return nil, e
	}
	if n == "stats" {
		s := d.db.Stats()
		return nativeStruct("DbPoolStats", int64(s.OpenConnections), int64(s.InUse), int64(s.Idle), s.WaitCount), nil
	}
	if n == "begin" {
		scope, _ := ctx.Value(dbScopeKey{}).(*dbScope)
		if scope == nil {
			return nil, dbFailure("DbConfigError", "transaction requires execution scope")
		}
		// The configured timeout also bounds the entire transaction lifetime.
		txCtx, cancel := context.WithTimeout(ctx, d.settings.timeout)
		tx, e := d.db.BeginTx(txCtx, nil)
		if e != nil {
			cancel()
			return nil, dbError(e)
		}
		t := &DBTransaction{tx: tx, connection: d, owner: scope, cancel: cancel}
		d.mu.Lock()
		if d.closed {
			d.mu.Unlock()
			_ = tx.Rollback()
			cancel()
			return nil, dbFailure("DbClosedError", "connection closed during begin")
		}
		d.transactions[t] = true
		d.mu.Unlock()
		if e = scope.add(t); e != nil {
			_ = t.close()
			return nil, e
		}
		return t, nil
	}
	opCtx, cancel := context.WithTimeout(ctx, d.settings.timeout)
	defer cancel()
	if n == "ping" {
		return nil, dbError(d.db.PingContext(opCtx))
	}
	params, e := dbParams(a[2].(*rt.Array), d.settings.maxBytes)
	if e != nil {
		return nil, e
	}
	query := a[1].(string)
	if len(query) > 1024*1024 {
		return nil, dbFailure("DbLimitError", "SQL text exceeds 1 MiB")
	}
	if n == "execute" || n == "executeTx" {
		r, e := exec.ExecContext(opCtx, query, params...)
		if e != nil {
			return nil, dbError(e)
		}
		affected, id := rt.Optional{}, rt.Optional{}
		if v, e := r.RowsAffected(); e == nil {
			affected = rt.Some(v)
		}
		if v, e := r.LastInsertId(); e == nil {
			id = rt.Some(v)
		}
		return nativeStruct("DbResult", affected, id), nil
	}
	return dbQuery(opCtx, exec, query, params, d.settings)
}

type dbExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func dbParams(a *rt.Array, limit int) ([]any, error) {
	if len(a.Items) > 2000 {
		return nil, dbFailure("DbLimitError", "at most 2000 SQL parameters")
	}
	params := make([]any, len(a.Items))
	size := 0
	for i, item := range a.Items {
		v := item.(DBValue)
		size += 16
		switch x := v.value.(type) {
		case Bytes:
			params[i] = []byte(x.data)
			size += len(x.data)
		case DateTime:
			params[i] = x.value
		case string:
			params[i] = x
			size += len(x)
		default:
			params[i] = x
		}
	}
	if size > limit {
		return nil, dbFailure("DbLimitError", "SQL parameters exceed configured maxBytes")
	}
	return params, nil
}
func dbQuery(ctx context.Context, exec dbExecutor, query string, params []any, cfg dbSettings) (any, error) {
	rows, e := exec.QueryContext(ctx, query, params...)
	if e != nil {
		return nil, dbError(e)
	}
	defer rows.Close()
	columns, e := rows.Columns()
	if e != nil {
		return nil, dbError(e)
	}
	types, e := rows.ColumnTypes()
	if e != nil {
		return nil, dbError(e)
	}
	if len(columns) > 256 {
		return nil, dbFailure("DbLimitError", "query exceeds 256 columns")
	}
	names := &rt.Array{}
	result := &rt.Array{}
	size := 0
	for _, n := range columns {
		names.Items = append(names.Items, n)
		size += len(n) + 16
	}
	for rows.Next() {
		if len(result.Items) >= cfg.maxRows {
			return nil, dbFailure("DbLimitError", "query exceeds maxRows; limit SQL or paginate explicitly")
		}
		values := make([]any, len(columns))
		dest := make([]any, len(columns))
		for i := range values {
			dest[i] = &values[i]
		}
		if e = rows.Scan(dest...); e != nil {
			return nil, dbError(e)
		}
		row := &rt.Array{}
		size += 16
		for i, v := range values {
			value, bytes, e := dbCell(v, types[i].DatabaseTypeName())
			if e != nil {
				return nil, e
			}
			size += bytes + 32
			if size > cfg.maxBytes {
				return nil, dbFailure("DbLimitError", "query exceeds maxBytes")
			}
			row.Items = append(row.Items, value)
		}
		result.Items = append(result.Items, row)
	}
	if e = rows.Err(); e != nil {
		return nil, dbError(e)
	}
	if rows.NextResultSet() {
		return nil, dbFailure("DbTypeError", "multiple result sets are not supported")
	}
	if e = rows.Err(); e != nil {
		return nil, dbError(e)
	}
	if size > cfg.maxBytes {
		return nil, dbFailure("DbLimitError", "query metadata exceeds maxBytes")
	}
	if e = rows.Close(); e != nil {
		return nil, dbError(e)
	}
	return nativeStruct("DbRows", names, result), nil
}
func dbCell(v any, typeName string) (DBValue, int, error) {
	if v == nil {
		return DBValue{kind: "null"}, 0, nil
	}
	n := strings.ToUpper(typeName)
	decimal := strings.Contains(n, "DECIMAL") || strings.Contains(n, "NUMERIC") || n == "MONEY" || n == "SMALLMONEY"
	if decimal {
		var text string
		switch x := v.(type) {
		case string:
			text = x
		case []byte:
			text = string(x)
		case int64:
			text = strconv.FormatInt(x, 10)
		default:
			return DBValue{}, 0, dbFailure("DbTypeError", "driver did not return an exact decimal representation")
		}
		if len(text) > 1024 || !decimalPattern.MatchString(text) {
			return DBValue{}, 0, dbFailure("DbTypeError", "unsupported decimal representation")
		}
		return DBValue{"decimal", text}, len(text), nil
	}
	switch x := v.(type) {
	case int64:
		return DBValue{"int", x}, 8, nil
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			break
		}
		return DBValue{"float", x}, 8, nil
	case bool:
		return DBValue{"bool", x}, 1, nil
	case time.Time:
		return DBValue{"time", DateTime{x}}, 32, nil
	case string:
		if !utf8.ValidString(x) {
			break
		}
		return DBValue{"text", x}, len(x), nil
	case []byte:
		if strings.Contains(n, "BINARY") || strings.Contains(n, "BLOB") || n == "BYTEA" || n == "IMAGE" || n == "BIT" || n == "VARBIT" {
			return DBValue{"bytes", Bytes{string(x)}}, len(x), nil
		}
		if strings.Contains(n, "INT") {
			value, e := strconv.ParseInt(string(x), 10, 64)
			if e != nil {
				break
			}
			return DBValue{"int", value}, 8, nil
		}
		if !utf8.Valid(x) {
			break
		}
		return DBValue{"text", string(x)}, len(x), nil
	}
	return DBValue{}, 0, dbFailure("DbTypeError", "unsupported database value or numeric overflow")
}
