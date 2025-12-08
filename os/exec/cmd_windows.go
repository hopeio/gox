package exec

import (
	"os/exec"
	"strings"
	"syscall"

	osx "github.com/hopeio/gox/os"
)

func CMD(s string, opts ...Option) *exec.Cmd {
	var cmd *exec.Cmd
	if strings.Contains(s, "\"") {
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
