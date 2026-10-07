//go:build !windows

package stdlib

import (
	"context"
	"fmt"
)

func startUI(context.Context) (*uiSession, error) {
	return nil, fmt.Errorf("UiUnavailable: native UI provider requires Windows")
}

func (u *uiSession) wakeHost() {}
