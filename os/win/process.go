/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package win

import (
	"fmt"
	"github.com/gonutz/w32/v2"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	CREATE_UNICODE_ENVIRONMENT = 0x00000400
	CREATE_NO_WINDOW           = 0x08000000
	NORMAL_PRIORITY_CLASS      = 0x20

	INVALID_SESSION_ID        = 0xFFFFFFFF
	WTS_CURRENT_SERVER_HANDLE = 0

	TOKEN_DUPLICATE    = 0x0002
	MAXIMUM_ALLOWED    = 0x2000000
	CREATE_NEW_CONSOLE = 0x00000010

	IDLE_PRIORITY_CLASS     = 0x40
	HIGH_PRIORITY_CLASS     = 0x80
	REALTIME_PRIORITY_CLASS = 0x100
	GENERIC_ALL_ACCESS      = 0x10000000
)

// 先来两个API,这个貌似使用syscall也可以.
// 刚刚开始写,不知道syscall已经实现了一部分API,就自己动手写了

// Win32进程结构体
type PROCESSENTRY32 struct {
	dwSize              uint32    // 结构大小
	cntUsage            uint32    // 此进程的引用计数
	th32ProcessID       uint32    // 进程id
	th32DefaultHeapID   uintptr   // 进程默认堆id
	th32ModuleID        uint32    // 进程模块id
	cntThreads          uint32    // 进程的线程数
	th32ParentProcessID uint32    // 父进程id
	pcPriClassBase      uint32    // 线程优先权
	dwFlags             uint32    // 保留
	szExeFile           [260]byte // 进程全名
}

type SW struct {
	SW_HIDE            uint16 // 0,
	SW_SHOWNORMAL      uint16 // 1,
	SW_NORMAL          uint16 // 1,
	SW_SHOWMINIMIZED   uint16 // 2,
	SW_SHOWMAXIMIZED   uint16 // 3,
	SW_MAXIMIZE        uint16 // 3,
	SW_SHOWNOACTIVATE  uint16 // 4,
	SW_SHOW            uint16 // 5,
	SW_MINIMIZE        uint16 // 6,
	SW_SHOWMINNOACTIVE uint16 // 7,
	SW_SHOWNA          uint16 // 8,
	SW_RESTORE         uint16 // 9,
	SW_SHOWDEFAULT     uint16 // 10,
	SW_MAX             uint16 // 10
}

var (
	GBKEncoder transform.Transformer = simplifiedchinese.GBK.NewEncoder()
	GBKDecoder transform.Transformer = simplifiedchinese.GBK.NewDecoder()
	ISW                              = SW{
		SW_HIDE:            0,
		SW_SHOWNORMAL:      1,
		SW_NORMAL:          1,
		SW_SHOWMINIMIZED:   2,
		SW_SHOWMAXIMIZED:   3,
		SW_MAXIMIZE:        3,
		SW_SHOWNOACTIVATE:  4,
		SW_SHOW:            5,
		SW_MINIMIZE:        6,
		SW_SHOWMINNOACTIVE: 7,
		SW_SHOWNA:          8,
		SW_RESTORE:         9,
		SW_SHOWDEFAULT:     10,
		SW_MAX:             10,
	}
)

func (p *PROCESSENTRY32) Name() string {
	// string(process.szExeFile[0:]
	name, _, _ := transform.String(GBKDecoder, string(p.szExeFile[0:])) //string(p.szExeFile[0:])
	name = name[:strings.LastIndex(name, ".exe")+4]
	return name
}
func (p *PROCESSENTRY32) ModuleID() string {
	return strconv.Itoa(int(p.th32ModuleID))
}
func (p *PROCESSENTRY32) PID() uint32 {
	return p.th32ProcessID
}

// GetProcessByName 根据pid获取windows系统的某一个进程
//
//	参数:
//	name    string  进程名称, 建议加上.exe结尾
//	return  Process
func GetProcessByName(name string) (PROCESSENTRY32, error) {
	var targetProcess PROCESSENTRY32
	targetProcess = PROCESSENTRY32{
		dwSize: 0,
	}

	pHandle := w32.CreateToolhelp32Snapshot(0x2, 0x0)

	if int(pHandle) == -1 {
		return targetProcess, fmt.Errorf("error:Can not find any proess.")
	}
	defer w32.CloseHandle(pHandle)

	for {
		var proc PROCESSENTRY32
		proc.dwSize = uint32(unsafe.Sizeof(proc))

		if Process32Next(pHandle, uintptr(unsafe.Pointer(&proc))) {
			pname := proc.Name()
			xpoint := strings.LastIndex(pname, ".exe")
			if pname == name || (xpoint > 0 && pname[:xpoint] == name) {
				return proc, nil
			}
		} else {
			break
		}
	}
	return targetProcess, fmt.Errorf("error:Can not find any proess.")
}

func StartProcessByPassUAC(applicationCmd string) error {
	winlogonEntry, err := GetProcessByName("winlogon.exe")
	if err != nil {
		return err
	}
	// 获取winlogon 进程的句柄
	winlogonProcess, err := windows.OpenProcess(MAXIMUM_ALLOWED, false, winlogonEntry.PID())
	// 此处可能会返回异常,但是不用担心,只要成功获取到进程就可以
	// if err != nil { // The operation completed successfully
	//  Ilog.DebugHandler("OpenProcess:", err)
	//  return err
	// }
	defer windows.CloseHandle(winlogonProcess)

	// flags that specify the priority and creation method of the process
	dwCreationFlags := CREATE_NEW_CONSOLE | CREATE_UNICODE_ENVIRONMENT
	// func() uint32 {
	//  if visible {
	//      return CREATE_NEW_CONSOLE
	//  } else {
	//      return CREATE_NO_WINDOW
	//  }
	// }() | CREATE_UNICODE_ENVIRONMENT

	var syshUserTokenDup syscall.Token
	syscall.OpenProcessToken(syscall.Handle(winlogonProcess), MAXIMUM_ALLOWED, &syshUserTokenDup)
	defer syshUserTokenDup.Close()

	var syslpProcessInformation syscall.ProcessInformation //= &syscall.ProcessInformation{}

	var syslpStartipInfo syscall.StartupInfo = syscall.StartupInfo{
		Desktop:    windows.StringToUTF16Ptr(`WinSta0\Default`),
		ShowWindow: ISW.SW_SHOW, // func() uint16 {
		//  if visible {
		//      return ISW.SW_SHOW
		//  } else {
		//      return ISW.SW_HIDE
		//  }
		// }()
	}
	syslpStartipInfo.Cb = uint32(unsafe.Sizeof(syslpStartipInfo))

	var syslpProcessAttributes *syscall.SecurityAttributes

	starterr := syscall.CreateProcessAsUser(
		syshUserTokenDup,
		nil,
		windows.StringToUTF16Ptr(applicationCmd),
		syslpProcessAttributes,
		syslpProcessAttributes,
		false,
		uint32(dwCreationFlags),
		nil,
		nil,
		&syslpStartipInfo,
		&syslpProcessInformation)
	return starterr
}
