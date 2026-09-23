//go:build linux || darwin || freebsd

package project

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func lockInstallation(root string) (func(), error) {
	f, e := os.OpenFile(filepath.Join(root, ".craft-install.busy"), os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	if e = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); e != nil {
		f.Close()
		return nil, fmt.Errorf("another package operation is active")
	}
	return func() { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }, nil
}
