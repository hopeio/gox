//go:build unix

package time

import (
	"os/exec"
	"time"
)

// 设置系统时间
func SetSysTime(t time.Time) {
	cmd := exec.Command("date", "-s", t.Format("01/02/2006 15:04:05.999999999"))
	cmd.Run()
}

func SyncHwTime() {
	cmd := exec.Command("clock --systohc")
	cmd.Run()
}
