package runtime

import (
	"log"
	"runtime"
)

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
