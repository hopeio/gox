/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	osx "github.com/hopeio/gox/os"
)

func CMD(s string, opts ...Option) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" && strings.Contains(s, "\"") {
		exe := s
		for i, c := range s {
			if c == ' ' {
				exe = s[:i]
				break
			}
		}
		cmd = exec.Command(exe)
		cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: s[len(exe):], HideWindow: true}
	} else {
		words := osx.Split(s)
		cmd = exec.Command(words[0], words[1:]...)
	}
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}
func WaitShutdown() {
	// Set up signal handling.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	done := make(chan bool, 1)
	go func() {
		sig := <-signals
		fmt.Println("")
		fmt.Println("Disconnection requested via Ctrl+C", sig)
		done <- true
	}()

	fmt.Println("Press Ctrl+C to disconnect.")
	<-done

	os.Exit(0)
}
