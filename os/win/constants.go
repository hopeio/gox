/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

 //go:build windows
package win

const (
	MEM_COMMIT     = 0x00001000
	MEM_RESERVE    = 0x00002000
	MEM_RESET      = 0x00080000
	MEM_RESET_UNDO = 0x1000000

	MEM_LARGE_PAGES = 0x20000000
	MEM_PHYSICAL    = 0x00400000
	MEM_TOP_DOWN    = 0x00100000

	MEM_DECOMMIT = 0x4000
	MEM_RELEASE  = 0x8000
)

const (
	PAGE_EXECUTE           = 0x10
	PAGE_EXECUTE_READ      = 0x20
	PAGE_EXECUTE_READWRITE = 0x40
	PAGE_EXECUTE_WRITECOPY = 0x80
	PAGE_NOACCESS          = 0x01
	PAGE_READWRITE         = 0x04
	PAGE_WRITECOPY         = 0x08
	PAGE_TARGETS_INVALID   = 0x40000000
	PAGE_TARGETS_NO_UPDATE = 0x40000000
)
