//go:build windows

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package exec

import (
	stringsx "github.com/hopeio/gox/strings"
	"os"
	"os/exec"
	"syscall"
)

func RunGetOutContainQuoted(s string, opts ...Option) (string, error) {
	cmd := ContainQuotedCMD(s)
	for _, opt := range opts {
		opt(cmd)
	}
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return stringsx.BytesToString(buf), err
	}
	if len(buf) == 0 {
		return "", nil
	}
	return stringsx.BytesToString(buf), nil
}

func RunContainQuoted(s string, opts ...Option) error {
	cmd := ContainQuotedCMD(s)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd.Run()
}

func ContainQuotedCMD(s string, opts ...Option) *exec.Cmd {
	exe := s
	for i, c := range s {
		if c == ' ' {
			exe = s[:i]
			break
		}
	}
	cmd := exec.Command(exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: s[len(exe):], HideWindow: true}
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}
