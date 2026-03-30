// Copyright 2010-2012 The W32 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows
package win

import (
	"github.com/gonutz/w32/v2"
	"syscall"
	"unsafe"
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetModuleHandle    = modkernel32.NewProc("GetModuleHandleW")
	procMulDiv             = modkernel32.NewProc("MulDiv")
	procGetConsoleWindow   = modkernel32.NewProc("GetConsoleWindow")
	procGetCurrentThread   = modkernel32.NewProc("GetCurrentThread")
	procGetLogicalDrives   = modkernel32.NewProc("GetLogicalDrives")
	procGetUserDefaultLCID = modkernel32.NewProc("GetUserDefaultLCID")
	procLstrlen            = modkernel32.NewProc("lstrlenW")
	procLstrcpy            = modkernel32.NewProc("lstrcpyW")
	procGlobalAlloc        = modkernel32.NewProc("GlobalAlloc")
	procGlobalFree         = modkernel32.NewProc("GlobalFree")
	procGlobalLock         = modkernel32.NewProc("GlobalLock")
	procGlobalUnlock       = modkernel32.NewProc("GlobalUnlock")
	procMoveMemory         = modkernel32.NewProc("RtlMoveMemory")
	procFindResource       = modkernel32.NewProc("FindResourceW")
	procSizeofResource     = modkernel32.NewProc("SizeofResource")
	procLockResource       = modkernel32.NewProc("LockResource")
	procLoadResource       = modkernel32.NewProc("LoadResource")
	procGetLastError       = modkernel32.NewProc("GetLastError")
	//procOpenProcess                = modkernel32.NewProc("OpenProcess")
	procTerminateProcess           = modkernel32.NewProc("TerminateProcess")
	procCloseHandle                = modkernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot   = modkernel32.NewProc("CreateToolhelp32Snapshot")
	procModule32First              = modkernel32.NewProc("Module32FirstW")
	procModule32Next               = modkernel32.NewProc("Module32NextW")
	procGetSystemTimes             = modkernel32.NewProc("GetSystemTimes")
	procGetConsoleScreenBufferInfo = modkernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleTextAttribute    = modkernel32.NewProc("SetConsoleTextAttribute")
	procGetDiskFreeSpaceEx         = modkernel32.NewProc("GetDiskFreeSpaceExW")
	procGetProcessTimes            = modkernel32.NewProc("GetProcessTimes")
	procSetSystemTime              = modkernel32.NewProc("SetSystemTime")
	procGetSystemTime              = modkernel32.NewProc("GetSystemTime")
	procVirtualAllocEx             = modkernel32.NewProc("VirtualAllocEx")
	procVirtualFreeEx              = modkernel32.NewProc("VirtualFreeEx")
	procWriteProcessMemory         = modkernel32.NewProc("WriteProcessMemory")
	procReadProcessMemory          = modkernel32.NewProc("ReadProcessMemory")
	procQueryPerformanceCounter    = modkernel32.NewProc("QueryPerformanceCounter")
	procQueryPerformanceFrequency  = modkernel32.NewProc("QueryPerformanceFrequency")
	process32Next                  = modkernel32.NewProc("Process32Next")

	procQueryFullProcessImageNameW = modkernel32.NewProc("QueryFullProcessImageNameW")

	procQueryDosDeviceW = modkernel32.NewProc("QueryDosDeviceW")
)

func GetModuleHandle(modulename string) w32.HINSTANCE {
	var mn uintptr
	if modulename == "" {
		mn = 0
	} else {
		mn = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(modulename)))
	}
	ret, _, _ := procGetModuleHandle.Call(mn)
	return w32.HINSTANCE(ret)
}

func MulDiv(number, numerator, denominator int) int {
	ret, _, _ := procMulDiv.Call(
		uintptr(number),
		uintptr(numerator),
		uintptr(denominator))

	return int(ret)
}

func GetConsoleWindow() w32.HWND {
	ret, _, _ := procGetConsoleWindow.Call()

	return w32.HWND(ret)
}

func GetCurrentThread() w32.HANDLE {
	ret, _, _ := procGetCurrentThread.Call()

	return w32.HANDLE(ret)
}

func GetLogicalDrives() uint32 {
	ret, _, _ := procGetLogicalDrives.Call()

	return uint32(ret)
}

func GetUserDefaultLCID() uint32 {
	ret, _, _ := procGetUserDefaultLCID.Call()

	return uint32(ret)
}

