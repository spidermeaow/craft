package stdlib

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestRev9WindowsLockedAndReadOnly(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(p, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	ptr, err := syscall.UTF16PtrFromString(p)
	if err != nil {
		t.Fatal(err)
	}
	h, err := syscall.CreateFile(ptr, syscall.GENERIC_READ, syscall.FILE_SHARE_READ, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = writeAtomic(p, []byte("new"))
	syscall.CloseHandle(h)
	if err == nil {
		t.Fatal("replaced locked destination")
	}
	if err = os.Chmod(p, 0400); err != nil {
		t.Fatal(err)
	}
	err = writeAtomic(p, []byte("new"))
	os.Chmod(p, 0600)
	if err == nil {
		t.Fatal("replaced read-only destination")
	}
	b, err := os.ReadFile(p)
	if err != nil || string(b) != "old" {
		t.Fatalf("original damaged: %q %v", b, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary files leaked: %v %v", entries, err)
	}
}

func TestRev9WindowsPathsAndLimits(t *testing.T) {
	for _, tc := range [][2]string{{`C:/a/../b`, `C:\b`}, {`\\server\share\a\..\b`, `\\server\share\b`}, {`C:a\..\b`, `C:b`}, {``, `.`}} {
		v, err := Invoke(context.Background(), io.Discard, "std.path.clean", []any{tc[0]})
		if err != nil || v != tc[1] {
			t.Fatalf("clean %q: %v %v", tc[0], v, err)
		}
	}
	if _, err := Invoke(context.Background(), io.Discard, "std.path.relative", []any{`C:\a`, `D:\b`}); err == nil {
		t.Fatal("cross-volume relative accepted")
	}
	p := filepath.Join(t.TempDir(), "out")
	if err := writeAtomic(p, []byte(strings.Repeat("x", MaxFileBytes+1))); err == nil {
		t.Fatal("oversized write accepted")
	}
	if _, err := Invoke(context.Background(), io.Discard, "std.fs.writeTextAtomic", []any{p, "\xff"}); err == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("invalid write created destination")
	}
}
