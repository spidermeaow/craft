package stdlib

import (
	"context"
	rt "craft/internal/runtime"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type signaledReader struct {
	reader  io.Reader
	started chan struct{}
	once    sync.Once
}

func (r *signaledReader) Read(p []byte) (int, error) {
	r.once.Do(func() { close(r.started) })
	return r.reader.Read(p)
}

func TestConcurrentInputRejectedAndCancellationPoisonsSession(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	in := &signaledReader{reader: reader, started: make(chan struct{})}
	s := NewSession(in, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, e := s.Invoke(ctx, io.Discard, "std.io.readLine", nil); done <- e }()
	select {
	case <-in.started:
	case <-time.After(time.Second):
		t.Fatal("reader did not start")
	}
	if _, e := s.Invoke(context.Background(), io.Discard, "std.io.readLine", nil); e == nil || !strings.Contains(e.Error(), "already has a reader") {
		t.Fatalf("overlapping read: %v", e)
	}
	cancel()
	select {
	case e := <-done:
		if e != context.Canceled {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("read did not cancel")
	}
	if _, e := s.Invoke(context.Background(), io.Discard, "std.io.readLine", nil); e == nil || !strings.Contains(e.Error(), "unavailable") {
		t.Fatalf("cancelled session unexpectedly reusable: %v", e)
	}
}

func TestInputLimitsAndEOF(t *testing.T) {
	s := NewSession(strings.NewReader(strings.Repeat("x", MaxLineBytes+1)+"\nvalid\n\n"), nil)
	if _, e := s.Invoke(context.Background(), io.Discard, "std.io.readLine", nil); e == nil {
		t.Fatal("oversized line accepted")
	}
	for _, want := range []rt.Optional{rt.Some("valid"), rt.Some(""), {}} {
		v, e := s.Invoke(context.Background(), io.Discard, "std.io.readLine", nil)
		if e != nil || v != want {
			t.Fatalf("%v %v want %v", v, e, want)
		}
	}
	s = NewSession(strings.NewReader("\xff\n"), nil)
	if _, e := s.Invoke(context.Background(), io.Discard, "std.io.readLine", nil); e == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
}
func TestInputCancellation(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewSession(reader, nil)
	done := make(chan error, 1)
	go func() { _, e := s.Invoke(ctx, io.Discard, "std.io.readLine", nil); done <- e }()
	cancel()
	select {
	case e := <-done:
		if e != context.Canceled {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("readLine ignored cancellation")
	}
}
func TestRev2FilesystemSafety(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "source.txt"), filepath.Join(dir, "copy.txt")
	call := func(n string, a ...any) (any, error) { return Invoke(context.Background(), io.Discard, "std.fs."+n, a) }
	if _, e := call("appendText", src, "โลก"); e != nil {
		t.Fatal(e)
	}
	if _, e := call("appendText", src, "!"); e != nil {
		t.Fatal(e)
	}
	if _, e := call("copyFile", src, dst); e != nil {
		t.Fatal(e)
	}
	if _, e := call("copyFile", src, dst); e == nil {
		t.Fatal("overwrote existing destination")
	}
	if _, e := call("moveFile", src, dst); e == nil {
		t.Fatal("move overwrote destination")
	}
	if b, e := os.ReadFile(src); e != nil || string(b) != "โลก!" {
		t.Fatalf("source damaged: %s %v", b, e)
	}
	if _, e := call("removeFile", dir); e == nil {
		t.Fatal("removeFile accepted a directory")
	}
	moved := filepath.Join(dir, "moved.txt")
	if _, e := call("moveFile", src, moved); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(src); !os.IsNotExist(e) {
		t.Fatal("source not removed")
	}
	v, e := call("listDirectory", dir)
	if e != nil || v.(*rt.Array).String() != "[copy.txt, moved.txt]" {
		t.Fatalf("listing: %v %v", v, e)
	}
}

func TestRev9AtomicFilesystemAndPaths(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.txt")
	call := func(n string, a ...any) (any, error) { return Invoke(context.Background(), io.Discard, "std.fs."+n, a) }
	if _, e := call("writeTextAtomic", path, "first"); e != nil {
		t.Fatal(e)
	}
	if _, e := call("writeTextAtomic", path, "second"); e != nil {
		t.Fatal(e)
	}
	if b, e := os.ReadFile(path); e != nil || string(b) != "second" {
		t.Fatalf("atomic replacement: %q %v", b, e)
	}
	if _, e := call("writeTextAtomic", dir, "must not replace directory"); e == nil {
		t.Fatal("atomic write accepted directory destination")
	}
	if _, e := os.Stat(path); e != nil {
		t.Fatalf("destination disappeared after failed replacement: %v", e)
	}
	bytesPath := filepath.Join(dir, "payload.bin")
	if _, e := call("writeBytesAtomic", bytesPath, Bytes{data: "\x00\xff"}); e != nil {
		t.Fatal(e)
	}
	v, e := call("readBytes", bytesPath)
	if e != nil || v.(Bytes).data != "\x00\xff" {
		t.Fatalf("bytes: %v %v", v, e)
	}
	clean, e := Invoke(context.Background(), io.Discard, "std.path.clean", []any{filepath.Join("a", "b", "..", "c")})
	if e != nil || clean != filepath.Join("a", "c") {
		t.Fatalf("clean: %v %v", clean, e)
	}
	abs, e := Invoke(context.Background(), io.Discard, "std.path.absolute", []any{path})
	if e != nil || !filepath.IsAbs(abs.(string)) {
		t.Fatalf("absolute: %v %v", abs, e)
	}
	rel, e := Invoke(context.Background(), io.Discard, "std.path.relative", []any{dir, path})
	if e != nil || rel != "state.txt" {
		t.Fatalf("relative: %v %v", rel, e)
	}
}
func TestRev2JSONLimitsAndPrecision(t *testing.T) {
	for _, input := range []string{`{"a":1,"a":2}`, `[] true`, `{"a":}`, strings.Repeat("[", 130) + "0" + strings.Repeat("]", 130)} {
		if _, e := jsonInvoke("std.json.parse", []any{input}); e == nil {
			t.Fatalf("accepted %s", input)
		}
	}
	v, e := jsonInvoke("std.json.parse", []any{"9223372036854775807"})
	if e != nil {
		t.Fatal(e)
	}
	n, e := jsonMethod(v.(JSON), "asInt", nil)
	if e != nil || n != int64(9223372036854775807) {
		t.Fatalf("precision lost: %v %v", n, e)
	}
	for _, input := range []string{"9223372036854775808", "1.5", "null"} {
		v, e := jsonInvoke("std.json.parse", []any{input})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = jsonMethod(v.(JSON), "asInt", nil); e == nil {
			t.Fatalf("asInt accepted %s", input)
		}
	}
}
func TestRev2ConversionAndTimeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []any
	}{
		{"std.convert.toInt", []any{"9223372036854775808"}}, {"std.convert.toInt", []any{"1_000"}}, {"std.convert.toFloat", []any{"Infinity"}}, {"std.convert.toFloat", []any{"0x1p2"}}, {"std.convert.toFloat", []any{"1e999"}},
		{"std.string.fixed", []any{1.0, int64(19)}}, {"std.time.parse", []any{"2023-02-29T00:00:00Z"}}, {"std.time.durationMilliseconds", []any{int64(9223372036854775807)}}, {"std.time.fromUnixSeconds", []any{int64(9223372036854775807)}},
	} {
		if _, e := Invoke(context.Background(), io.Discard, tc.name, tc.args); e == nil {
			t.Fatalf("accepted %s %v", tc.name, tc.args)
		}
	}
	t.Setenv("CRAFT_REV2_EMPTY", "")
	v, e := Invoke(context.Background(), io.Discard, "std.env.lookup", []any{"CRAFT_REV2_EMPTY"})
	if e != nil || !v.(rt.Optional).Valid || v.(rt.Optional).Value != "" {
		t.Fatal("empty environment lost")
	}
}
