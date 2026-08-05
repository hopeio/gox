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

// CreateTime creates and returns a new instance.
func CreateTime(path string) time.Time {
	fileInfo, _ := os.Stat(path)
	wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	tNanSeconds := wFileSys.CreationTime.Nanoseconds() //returns nanoseconds
	return time.Unix(0, tNanSeconds)
}

// CreateTimeByInfo creates and returns a new instance.
func CreateTimeByInfo(fileInfo os.FileInfo) time.Time {
	wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	tNanSeconds := wFileSys.CreationTime.Nanoseconds() //returns nanoseconds
	return time.Unix(0, tNanSeconds)
}

// CreateTimeByEntry creates and returns a new instance.
func CreateTimeByEntry(entry os.DirEntry) time.Time {
	fileInfo, _ := entry.Info()
	wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	tNanSeconds := wFileSys.CreationTime.Nanoseconds() //returns nanoseconds
	return time.Unix(0, tNanSeconds)
}

// LastWriteTimeByEntry returns the result.
func LastWriteTimeByEntry(entry os.DirEntry) time.Time {
	fileInfo, _ := entry.Info()
	wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	tNanSeconds := wFileSys.LastWriteTime.Nanoseconds() //returns nanoseconds
	return time.Unix(0, tNanSeconds)
}
