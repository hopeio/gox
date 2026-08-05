/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package shm

import (
	"encoding/binary"
	"errors"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

const (
	FILE_MAP_ALL_ACCESS = 0xF001F
	PAGE_READWRITE      = 0x04
)

var (
	kernel32        = windows.NewLazySystemDLL("kernel32.dll")
	openFileMapping = kernel32.NewProc("OpenFileMappingW")
)

type SharedMemory struct {
	name    string
	handle  uintptr
	addr    uintptr
	size    int
	content []byte
}

// New ...
func New(name string, size int) (*SharedMemory, error) {
	// init shm
	shm, addr, err := OpenShm(name, size)
	if err != nil {
		return nil, err
	}
	content := unsafe.Slice((*byte)(unsafe.Pointer(addr)), size)
	return &SharedMemory{
		name:    name,
		handle:  shm,
		addr:    addr,
		size:    size,
		content: content,
	}, nil
}

// OpenShm creates and returns a new instance.
func OpenShm(name string, size int) (uintptr, uintptr, error) {
	shm0, _, _ := openFileMapping.Call(
		FILE_MAP_ALL_ACCESS,
		0,
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(name))),
	)
	shm := windows.Handle(shm0)
	var err error
	if shm == 0 {
		shm, err = windows.CreateFileMapping(
			windows.InvalidHandle,
			nil,
			syscall.PAGE_READWRITE,
			uint32(size>>32),
			uint32(size&0xFFFFFFFF),
			windows.StringToUTF16Ptr(name),
		)
		if err != nil {
			return 0, 0, err
		}
	}

	addr, err := windows.MapViewOfFile(
		shm,
		FILE_MAP_ALL_ACCESS,
		0,
		0,
		uintptr(size),
	)
	if err != nil {
		_ = windows.CloseHandle(shm)
		return 0, 0, err
	}

	return uintptr(shm), addr, nil
}

// ReadMemory ...
func (shm *SharedMemory) ReadMemory(begin, end int, data any) error {
	if begin > end {
		return errors.New("invalid addr")
	}
	if end > len(shm.content) {
		return errors.New("out of range")
	}
	switch v := data.(type) {
	case *string:
		*v = unsafe.String((*byte)(unsafe.Pointer(uintptr(begin))), end-begin)
	case *uint8:
		*v = shm.content[begin]
	case *uint16:
		*v = binary.LittleEndian.Uint16(shm.content[begin:end])
	case *uint32:
		*v = binary.LittleEndian.Uint32(shm.content[begin:end])
	case *uint64:
		*v = binary.LittleEndian.Uint64(shm.content[begin:end])
	case []byte:
		copy(v, shm.content[begin:end])
	case *[]byte:
		*v = shm.content[begin:end:end]
	}
	return nil
}

// WriteMemory ...
func (shm *SharedMemory) WriteMemory(begin, end int, data any) (err error) {
	if begin > end {
		return errors.New("invalid addr")
	}
	if end > len(shm.content) {
		return errors.New("out of range")
	}
	switch v := data.(type) {
	case string:
		copy(shm.content[begin:end], v)
	case *string:
		copy(shm.content[begin:end], *v)
	case uint8:
		shm.content[begin] = v
	case *uint8:
		shm.content[begin] = *v
	case uint16:
		binary.LittleEndian.PutUint16(shm.content[begin:end], v)
	case *uint16:
		binary.LittleEndian.PutUint16(shm.content[begin:end], *v)
	case uint32:
		binary.LittleEndian.PutUint32(shm.content[begin:end], v)
	case *uint32:
		binary.LittleEndian.PutUint32(shm.content[begin:end], *v)
	case uint64:
		binary.LittleEndian.PutUint64(shm.content[begin:end], v)
	case *uint64:
		binary.LittleEndian.PutUint64(shm.content[begin:end], *v)
	case uint:
		binary.LittleEndian.PutUint64(shm.content[begin:end], uint64(v))
	case *uint:
		binary.LittleEndian.PutUint64(shm.content[begin:end], uint64(*v))
	case []byte:
		copy(shm.content[begin:end], v)
	case *[]byte:
		copy(shm.content[begin:end], *v)
	}
	return nil
}

// Close closes and releases resources.
func (shm *SharedMemory) Close() error {
	err := windows.UnmapViewOfFile(shm.addr)
	if err != nil {
		return err
	}
	return windows.CloseHandle(windows.Handle(shm.handle))
}
