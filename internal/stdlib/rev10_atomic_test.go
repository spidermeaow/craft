package stdlib

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type faultFile struct {
	*os.File
	fault  string
	cancel context.CancelFunc
}

func (f faultFile) Write(b []byte) (int, error) {
	if f.fault == "write" {
		return 0, errors.New("injected")
	}
	if f.fault == "short" {
		return 0, nil
	}
	return f.File.Write(b)
}
func (f faultFile) Close() error {
	err := f.File.Close()
	if f.cancel != nil {
		f.cancel()
	}
	if f.fault == "close" {
		return errors.New("injected")
	}
	return err
}
func TestRev10AtomicFaults(t *testing.T) {
	for _, fault := range []string{"write", "short", "close", "replace", "cancel"} {
		t.Run(fault, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "target")
			if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			create := func(dir string) (atomicFile, error) {
				f, e := os.CreateTemp(dir, ".craft-*")
				var c context.CancelFunc
				if fault == "cancel" {
					c = cancel
				}
				return faultFile{f, fault, c}, e
			}
			rename := func(a, b string) error {
				if fault == "replace" {
					return errors.New("injected")
				}
				return os.Rename(a, b)
			}
			if err := writeAtomicWith(ctx, path, []byte("new"), create, rename, os.Remove); err == nil {
				t.Fatal("failure lost")
			}
			b, e := os.ReadFile(path)
			if e != nil || string(b) != "old" {
				t.Fatal("original damaged")
			}
			entries, e := os.ReadDir(dir)
			if e != nil || len(entries) != 1 {
				t.Fatal("temporary file leaked")
			}
		})
	}
}
