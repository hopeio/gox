//go:build unix

package exec

import (
	"os/exec"
	"strings"

	osx "github.com/hopeio/gox/os"
)

// CMD ...
func CMD(s string, opts ...Option) *exec.Cmd {
	words := osx.Split(s)
	cmd := exec.Command(words[0], words[1:]...)
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

// CmdString ...
func CmdString(cmd *exec.Cmd) string {
	if len(cmd.Args) == 0 {
		return cmd.Path
	}
	return strings.Join(cmd.Args, " ")
}
