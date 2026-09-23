package stdlib

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFilesystemAndJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "text.txt")
	var out bytes.Buffer
	ctx := context.Background()
	if _, e := Invoke(ctx, &out, "std.fs.writeText", []any{path, "hello โลก"}); e != nil {
		t.Fatal(e)
	}
	value, e := Invoke(ctx, &out, "std.fs.readText", []any{path})
	if e != nil || value != "hello โลก" {
		t.Fatalf("read: %v %v", value, e)
	}
	value, e = Invoke(ctx, &out, "std.fs.exists", []any{path})
	if e != nil || value != true {
		t.Fatalf("exists: %v %v", value, e)
	}
	value, e = Invoke(ctx, &out, "std.fs.exists", []any{path + ".missing"})
	if e != nil || value != false {
		t.Fatalf("missing: %v %v", value, e)
	}
	if e = os.WriteFile(path, []byte{0xff}, 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = Invoke(ctx, &out, "std.fs.readText", []any{path}); e == nil {
		t.Fatal("read accepted invalid UTF-8")
	}
	value, e = Invoke(ctx, &out, "std.json.quote", []any{"a\nb"})
	if e != nil || value != "\"a\\nb\"" {
		t.Fatalf("quote: %v %v", value, e)
	}
}
func TestTimeAndEnvironment(t *testing.T) {
	t.Setenv("CRAFT_REV1_TEST_VALUE", "visible")
	var out bytes.Buffer
	ctx := context.Background()
	v, e := Invoke(ctx, &out, "std.env.get", []any{"CRAFT_REV1_TEST_VALUE"})
	if e != nil || v != "visible" {
		t.Fatalf("env: %v %v", v, e)
	}
	v, e = Invoke(ctx, &out, "std.time.now", nil)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = time.Parse(time.RFC3339Nano, v.(string)); e != nil {
		t.Fatal(e)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, e = Invoke(canceled, &out, "std.time.sleep", []any{int64(1000)}); e != context.Canceled {
		t.Fatalf("sleep cancellation: %v", e)
	}
	if _, e = Invoke(ctx, &out, "std.time.sleep", []any{int64(-1)}); e == nil {
		t.Fatal("negative sleep accepted")
	}
}
