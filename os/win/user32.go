/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

 //go:build windows
package win

import (
	"syscall"
)

var (
	user32          = syscall.NewLazyDLL("User32.dll")
	procEnumWindows = user32.NewProc("EnumWindows")
	findWindow      = user32.NewProc("FindWindowW")
)
