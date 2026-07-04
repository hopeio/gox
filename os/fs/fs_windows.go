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

func CreateTime(path string) time.Time {
	fileInfo, _ := os.Stat(path)
	wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	tNanSeconds := wFileSys.CreationTime.Nanoseconds() /// 返回的是纳秒
	return time.Unix(0, tNanSeconds)
}

func CreateTimeByInfo(fileInfo os.FileInfo) time.Time {
	wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	tNanSeconds := wFileSys.CreationTime.Nanoseconds() /// 返回的是纳秒
	return time.Unix(0, tNanSeconds)
}

func CreateTimeByEntry(entry os.DirEntry) time.Time {
	fileInfo, _ := entry.Info()
	wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	tNanSeconds := wFileSys.CreationTime.Nanoseconds() /// 返回的是纳秒
	return time.Unix(0, tNanSeconds)
}

func LastWriteTimeByEntry(entry os.DirEntry) time.Time {
	fileInfo, _ := entry.Info()
	wFileSys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	tNanSeconds := wFileSys.LastWriteTime.Nanoseconds() /// 返回的是纳秒
	return time.Unix(0, tNanSeconds)
}
