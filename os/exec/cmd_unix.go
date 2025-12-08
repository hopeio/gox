//go:build unix

package exec

import (
	"os/exec"

	osx "github.com/hopeio/gox/os"
)

func CMD(s string, opts ...Option) *exec.Cmd {
	words := osx.Split(s)
	cmd := exec.Command(words[0], words[1:]...)
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}
