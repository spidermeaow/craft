package stdlib

import (
	"bytes"
	"context"
	rt "craft/internal/runtime"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestRev10Logging(t *testing.T) {
	s := NewSession(nil, nil)
	var out bytes.Buffer
	s.SetLogWriter(&out)
	fields := rt.NewMap()
	fields.Set("token", "synthetic-secret")
	args := []any{"info", "hello\nworld", fields, &rt.Array{Items: []any{"token"}}}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Invoke(context.Background(), io.Discard, "std.log.write", args); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if strings.Contains(out.String(), "synthetic-secret") {
		t.Fatal("secret leaked")
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 20 {
		t.Fatal("records interleaved")
	}
	for _, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatal("invalid JSON")
		}
	}
	args[0] = "debug"
	before := out.Len()
	s.Invoke(context.Background(), io.Discard, "std.log.write", args)
	if out.Len() != before {
		t.Fatal("filter failed")
	}
	s.SetLogWriter(failingLogWriter{})
	args[0] = "error"
	if _, err := s.Invoke(context.Background(), io.Discard, "std.log.write", args); err == nil {
		t.Fatal("writer failure lost")
	}
}

type failingLogWriter struct{}

func (failingLogWriter) Write([]byte) (int, error) { return 0, errors.New("synthetic-sensitive-error") }

func TestRev10ConfigAndErrors(t *testing.T) {
	_, err := configInvoke("std.config.intRange", []any{"port", "synthetic-secret", int64(1), int64(65535)})
	if err == nil || strings.Contains(err.Error(), "synthetic-secret") {
		t.Fatal("config failure/redaction")
	}
	v, err := configInvoke("std.config.defaultValue", []any{rt.Some(""), "default"})
	if err != nil || v != "" {
		t.Fatal("empty changed")
	}
	_, err = configInvoke("std.config.required", []any{"name", rt.Optional{}})
	if err == nil {
		t.Fatal("missing accepted")
	}
	var ex *rt.Exception
	_, err = readFile(filepath.Join(t.TempDir(), "missing"))
	if !errors.As(err, &ex) || ex.Kind != "FsNotFoundError" {
		t.Fatalf("type: %v", err)
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "target")
	os.WriteFile(p, []byte("old"), 0600)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := writeAtomicContext(ctx, p, []byte("new")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "old" {
		t.Fatal("cancel damaged destination")
	}
}
