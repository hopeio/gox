/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package exec

import (
	osx "github.com/hopeio/gox/os"
	stringsx "github.com/hopeio/gox/strings"
	"log"
	"os"
	"os/exec"
)

func RunWithLog(arg string, opts ...Option) error {
	words := osx.Split(arg)
	cmd := exec.Command(words[0], words[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	for _, opt := range opts {
		opt(cmd)
	}
	log.Println(cmd.String())
	return cmd.Run()
}

func Run(s string, opts ...Option) error {
	words := osx.Split(s)
	cmd := exec.Command(words[0], words[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd.Run()
}

type Option func(cmd *exec.Cmd)

func RunGetOut(s string, opts ...Option) (string, error) {
	words := osx.Split(s)
	cmd := exec.Command(words[0], words[1:]...)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return stringsx.BytesToString(buf), err
	}
	if len(buf) == 0 {
		return "", nil
	}
	lastIndex := len(buf) - 1
	if buf[lastIndex] == '\n' {
		buf = buf[:lastIndex]
	}
	for _, opt := range opts {
		opt(cmd)
	}
	return stringsx.BytesToString(buf), nil
}

func RunGetOutWithLog(s string, opts ...Option) (string, error) {
	out, err := RunGetOut(s, opts...)
	if err != nil {
		log.Printf(`exec:"%s" failed,out:%v,err:%v`, s, out, err)
		return out, err
	}
	log.Printf(`exec:"%s"`, s)
	return out, err
}
