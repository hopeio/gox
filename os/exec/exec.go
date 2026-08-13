/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package exec

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	stringsx "github.com/hopeio/gox/strings"
)

// Run executes the operation.
func Run(s string, opts ...Option) error {
	cmd := CMD(s, opts...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunWithLog executes the operation.
func RunWithLog(s string, opts ...Option) error {
	opts = append(opts, func(cmd *exec.Cmd) {
		log.Printf(`exec:"%s"`, CmdString(cmd))
	})
	return Run(s, opts...)
}

type Option func(cmd *exec.Cmd)

// RunGetOut executes the operation.
func RunGetOut(s string, opts ...Option) (string, error) {
	return runGetOutCmd(CMD(s, opts...))
}

// runGetOutCmd performs the operation.
func runGetOutCmd(cmd *exec.Cmd) (string, error) {
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return stringsx.FromBytes(buf), err
	}
	if len(buf) == 0 {
		return "", nil
	}
	lastIndex := len(buf) - 1
	if buf[lastIndex] == '\n' {
		buf = buf[:lastIndex]
	}
	return stringsx.FromBytes(buf), nil
}

// RunGetOutWithLog executes the operation.
func RunGetOutWithLog(s string, opts ...Option) (string, error) {
	cmd := CMD(s, opts...)
	out, err := runGetOutCmd(cmd)
	line := CmdString(cmd)
	if err != nil {
		log.Printf(`exec:"%s" failed,out:%v,err:%v`, line, out, err)
		return out, err
	}
	log.Printf(`exec:"%s"`, line)
	return out, err
}

// Shell run shell
// e.g. Shell("bash", "echo hello world")
func Shell(interpreter, shell string, opts ...Option) error {
	// strconv.Quote 已含引号，再包一层 "..." 会把命令行打坏；opts 由 CMD 应用一次即可
	cmd := CMD(fmt.Sprintf(`%s -c %s`, interpreter, strconv.Quote(shell)), opts...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
