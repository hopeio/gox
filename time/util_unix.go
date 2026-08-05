//go:build unix

package time

import (
	"os/exec"
	"time"
)

// SetSysTime ...
func SetSysTime(t time.Time) {
	cmd := exec.Command("date", "-s", t.Format("01/02/2006 15:04:05.999999999"))
	cmd.Run()
}

// SyncHwTime ...
func SyncHwTime() {
	cmd := exec.Command("clock", "--systohc")
	cmd.Run()
}
