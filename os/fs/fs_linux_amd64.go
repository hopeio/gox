/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fs

import (
	"os"
	"syscall"
	"time"
)

// init initializes package state.
func init() {
	syscall.Umask(0)
}

// CreateTime creates and returns a new instance.
func CreateTime(path string) time.Time {
	fileInfo, _ := os.Stat(path)
	stat_t := fileInfo.Sys().(*syscall.Stat_t)
	return time.Unix(stat_t.Ctim.Sec, stat_t.Ctim.Nsec)
}

// CreateTimeByInfo creates and returns a new instance.
func CreateTimeByInfo(fileInfo os.FileInfo) time.Time {
	stat_t := fileInfo.Sys().(*syscall.Stat_t)
	return time.Unix(stat_t.Ctim.Sec, stat_t.Ctim.Nsec)
}

// CreateTimeByEntry creates and returns a new instance.
func CreateTimeByEntry(entry os.DirEntry) time.Time {
	fileInfo, _ := entry.Info()
	stat_t := fileInfo.Sys().(*syscall.Stat_t)
	return time.Unix(stat_t.Ctim.Sec, stat_t.Ctim.Nsec)
}
