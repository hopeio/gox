package exec

import (
	"os/exec"
	"strings"
	"syscall"

	osx "github.com/hopeio/gox/os"
)

// CMD returns the result.
func CMD(s string, opts ...Option) *exec.Cmd {
	var cmd *exec.Cmd
	if strings.Contains(s, "\"") {
		exe := s
		rest := ""
		if s[0] == '"' {
			// 带引号的可执行路径（如 "C:\Program Files\x.exe"）不能按第一个空格切
			if end := strings.IndexByte(s[1:], '"'); end >= 0 {
				exe = s[1 : 1+end]
				rest = s[2+end:]
			}
		} else if i := strings.IndexByte(s, ' '); i >= 0 {
			exe = s[:i]
			rest = s[i:]
		}
		cmd = exec.Command(exe)
		cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: rest, HideWindow: true}
	} else {
		words := osx.Split(s)
		cmd = exec.Command(words[0], words[1:]...)
	}
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

// CmdString returns the result.
func CmdString(cmd *exec.Cmd) string {
	if cmd.SysProcAttr != nil && cmd.SysProcAttr.CmdLine != "" {
		return cmd.Path + cmd.SysProcAttr.CmdLine
	}
	if len(cmd.Args) == 0 {
		return cmd.Path
	}
	return strings.Join(cmd.Args, " ")
}
