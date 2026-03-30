/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

 //go:build windows
package win

import "golang.org/x/sys/windows"

var (
	modpsapi                     = windows.NewLazySystemDLL("psapi.dll")
	procGetProcessMemoryInfo     = modpsapi.NewProc("GetProcessMemoryInfo")
	procGetProcessImageFileNameW = modpsapi.NewProc("GetProcessImageFileNameW")
)
