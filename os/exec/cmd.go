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
	"syscall"

	osx "github.com/hopeio/gox/os"
)

func CMD(s string) *exec.Cmd {
	words := osx.Split(s)
	cmd := exec.Command(words[0], words[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