func Lstrlen(lpString *uint16) int {
	ret, _, _ := procLstrlen.Call(uintptr(unsafe.Pointer(lpString)))

	return int(ret)
}

func Lstrcpy(buf []uint16, lpString *uint16) {
	procLstrcpy.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(lpString)))
}

func GlobalAlloc(uFlags uint, dwBytes uint32) w32.HGLOBAL {
	ret, _, _ := procGlobalAlloc.Call(
		uintptr(uFlags),
		uintptr(dwBytes))

	if ret == 0 {
		panic("GlobalAlloc failed")
	}

	return w32.HGLOBAL(ret)
}

func GlobalFree(hMem w32.HGLOBAL) {
	ret, _, _ := procGlobalFree.Call(uintptr(hMem))

	if ret != 0 {
		panic("GlobalFree failed")
	}
}

func GlobalLock(hMem w32.HGLOBAL) unsafe.Pointer {
	ret, _, _ := procGlobalLock.Call(uintptr(hMem))

	if ret == 0 {
		panic("GlobalLock failed")
	}

	return unsafe.Pointer(ret)
}

func GlobalUnlock(hMem w32.HGLOBAL) bool {
	ret, _, _ := procGlobalUnlock.Call(uintptr(hMem))

	return ret != 0
}

func MoveMemory(destination, source unsafe.Pointer, length uint32) {
	procMoveMemory.Call(
		uintptr(unsafe.Pointer(destination)),
		uintptr(source),
		uintptr(length))
}

func FindResource(hModule w32.HMODULE, lpName, lpType *uint16) (w32.HRSRC, error) {
	ret, _, _ := procFindResource.Call(
		uintptr(hModule),
		uintptr(unsafe.Pointer(lpName)),
		uintptr(unsafe.Pointer(lpType)))

	if ret == 0 {
		return 0, syscall.GetLastError()
	}

	return w32.HRSRC(ret), nil
}

func SizeofResource(hModule w32.HMODULE, hResInfo w32.HRSRC) uint32 {
	ret, _, _ := procSizeofResource.Call(
		uintptr(hModule),
		uintptr(hResInfo))

	if ret == 0 {
		panic("SizeofResource failed")
	}

	return uint32(ret)
}

func LockResource(hResData w32.HGLOBAL) unsafe.Pointer {
	ret, _, _ := procLockResource.Call(uintptr(hResData))

	if ret == 0 {
		panic("LockResource failed")
	}

	return unsafe.Pointer(ret)
}

func LoadResource(hModule w32.HMODULE, hResInfo w32.HRSRC) w32.HGLOBAL {
	ret, _, _ := procLoadResource.Call(
		uintptr(hModule),
		uintptr(hResInfo))

	if ret == 0 {
		panic("LoadResource failed")
	}

	return w32.HGLOBAL(ret)
}

func GetLastError() uint32 {
	ret, _, _ := procGetLastError.Call()
	return uint32(ret)
}

// func OpenProcess(desiredAccess uint32, inheritHandle bool, processId uint32) HANDLE {
// 	inherit := 0
// 	if inheritHandle {
// 		inherit = 1
// 	}

// 	ret, _, _ := procOpenProcess.Call(
// 		uintptr(desiredAccess),
// 		uintptr(inherit),
// 		uintptr(processId))
// 	return HANDLE(ret)
// }

// func TerminateProcess(hProcess HANDLE, uExitCode uint) bool {
// 	ret, _, _ := procTerminateProcess.Call(
// 		uintptr(hProcess),
// 		uintptr(uExitCode))
// 	return ret != 0
// }

func CreateToolhelp32Snapshot(flags, processId uint32) w32.HANDLE {
	ret, _, _ := procCreateToolhelp32Snapshot.Call(
		uintptr(flags),
		uintptr(processId))

	if ret <= 0 {
		return w32.HANDLE(0)
	}

	return w32.HANDLE(ret)
}

func Module32First(snapshot w32.HANDLE, me *w32.MODULEENTRY32) bool {
	ret, _, _ := procModule32First.Call(
		uintptr(snapshot),
		uintptr(unsafe.Pointer(me)))

	return ret != 0
}

func Module32Next(snapshot w32.HANDLE, me *w32.MODULEENTRY32) bool {
	ret, _, _ := procModule32Next.Call(
		uintptr(snapshot),
		uintptr(unsafe.Pointer(me)))

	return ret != 0
}

