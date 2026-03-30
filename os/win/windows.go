/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

 //go:build windows
package win

import (
	"fmt"
	"github.com/gonutz/w32/v2"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"path/filepath"
	"syscall"
	"unsafe"
)

func StringToCharPtr(str string) *uint8 {
	chars := append([]byte(str), 0)
	return &chars[0]
}

// 回调函数，用于EnumWindows中的回调函数，第一个参数是hWnd，第二个是自定义穿的参数
func AddElementFunc(hWnd win.HWND, hWndList *[]win.HWND) uintptr {
	*hWndList = append(*hWndList, hWnd)
	return 1
}

// 获取桌面下的所有窗口句柄，包括没有Windows标题的或者是窗口的。
func DesktopWindowHWND() []win.HWND {
	var hWndList []win.HWND
	hL := &hWndList
	_, _, err := syscall.Syscall(procEnumWindows.Addr(), 2, syscall.NewCallback(AddElementFunc), uintptr(unsafe.Pointer(hL)), 0)
	if err != 0 {
		fmt.Println(err)
	}
	return hWndList
}

func FindWindow(title, processName string) win.HWND {
	hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
	if hwnd == 0 {
		handle, _ := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
		if handle == syscall.InvalidHandle {
			return 0
		}
		defer syscall.CloseHandle(handle)
		var processEntry syscall.ProcessEntry32
		processEntry.Size = uint32(unsafe.Sizeof(processEntry))
		for syscall.Process32Next(handle, &processEntry) == nil {
			if syscall.UTF16ToString(processEntry.ExeFile[:]) == processName {
				handle, _ := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, processEntry.ProcessID)
				if handle == 0 {
					return 0
				}
				defer syscall.CloseHandle(handle)
				hwnd, _, _ = findWindow.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(processName))))
				break
			}
		}
	}
	return win.HWND(hwnd)
}

func findProcess(name string) uint32 {

	processIDs, ok := w32.EnumProcesses(make([]uint32, 256))
	if !ok {
		return 0
	}

	for i := range processIDs {
		if processIDs[i] != 0 {
			if name == GetProcName(processIDs[i]) {
				return processIDs[i]
			}
		}
	}

	return 0
}

func GetProcName(pid uint32) string {
	if pid == 0 {
		return "System Idle Process"
	}
	if pid == 4 {
		return "System"
	}

	hProcess, err := windows.OpenProcess(
		windows.PROCESS_VM_READ|windows.PROCESS_QUERY_INFORMATION,
		false,
		pid,
	)
	if err != nil {
		fmt.Println(err)
	}
	defer windows.CloseHandle(hProcess)

	buf := make([]uint16, syscall.MAX_LONG_PATH)
	size := uint32(syscall.MAX_LONG_PATH)
	if err := procQueryFullProcessImageNameW.Find(); err == nil { // Vista+
		ret, _, err := procQueryFullProcessImageNameW.Call(
			uintptr(hProcess),
			uintptr(0),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)))
		if ret == 0 {
			fmt.Println(err)
		}
		return filepath.Base(windows.UTF16ToString(buf[:]))
	}
	// XP fallback
	ret, _, err := procGetProcessImageFileNameW.Call(uintptr(hProcess), uintptr(unsafe.Pointer(&buf[0])), uintptr(size))
	if ret == 0 {
		fmt.Println(err)
	}
	return filepath.Base(ConvertDosPath(windows.UTF16ToString(buf[:])))

}
