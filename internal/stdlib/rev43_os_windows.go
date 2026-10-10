//go:build windows

package stdlib

import (
	"context"
	"craft/internal/ast"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

var rev43ShellExecute = syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW")
var rev43GetVersion = syscall.NewLazyDLL("ntdll.dll").NewProc("RtlGetVersion")

type rev43Version struct {
	size     uint32
	major    uint32
	minor    uint32
	build    uint32
	platform uint32
	service  [128]uint16
}

func init() {
	signatures["std.os.version"] = Signature{Result: ast.String}
	signatures["std.os.architecture"] = Signature{Result: ast.String}
	signatures["std.os.processId"] = Signature{Result: ast.Int}
	signatures["std.os.open"] = Signature{Params: []ast.Type{ast.String}, Result: ast.Void}
}

func invokeOS(ctx context.Context, name string, a []any) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch name {
	case "std.os.version":
		version := rev43Version{size: uint32(unsafe.Sizeof(rev43Version{}))}
		result, _, _ := rev43GetVersion.Call(uintptr(unsafe.Pointer(&version)))
		if result != 0 {
			return nil, toolingError("OsError", "could not read Windows version")
		}
		return fmt.Sprintf("Windows %d.%d build %d", version.major, version.minor, version.build), nil
	case "std.os.architecture":
		return runtime.GOARCH, nil
	case "std.os.processId":
		return int64(os.Getpid()), nil
	case "std.os.open":
		target := a[0].(string)
		if len(target) == 0 || len(target) > 32768 || strings.ContainsAny(target, "\x00\r\n") {
			return nil, toolingError("OsOpenError", "invalid open target")
		}
		isDrivePath := len(target) >= 3 && target[1] == ':' && (target[2] == '\\' || target[2] == '/')
		if !isDrivePath && !strings.HasPrefix(target, `\\`) {
			parsed, err := url.Parse(target)
			if err != nil {
				return nil, toolingError("OsOpenError", "invalid open target")
			}
			if parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "mailto" {
				return nil, toolingError("OsOpenError", "unsupported URL scheme")
			}
		}
		open, _ := syscall.UTF16PtrFromString("open")
		path, err := syscall.UTF16PtrFromString(target)
		if err != nil {
			return nil, toolingError("OsOpenError", "invalid open target")
		}
		result, _, _ := rev43ShellExecute.Call(0, uintptr(unsafe.Pointer(open)), uintptr(unsafe.Pointer(path)), 0, 0, 1)
		if result <= 32 {
			return nil, toolingError("OsOpenError", "Windows could not open target")
		}
		return nil, nil
	}
	return nil, toolingError("OsError", "unknown OS operation")
}
