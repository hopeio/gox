/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package debug

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

// go tool pprof -http=:8080 ./cpu.pprof
func PprofCPU(filename string) func() {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("could not create file: ", err)
	}
	// StartCPUProfile为当前进程开启CPU profile。
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	// StopCPUProfile会停止当前的CPU profile（如果有）
	return func() {
		pprof.StopCPUProfile()
		f.Close()
	}
}

func PprofHeap(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("could not create file: ", err)
	}
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write heap profile: ", err)
	}
	f.Close()
}

func PprofByName(filename, pname string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("could not create file: ", err)
	}

	if err := pprof.Lookup(pname).WriteTo(f, 1); err != nil {
		log.Fatal("could not write: ", err)
	}
	f.Close()
}

func PrintMemoryUsage(flag any) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("%v TotalAlloc = %.2f MiB,HeapAlloc = %.2f MiB,Sys = %.2f MiB,HeapSys = %.2f MiB,StackSys = %.2f MiB,HeapInuse = %.2f MiB,StackInuse = %.2f MiB,Mallocs = %d,Frees = %d,NumGC = %d", flag, bToMb(m.TotalAlloc), bToMb(m.HeapAlloc), bToMb(m.Sys), bToMb(m.HeapSys), bToMb(m.StackSys), bToMb(m.HeapInuse), bToMb(m.StackInuse), m.Mallocs, m.Frees, m.NumGC)
}

func PrintStack(flag any) {
	// 创建一个 1MB 的缓冲区来存储堆栈信息
	buf := make([]byte, 1<<20) // 1MB 缓冲区
	// 获取当前 Goroutine 的堆栈信息
	stackLen := runtime.Stack(buf, false)
	log.Printf("%v Stack:\n%s", flag, buf[:stackLen])
}

func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}
