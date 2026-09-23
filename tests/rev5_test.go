package tests

import (
	"bytes"
	"context"
	"craft/internal/interpreter"
	"craft/internal/project"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runRev5(t *testing.T, source string) {
	t.Helper()
	p := checked(t, source)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	var out bytes.Buffer
	if e := interpreter.Run(ctx, p, &out); e != nil {
		t.Fatalf("Craft database fixture failed: %v\n%s", e, out.String())
	}
	t.Log(strings.TrimSpace(out.String()))
}

func TestRev5ValueAndTypeContracts(t *testing.T) {
	runRev5(t, `func main(){
assert std.db.isNull(std.db.nullValue())
assert std.db.asInt(std.db.int(9223372036854775807)) == 9223372036854775807
assert std.db.asDecimal(std.db.decimal("12345678901234567890.1234")) == "12345678901234567890.1234"
assert std.db.asBytes(std.db.bytes(std.bytes.fromArray([0,255]))).length() == 2
var caught:Bool=false
try { std.db.asInt(std.db.nullValue()) } catch e { caught=e.type=="DbTypeError" }
assert caught
caught=false
try { std.db.decimal("1e9") } catch e { caught=e.type=="DbTypeError" }
assert caught
print("DB value contracts passed")
}`)
	for _, source := range []string{
		`func main(){let x:DbConnection=DbConnection()}`,
		`func main(){std.db.query("invalid","SELECT 1",[])}`,
		`func main(){let v:DbValue=std.db.int("1")}`,
		`struct DbValue{} func main(){}`,
		`struct DbConfig{} func main(){}`,
		`func main(){std.json.stringify(std.db.nullValue())}`,
	} {
		p := &project.Project{Sources: []project.Source{{Path: "src/main.craft", Text: source}}}
		if _, _, e := p.Compile(); e == nil {
			t.Fatal("accepted invalid database program")
		}
	}
}

func TestRev5DatabasePortableBundle(t *testing.T) {
	if os.Getenv("CRAFT_POSTGRES_DSN") == "" {
		t.Skip("database credentials not loaded")
	}
	dir := t.TempDir()
	if e := os.MkdirAll(filepath.Join(dir, "src"), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(dir, "craft.toml"), []byte("name = \"rev5-portable\"\nversion = \"0.1.0\"\nedition = \"2026\"\nentry = \"main\"\n"), 0600); e != nil {
		t.Fatal(e)
	}
	source := `func main(){let db:DbConnection=std.db.connect("postgres",std.env.get("CRAFT_POSTGRES_DSN"),std.db.config());let r:DbRows=std.db.query(db,"SELECT current_database()",[]);assert std.db.asText(r.rows[0][0])=="craft_test";print("database bundle passed")}`
	if e := os.WriteFile(filepath.Join(dir, "src", "main.craft"), []byte(source), 0600); e != nil {
		t.Fatal(e)
	}
	p, e := project.Load(dir)
	if e != nil {
		t.Fatal(e)
	}
	path, e := p.Build()
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(data, []byte(os.Getenv("CRAFT_POSTGRES_DSN"))) {
		t.Fatal("bundle contains credentials")
	}
	portable := filepath.Join(t.TempDir(), "portable.craftbundle")
	if e = os.WriteFile(portable, data, 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.Remove(filepath.Join(dir, "src", "main.craft")); e != nil {
		t.Fatal(e)
	}
	p, e = project.LoadBundle(portable)
	if e != nil {
		t.Fatal(e)
	}
	program, _, e := p.Compile()
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var out bytes.Buffer
	if e = interpreter.Run(ctx, program, &out); e != nil {
		t.Fatal(e)
	}
	t.Log(strings.TrimSpace(out.String()))
}

func TestRev5DatabaseExample(t *testing.T) {
	p, e := project.LoadTests("../examples/rev5-database")
	if e != nil {
		t.Fatal(e)
	}
	program, _, e := p.Compile()
	if e != nil {
		t.Fatal(e)
	}
	for _, test := range program.Tests {
		if e = interpreter.RunTest(context.Background(), program, test, &bytes.Buffer{}); e != nil {
			t.Fatal(e)
		}
	}
	for _, driver := range []string{"mysql", "postgres", "sqlserver"} {
		t.Run(driver, func(t *testing.T) {
			key := map[string]string{"mysql": "CRAFT_MYSQL_DSN", "postgres": "CRAFT_POSTGRES_DSN", "sqlserver": "CRAFT_SQLSERVER_DSN"}[driver]
			if os.Getenv(key) == "" {
				t.Skip("DB credentials not loaded")
			}
			t.Setenv("CRAFT_DB_DRIVER", driver)
			t.Setenv("CRAFT_DB_DSN", os.Getenv(key))
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			var out bytes.Buffer
			if e := interpreter.Run(ctx, program, &out); e != nil {
				t.Fatal(e)
			}
			if !strings.Contains(out.String(), "database: craft_test") || !strings.Contains(out.String(), "Connections in use: 0") {
				t.Fatal("unexpected example output")
			}
		})
	}
}

func TestRev5DatabaseIntegration(t *testing.T) {
	for _, driver := range []string{"mysql", "postgres", "sqlserver"} {
		t.Run(driver, func(t *testing.T) {
			env := map[string]string{"mysql": "CRAFT_MYSQL_DSN", "postgres": "CRAFT_POSTGRES_DSN", "sqlserver": "CRAFT_SQLSERVER_DSN"}[driver]
			if os.Getenv(env) == "" {
				t.Skip("dedicated database credentials not loaded")
			}
			table := fmt.Sprintf("craft_rev5_%s_%d", driver, time.Now().UnixNano())
			create := "CREATE TABLE " + table + " (id BIGINT PRIMARY KEY, label VARCHAR(100), amount DECIMAL(24,4), payload VARBINARY(32), stamp DATETIME(6), optional_value BIGINT NULL, enabled BOOLEAN)"
			identity := "SELECT DATABASE()"
			delay := "SELECT SLEEP(1)"
			placeholders := "?,?,?,?,?,?,?"
			update := "UPDATE " + table + " SET label=? WHERE id=?"
			remove := "DELETE FROM " + table + " WHERE id=?"
			if driver == "postgres" {
				create = "CREATE TABLE " + table + " (id BIGINT PRIMARY KEY, label VARCHAR(100), amount DECIMAL(24,4), payload BYTEA, stamp TIMESTAMPTZ, optional_value BIGINT NULL, enabled BOOLEAN)"
				identity = "SELECT current_database()"
				delay = "SELECT pg_sleep(1)"
				placeholders = "$1,$2,$3,$4,$5,$6,$7"
				update = "UPDATE " + table + " SET label=$1 WHERE id=$2"
				remove = "DELETE FROM " + table + " WHERE id=$1"
			}
			if driver == "sqlserver" {
				create = "CREATE TABLE " + table + " (id BIGINT PRIMARY KEY, label NVARCHAR(100), amount DECIMAL(24,4), payload VARBINARY(32), stamp DATETIMEOFFSET, optional_value BIGINT NULL, enabled BIT)"
				identity = "SELECT DB_NAME()"
				delay = "WAITFOR DELAY '00:00:01'; SELECT 1"
				placeholders = "@p1,@p2,@p3,@p4,@p5,@p6,@p7"
				update = "UPDATE " + table + " SET label=@p1 WHERE id=@p2"
				remove = "DELETE FROM " + table + " WHERE id=@p1"
			}
			insert := "INSERT INTO " + table + " VALUES (" + placeholders + ")"
			source := fmt.Sprintf(`
func escaped():DbConnection { return std.db.connect(%q,std.env.get(%q),std.db.config()) }
func failTransaction(db:DbConnection){
 let tx:DbTransaction=std.db.begin(db)
 std.db.executeTx(tx,%q,[std.db.int(99),std.db.text("rollback"),std.db.decimal("1.0000"),std.db.bytes(std.bytes.fromArray([1])),std.db.time(std.time.parse("2026-09-20T00:00:00Z")),std.db.nullValue(),std.db.bool(true)])
 throw Exception("rollback fixture")
}
func concurrent(db:DbConnection){ let r:DbRows=std.db.query(db,"SELECT 1",[]); assert std.db.asInt(r.rows[0][0])==1 }
func main(){
 let db:DbConnection=std.db.connect(%q,std.env.get(%q),std.db.config())
 defer { std.db.close(db) }
 let identity:DbRows=std.db.query(db,%q,[])
 assert std.db.asText(identity.rows[0][0])=="craft_test"
 std.db.execute(db,%q,[])
 defer { std.db.execute(db,%q,[]) }
 let values:DbValue[]=[std.db.int(1),std.db.text("ไทย O'Brien"),std.db.decimal("12345678901234567890.1234"),std.db.bytes(std.bytes.fromArray([0,255,1])),std.db.time(std.time.parse("2026-09-20T00:00:00Z")),std.db.nullValue(),std.db.bool(true)]
 let tx:DbTransaction=std.db.begin(db)
 let ins:DbResult=std.db.executeTx(tx,%q,values)
 assert ins.affectedRows != null
 std.db.commit(tx)
 let rows:DbRows=std.db.query(db,%q,[])
 assert rows.rows.length==1
 assert rows.columns.length==7
 assert std.db.asInt(rows.rows[0][0])==1
 assert std.db.asText(rows.rows[0][1])=="ไทย O'Brien"
 assert std.db.asDecimal(rows.rows[0][2])=="12345678901234567890.1234"
 assert std.db.asBytes(rows.rows[0][3]).toArray()[1]==255
 assert std.db.kind(rows.rows[0][4])=="time"
 assert std.db.isNull(rows.rows[0][5])
 assert std.db.kind(rows.rows[0][6])=="bool" || std.db.kind(rows.rows[0][6])=="int"
 var failed:Bool=false
 try { std.db.execute(db,%q,values) } catch e { failed=e.type=="DbConstraintError" }
 assert failed
 std.db.execute(db,%q,[std.db.text("updated"),std.db.int(1)])
 let updated:DbRows=std.db.query(db,%q,[])
 assert std.db.asText(updated.rows[0][0])=="updated"
 try { failTransaction(db) } catch e { assert e.message=="rollback fixture" }
 let count:DbRows=std.db.query(db,%q,[])
 assert std.db.asInt(count.rows[0][0])==1
 let rolled:DbTransaction=std.db.begin(db)
 std.db.executeTx(rolled,%q,[std.db.int(1)])
 std.db.rollback(rolled)
 let count2:DbRows=std.db.query(db,%q,[])
 assert std.db.asInt(count2.rows[0][0])==1
 let child:Task=std.task.spawn(concurrent,db)
 child.join()
 let stats:DbPoolStats=std.db.stats(db)
 assert stats.inUse==0
 for i in 0..270 { let t:DbTransaction=std.db.begin(db); std.db.rollback(t) }
 let closed:DbConnection=escaped()
 failed=false
 try { std.db.ping(closed) } catch e { failed=e.type=="DbClosedError" }
 assert failed
 var config:DbConfig=std.db.config()
 config.timeoutMs=100
 let short:DbConnection=std.db.connect(%q,std.env.get(%q),config)
 failed=false
 try { std.db.query(short,%q,[]) } catch e { failed=true }
 assert failed
 std.db.close(short)
 config.timeoutMs=5000
 config.maxRows=1
 let limited:DbConnection=std.db.connect(%q,std.env.get(%q),config)
 failed=false
 try { std.db.query(limited,"SELECT 1 UNION ALL SELECT 2",[]) } catch e { failed=e.type=="DbLimitError" }
 assert failed
 assert std.db.stats(limited).inUse==0
 std.db.close(limited)

 config.maxRows=1000
 config.maxBytes=32
 let bounded:DbConnection=std.db.connect(%q,std.env.get(%q),config)
 failed=false
 try { std.db.query(bounded,"SELECT 'a long value that exceeds the configured budget'",[]) } catch e { failed=e.type=="DbLimitError" }
 assert failed
 std.db.close(bounded)
 config.maxBytes=4194304
 config.maxOpen=1
 config.maxIdle=1
 config.timeoutMs=200
 let pooled:DbConnection=std.db.connect(%q,std.env.get(%q),config)
 let held:DbTransaction=std.db.begin(pooled)
 assert std.db.stats(pooled).inUse==1
 std.db.rollback(held)
 assert std.db.stats(pooled).open<=1
 std.db.close(pooled)
 failed=false
 try { std.db.query(db,"SELECT craft_rev5_secret_marker FROM craft_rev5_missing",[]) } catch e {
   failed=e.type=="DbQueryError"
   assert !e.message.contains("secret_marker")
 }
 assert failed
 std.db.execute(db,%q,[std.db.int(1)])
 let empty:DbRows=std.db.query(db,%q,[])
 assert std.db.asInt(empty.rows[0][0])==0
 print("Rev.5 database acceptance passed: %s")
}`, driver, env, insert, driver, env, identity, create, "DROP TABLE "+table, insert,
				"SELECT id,label,amount,payload,stamp,optional_value,enabled FROM "+table,
				insert, update, "SELECT label FROM "+table, "SELECT COUNT(*) FROM "+table, remove, "SELECT COUNT(*) FROM "+table,
				driver, env, delay, driver, env, driver, env, driver, env, remove, "SELECT COUNT(*) FROM "+table, driver)
			runRev5(t, source)
		})
	}
}
