package os

import "os"

type ProcAttrOption func(attr *os.ProcAttr)

// StartProcess ...
func StartProcess(cmd string, opts ...ProcAttrOption) (*os.Process, error) {
	attr := &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}}
	for _, opt := range opts {
		opt(attr)
	}
	args := Split(cmd)
	return os.StartProcess(args[0], args, attr)
}
