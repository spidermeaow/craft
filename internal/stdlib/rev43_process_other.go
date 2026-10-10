//go:build !windows

package stdlib

import "os/exec"

func configureChildProcess(cmd *exec.Cmd) {}