func GetSystemTimes(lpIdleTime, lpKernelTime, lpUserTime *w32.FILETIME) bool {
	ret, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(lpIdleTime)),
		uintptr(unsafe.Pointer(lpKernelTime)),
		uintptr(unsafe.Pointer(lpUserTime)))

	return ret != 0
}

func GetProcessTimes(hProcess w32.HANDLE, lpCreationTime, lpExitTime, lpKernelTime, lpUserTime *w32.FILETIME) bool {
	ret, _, _ := procGetProcessTimes.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(lpCreationTime)),
		uintptr(unsafe.Pointer(lpExitTime)),
		uintptr(unsafe.Pointer(lpKernelTime)),
		uintptr(unsafe.Pointer(lpUserTime)))

	return ret != 0
}

func GetConsoleScreenBufferInfo(hConsoleOutput w32.HANDLE) *w32.CONSOLE_SCREEN_BUFFER_INFO {
	var csbi w32.CONSOLE_SCREEN_BUFFER_INFO
	ret, _, _ := procGetConsoleScreenBufferInfo.Call(
		uintptr(hConsoleOutput),
		uintptr(unsafe.Pointer(&csbi)))
	if ret == 0 {
		return nil
	}
	return &csbi
}

func SetConsoleTextAttribute(hConsoleOutput w32.HANDLE, wAttributes uint16) bool {
	ret, _, _ := procSetConsoleTextAttribute.Call(
		uintptr(hConsoleOutput),
		uintptr(wAttributes))
	return ret != 0
}

func GetDiskFreeSpaceEx(dirName string) (r bool,
	freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64) {
	ret, _, _ := procGetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(dirName))),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)))
	return ret != 0,
		freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes
}

func GetSystemTime() *w32.SYSTEMTIME {
	var time w32.SYSTEMTIME
	procGetSystemTime.Call(
		uintptr(unsafe.Pointer(&time)))
	return &time
}

func SetSystemTime(time *w32.SYSTEMTIME) bool {
	ret, _, _ := procSetSystemTime.Call(
		uintptr(unsafe.Pointer(time)))
	return ret != 0
}

func VirtualAllocEx(hProcess w32.HANDLE, lpAddress, dwSize uintptr, flAllocationType, flProtect uint32) uintptr {
	ret, _, _ := procVirtualAllocEx.Call(
		uintptr(hProcess),
		lpAddress,
		dwSize,
		uintptr(flAllocationType),
		uintptr(flProtect),
	)

	return ret
}

func VirtualFreeEx(hProcess w32.HANDLE, lpAddress, dwSize uintptr, dwFreeType uint32) bool {
	ret, _, _ := procVirtualFreeEx.Call(
		uintptr(hProcess),
		lpAddress,
		dwSize,
		uintptr(dwFreeType),
	)

	return ret != 0
}

func WriteProcessMemory(hProcess w32.HANDLE, lpBaseAddress, lpBuffer, nSize uintptr) (int, bool) {
	var nBytesWritten int
	ret, _, _ := procWriteProcessMemory.Call(
		uintptr(hProcess),
		lpBaseAddress,
		lpBuffer,
		nSize,
		uintptr(unsafe.Pointer(&nBytesWritten)),
	)

	return nBytesWritten, ret != 0
}

func ReadProcessMemory(hProcess w32.HANDLE, lpBaseAddress, nSize uintptr) (lpBuffer []uint16, lpNumberOfBytesRead int, ok bool) {

	var nBytesRead int
	buf := make([]uint16, nSize)
	ret, _, _ := procReadProcessMemory.Call(
		uintptr(hProcess),
		lpBaseAddress,
		uintptr(unsafe.Pointer(&buf[0])),
		nSize,
		uintptr(unsafe.Pointer(&nBytesRead)),
	)

	return buf, nBytesRead, ret != 0
}

func QueryPerformanceCounter() uint64 {
	result := uint64(0)
	procQueryPerformanceCounter.Call(
		uintptr(unsafe.Pointer(&result)),
	)

	return result
}

func QueryPerformanceFrequency() uint64 {
	result := uint64(0)
	procQueryPerformanceFrequency.Call(
		uintptr(unsafe.Pointer(&result)),
	)

	return result
}

func Process32Next(pHandle w32.HANDLE, proc uintptr) bool {
	rt, _, _ := process32Next.Call(uintptr(pHandle), proc)
	return rt == 1
}
