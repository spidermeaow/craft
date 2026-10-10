//go:build windows

package stdlib

import (
	"fmt"
	"unicode/utf16"
	"unsafe"
)

const (
	cfUnicodeText = 13
	gmMoveable    = 0x0002
)

var (
	uiOpenClipboard  = uiUser32.NewProc("OpenClipboard")
	uiCloseClipboard = uiUser32.NewProc("CloseClipboard")
	uiEmptyClipboard = uiUser32.NewProc("EmptyClipboard")
	uiGetClipboard   = uiUser32.NewProc("GetClipboardData")
	uiSetClipboard   = uiUser32.NewProc("SetClipboardData")
	uiGlobalAlloc    = uiKernel32.NewProc("GlobalAlloc")
	uiGlobalLock     = uiKernel32.NewProc("GlobalLock")
	uiGlobalUnlock   = uiKernel32.NewProc("GlobalUnlock")
	uiGlobalFree     = uiKernel32.NewProc("GlobalFree")
	uiGlobalSize     = uiKernel32.NewProc("GlobalSize")
)

func openUIClipboard(hwnd uintptr) error {
	if ok, _, err := uiOpenClipboard.Call(hwnd); ok == 0 {
		return fmt.Errorf("UiClipboardError: clipboard is unavailable: %v", err)
	}
	return nil
}

func writeClipboardText(hwnd uintptr, value string) error {
	units := append(utf16.Encode([]rune(value)), 0)
	bytes := uintptr(len(units) * 2)
	if err := openUIClipboard(hwnd); err != nil {
		return err
	}
	defer uiCloseClipboard.Call()
	memory, _, err := uiGlobalAlloc.Call(gmMoveable, bytes)
	if memory == 0 {
		return fmt.Errorf("UiClipboardError: cannot allocate clipboard text: %v", err)
	}
	defer func() {
		if memory != 0 {
			uiGlobalFree.Call(memory)
		}
	}()
	data, _, err := uiGlobalLock.Call(memory)
	if data == 0 {
		return fmt.Errorf("UiClipboardError: cannot lock clipboard text: %v", err)
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(data)), len(units)), units)
	uiGlobalUnlock.Call(memory)
	if ok, _, err := uiEmptyClipboard.Call(); ok == 0 {
		return fmt.Errorf("UiClipboardError: cannot clear clipboard: %v", err)
	}
	if ok, _, err := uiSetClipboard.Call(cfUnicodeText, memory); ok == 0 {
		return fmt.Errorf("UiClipboardError: cannot write clipboard: %v", err)
	}
	memory = 0 // Windows owns the movable block after SetClipboardData succeeds.
	return nil
}

func readClipboardText(hwnd uintptr) (string, error) {
	if err := openUIClipboard(hwnd); err != nil {
		return "", err
	}
	defer uiCloseClipboard.Call()
	memory, _, err := uiGetClipboard.Call(cfUnicodeText)
	if memory == 0 {
		return "", fmt.Errorf("UiClipboardError: Unicode text is unavailable: %v", err)
	}
	size, _, _ := uiGlobalSize.Call(memory)
	if size == 0 || size > (16<<20)+2 {
		return "", fmt.Errorf("UiClipboardError: clipboard text exceeds the supported size")
	}
	data, _, err := uiGlobalLock.Call(memory)
	if data == 0 {
		return "", fmt.Errorf("UiClipboardError: cannot lock clipboard text: %v", err)
	}
	defer uiGlobalUnlock.Call(memory)
	units := unsafe.Slice((*uint16)(unsafe.Pointer(data)), int(size/2))
	length := 0
	for length < len(units) && units[length] != 0 {
		length++
	}
	if length == len(units) {
		return "", fmt.Errorf("UiClipboardError: clipboard text is not terminated")
	}
	return string(utf16.Decode(units[:length])), nil
}
