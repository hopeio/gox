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
	"syscall"
	"unsafe"
)

func ListViews(hwnd w32.HWND) []w32.HWND {
	var listViewHwnds []w32.HWND
	fnOfEnumListView := func(childHwnd w32.HWND) bool {
		className, _ := w32.GetClassName(childHwnd)

		if className == "SysListView32" {
			listViewHwnds = append(listViewHwnds, childHwnd)
		}
		return true
	}

	w32.EnumChildWindows(hwnd, fnOfEnumListView)

	return listViewHwnds
}

func GetLVItemRowCount(hwnd w32.HWND) int {
	rowCount := w32.SendMessage(hwnd, w32.LVM_GETITEMCOUNT, 0, 0)
	return int(rowCount)
}

func GetLVItem(hwnd w32.HWND, row, col int) string {

	/*	rowCount := w32.SendMessage(hwnd, w32.LVM_GETITEMCOUNT, 0, 0)
		if rowCount == 0 {
			return ""
		}

		if row-1 > int(rowCount) {
			return ""
		}*/

	_, pid := w32.GetWindowThreadProcessId(hwnd)

	hProcess := w32.OpenProcess(
		w32.PROCESS_VM_READ|w32.PROCESS_VM_WRITE|w32.PROCESS_VM_OPERATION|w32.PROCESS_QUERY_INFORMATION,
		false,
		uint32(pid),
	)

	if hProcess == 0 {
		fmt.Println("开启远程hProcess失败")
		return ""
	}

	defer w32.CloseHandle(hProcess)

	lpLvItem := VirtualAllocEx(hProcess, 0, unsafe.Sizeof(w32.LVITEM{}), MEM_COMMIT, PAGE_READWRITE)
	if lpLvItem == 0 {
		fmt.Println("申请远程内存空间失败")
		return ""
	}

	defer VirtualFreeEx(hProcess, lpLvItem, 0, MEM_RELEASE)

	lpStr := VirtualAllocEx(hProcess, 0, 512, MEM_COMMIT, PAGE_READWRITE)
	if lpStr == 0 {
		fmt.Println("申请远程内存空间失败")
		return ""
	}

	defer VirtualFreeEx(hProcess, lpStr, 0, MEM_RELEASE)

	item := &w32.LVITEM{
		Mask:       w32.LVIF_TEXT,
		IItem:      int32(row),
		ISubItem:   int32(col),
		PszText:    (*uint16)(unsafe.Pointer(lpStr)),
		CchTextMax: 512,
	}
	_, ok := WriteProcessMemory(hProcess, lpLvItem, uintptr(unsafe.Pointer(item)), unsafe.Sizeof(w32.LVITEM{}))
	if !ok {
		return ""
	}

	ret := w32.SendMessage(hwnd, w32.LVM_GETITEMTEXT, uintptr(row), lpLvItem)
	if int(ret) > 0 {
		redBuf, _, _ := ReadProcessMemory(hProcess, lpStr, ret*2)
		s := syscall.UTF16ToString(redBuf)
		return s
	}

	return ""
}

func GetList(hwnd w32.HWND, columns []int) [][]string {

	rowCount := w32.SendMessage(hwnd, w32.LVM_GETITEMCOUNT, 0, 0)
	if rowCount == 0 {
		return nil
	}

	_, pid := w32.GetWindowThreadProcessId(hwnd)

	hProcess := w32.OpenProcess(
		w32.PROCESS_VM_READ|w32.PROCESS_VM_WRITE|w32.PROCESS_VM_OPERATION|w32.PROCESS_QUERY_INFORMATION,
		false,
		uint32(pid),
	)

	if hProcess == 0 {
		fmt.Println("开启远程hProcess失败")
		return nil
	}

	defer w32.CloseHandle(hProcess)

	lpLvItem := VirtualAllocEx(hProcess, 0, unsafe.Sizeof(w32.LVITEM{}), MEM_COMMIT, PAGE_READWRITE)
	if lpLvItem == 0 {
		fmt.Println("申请远程内存空间失败")
		return nil
	}

	defer VirtualFreeEx(hProcess, lpLvItem, 0, MEM_RELEASE)

	lpStr := VirtualAllocEx(hProcess, 0, 512, MEM_COMMIT, PAGE_READWRITE)
	if lpStr == 0 {
		fmt.Println("申请远程内存空间失败")
		return nil
	}

	defer VirtualFreeEx(hProcess, lpStr, 0, MEM_RELEASE)

	item := &w32.LVITEM{
		Mask:       w32.LVIF_TEXT,
		IItem:      int32(0),
		ISubItem:   int32(0),
		PszText:    (*uint16)(unsafe.Pointer(lpStr)),
		CchTextMax: 512,
	}
	var ret [][]string
	for row := 0; row < int(rowCount); row++ {
		item.IItem = int32(row)
		var columnStrs []string
		for _, column := range columns {
			item.ISubItem = int32(column)
			_, ok := WriteProcessMemory(hProcess, lpLvItem, uintptr(unsafe.Pointer(item)), unsafe.Sizeof(w32.LVITEM{}))
			if !ok {
				return nil
			}

			ret := w32.SendMessage(hwnd, w32.LVM_GETITEMTEXT, uintptr(row), lpLvItem)
			if int(ret) > 0 {
				redBuf, _, _ := ReadProcessMemory(hProcess, lpStr, ret*2)
				s := syscall.UTF16ToString(redBuf)
				columnStrs = append(columnStrs, s)
			}
		}
		ret = append(ret, columnStrs)
	}

	return ret
}
