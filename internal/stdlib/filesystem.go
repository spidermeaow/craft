package stdlib

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"unicode/utf8"
)

const MaxFileBytes = 16 * 1024 * 1024

// New destructive operations accept regular files only and reject symlinks.
// This is a local scripting API, not a sandbox against hostile filesystem races.
func regular(path string) error {
	info, e := os.Lstat(path)
	if e != nil {
		return fsFailure("inspect", path, e)
	}
	if !info.Mode().IsRegular() {
		return toolingError("FsKindError", fmt.Sprintf("filesystem inspect %q failed: requires a regular file (no symlinks)", path))
	}
	return nil
}

// fsFailure keeps diagnostics useful without exposing arbitrary operating-system
// messages, which may contain configuration or credential-bearing path details.
func fsFailure(operation, path string, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return toolingError("FsNotFoundError", fmt.Sprintf("filesystem %s %q failed: path does not exist", operation, path))
	}
	if errors.Is(err, os.ErrPermission) {
		return toolingError("FsPermissionError", fmt.Sprintf("filesystem %s %q failed: permission denied", operation, path))
	}
	if errors.Is(err, os.ErrInvalid) {
		return toolingError("FsPathError", fmt.Sprintf("filesystem %s %q failed: invalid path", operation, path))
	}
	return toolingError("FsIOError", fmt.Sprintf("filesystem %s %q failed", operation, path))
}

func readFile(path string) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, fsFailure("read", path, e)
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		return nil, fsFailure("inspect", path, e)
	}
	if !info.Mode().IsRegular() {
		return nil, toolingError("FsKindError", fmt.Sprintf("filesystem read %q failed: requires a regular file", path))
	}
	b, e := io.ReadAll(io.LimitReader(f, MaxFileBytes+1))
	if e != nil {
		return nil, fsFailure("read", path, e)
	}
	if len(b) > MaxFileBytes {
		return nil, toolingError("FsLimitError", fmt.Sprintf("filesystem read %q failed: file exceeds 16 MiB limit", path))
	}
	return b, nil
}

// writeAtomic writes a complete replacement in the destination directory. Rename
// is atomic only to the extent provided by the operating system and filesystem.
func writeAtomic(path string, data []byte) error {
	return writeAtomicContext(context.Background(), path, data)
}
func writeAtomicContext(ctx context.Context, path string, data []byte) (result error) {
	return writeAtomicWith(ctx, path, data, func(dir string) (atomicFile, error) { return os.CreateTemp(dir, ".craft-*") }, os.Rename, os.Remove)
}

type atomicFile interface {
	Name() string
	Chmod(os.FileMode) error
	Write([]byte) (int, error)
	Close() error
}

func writeAtomicWith(ctx context.Context, path string, data []byte, create func(string) (atomicFile, error), rename func(string, string) error, remove func(string) error) (result error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(data) > MaxFileBytes {
		return toolingError("FsLimitError", fmt.Sprintf("filesystem write %q failed: file exceeds 16 MiB limit", path))
	}
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() {
			return toolingError("FsKindError", fmt.Sprintf("filesystem replace %q failed: requires a regular file (no symlinks)", path))
		}
	} else if !os.IsNotExist(err) {
		return fsFailure("inspect", path, err)
	}
	dir := filepath.Dir(path)
	tmp, e := create(dir)
	if e != nil {
		return fsFailure("create temporary file for", path, e)
	}
	tmpName := tmp.Name()
	completed := false
	defer func() {
		if !completed {
			if err := remove(tmpName); err != nil && !os.IsNotExist(err) {
				if result != nil {
					result = fmt.Errorf("%w; temporary cleanup failed", result)
				} else {
					result = toolingError("FsIOError", "temporary cleanup failed")
				}
			}
		}
	}()
	if e = tmp.Chmod(0644); e != nil {
		_ = tmp.Close()
		return fsFailure("set permissions for", path, e)
	}
	n, e := tmp.Write(data)
	if e == nil && n != len(data) {
		e = io.ErrShortWrite
	}
	if e != nil {
		_ = tmp.Close()
		return fsFailure("write", path, e)
	}
	if e = tmp.Close(); e != nil {
		return fsFailure("close", path, e)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if e = rename(tmpName, path); e != nil {
		return fsFailure("replace", path, e)
	}
	completed = true
	return nil
}

func copyExclusive(src, dst string) error {
	if e := regular(src); e != nil {
		return e
	}
	in, e := os.Open(src)
	if e != nil {
		return fsFailure("open", src, e)
	}
	defer in.Close()
	out, e := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if e != nil {
		return fsFailure("create", dst, e)
	}
	success := false
	defer func() {
		out.Close()
		if !success {
			os.Remove(dst)
		}
	}()
	if _, e = io.Copy(out, in); e != nil {
		return fsFailure("copy", dst, e)
	}
	if e = out.Close(); e != nil {
		return fsFailure("close", dst, e)
	}
	success = true
	return nil
}
func filesystem(name string, args []any) (any, error) {
	p := args[0].(string)
	switch name {
	case "std.fs.appendText":
		if _, e := os.Lstat(p); e == nil {
			if e = regular(p); e != nil {
				return nil, e
			}
		} else if !os.IsNotExist(e) {
			return nil, fsFailure("inspect", p, e)
		}
		f, e := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if e != nil {
			return nil, fsFailure("append", p, e)
		}
		_, e = io.WriteString(f, args[1].(string))
		closeErr := f.Close()
		if e != nil {
			return nil, fsFailure("append", p, e)
		}
		if closeErr != nil {
			return nil, fsFailure("close", p, closeErr)
		}
		return nil, nil
	case "std.fs.createDirectory":
		if e := os.MkdirAll(p, 0755); e != nil {
			return nil, fsFailure("create directory", p, e)
		}
		return nil, nil
	case "std.fs.listDirectory":
		entries, e := os.ReadDir(p)
		if e != nil {
			return nil, fsFailure("list directory", p, e)
		}
		names := []string{}
		for _, v := range entries {
			names = append(names, v.Name())
		}
		sort.Strings(names)
		return stringArray(names), nil
	case "std.fs.isDirectory":
		info, e := os.Stat(p)
		if os.IsNotExist(e) {
			return false, nil
		}
		if e != nil {
			return nil, fsFailure("inspect", p, e)
		}
		return info.IsDir(), nil
	case "std.fs.removeFile":
		if e := regular(p); e != nil {
			return nil, e
		}
		if e := os.Remove(p); e != nil {
			return nil, fsFailure("remove", p, e)
		}
		return nil, nil
	case "std.fs.copyFile", "std.fs.moveFile":
		if e := copyExclusive(p, args[1].(string)); e != nil {
			return nil, e
		}
		if name == "std.fs.moveFile" {
			if e := os.Remove(p); e != nil {
				return nil, toolingError("FsIOError", fmt.Sprintf("filesystem move %q failed: destination was copied but source removal failed", p))
			}
		}
		return nil, nil
	case "std.fs.writeTextAtomic":
		if !utf8.ValidString(args[1].(string)) {
			return nil, toolingError("FsEncodingError", fmt.Sprintf("filesystem write %q failed: requires UTF-8 text", p))
		}
		return nil, writeAtomic(p, []byte(args[1].(string)))
	case "std.fs.readBytes":
		b, e := readFile(p)
		if e != nil {
			return nil, e
		}
		return Bytes{string(b)}, nil
	case "std.fs.writeBytesAtomic":
		return nil, writeAtomic(p, []byte(args[1].(Bytes).data))
	}
	return nil, fmt.Errorf("unknown filesystem function %s", name)
}
