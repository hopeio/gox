/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package pprof

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

// go tool pprof -http=:8080 ./cpu.pprof
func StartCPUProfile(filename string) func() {
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

func WriteHeapProfile(filename string) {
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

func WriteProfile(filename, pname string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("could not create file: ", err)
	}

	if err := pprof.Lookup(pname).WriteTo(f, 1); err != nil {
		log.Fatal("could not write: ", err)
	}
	f.Close()
}
