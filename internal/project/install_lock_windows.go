package project

import (
	"fmt"
	"path/filepath"
	"syscall"
)

func lockInstallation(root string) (func(), error) {
	p, e := syscall.UTF16PtrFromString(filepath.Join(root, ".craft-install.busy"))
	if e != nil {
		return nil, e
	}
	h, e := syscall.CreateFile(p, syscall.GENERIC_READ|syscall.GENERIC_WRITE|0x10000, 0, nil, syscall.OPEN_ALWAYS, syscall.FILE_ATTRIBUTE_NORMAL|0x04000000, 0)
	if e != nil {
		return nil, fmt.Errorf("another package operation is active or project is not writable")
	}
	return func() { syscall.CloseHandle(h) }, nil
}
