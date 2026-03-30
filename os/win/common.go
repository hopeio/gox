/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */
//go:build windows
package win

import (
	"golang.org/x/sys/windows"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

func ConvertDosPath(p string) string {
	rawDrive := strings.Join(strings.Split(p, `\`)[:3], `\`)

	for d := 'A'; d <= 'Z'; d++ {
		szDeviceName := string(d) + ":"
		szTarget := make([]uint16, 512)
		ret, _, _ := procQueryDosDeviceW.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(szDeviceName))),
			uintptr(unsafe.Pointer(&szTarget[0])),
			uintptr(len(szTarget)))
		if ret != 0 && windows.UTF16ToString(szTarget[:]) == rawDrive {
			return filepath.Join(szDeviceName, p[len(rawDrive):])
		}
	}
	return p
}
