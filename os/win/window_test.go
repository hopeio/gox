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
	"testing"
)

func TestWindows(t *testing.T) {
	/*	w32.EnumWindows(func(w w32.HWND) bool {
			name, _ := w32.GetClassName(w)
			fmt.Println(w32.GetWindowText(w), name, w)
			lvHwnds := ListViews(w)
			if len(lvHwnds) > 0 {
				logsCount := GetLVItemRowCount(lvHwnds[0])
				fmt.Println("count:", logsCount)
			}

			return true
		})
	*/
	mainHwnd := w32.FindWindow("#32770", "Internet Download Manager 6.36")
	lvHwnds := ListViews(mainHwnd)
	logsCount := GetLVItemRowCount(lvHwnds[0])
	fmt.Println(logsCount)
	fmt.Println(GetLVItem(lvHwnds[0], 0, 1))
	/*	processIDs, ok := w32.EnumProcesses(make([]uint32, 1024))
		if !ok {
			return
		}

		for i := 0; i < len(processIDs); i++ {
			if processIDs[i] != 0 {
				fmt.Println(GetProcName(processIDs[i]))
			}
		}*/

	/*	mainFormTitle := "任务管理器"
		mainFormClass := "TaskManagerWindow"
		hwnd := w32.FindWindow(mainFormClass, mainFormTitle)
		w32.ShowWindow(hwnd, w32.SW_NORMAL)
		w32.SetForegroundWindow(hwnd)*/
	/*var windows = DesktopWindowHWND()
	for _, w := range windows {

	}*/
}
