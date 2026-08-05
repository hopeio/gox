package runtime

import (
	"log"
	"runtime"
)

// PrintMemoryUsage performs the operation.
func PrintMemoryUsage(flag any) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("%v TotalAlloc = %.2f MiB,HeapAlloc = %.2f MiB,Sys = %.2f MiB,HeapSys = %.2f MiB,StackSys = %.2f MiB,HeapInuse = %.2f MiB,StackInuse = %.2f MiB,Mallocs = %d,Frees = %d,NumGC = %d", flag, bToMb(m.TotalAlloc), bToMb(m.HeapAlloc), bToMb(m.Sys), bToMb(m.HeapSys), bToMb(m.StackSys), bToMb(m.HeapInuse), bToMb(m.StackInuse), m.Mallocs, m.Frees, m.NumGC)
}

// PrintStack performs the operation.
func PrintStack(flag any) {
	// Create a 1MB buffer for stack traces
	buf := make([]byte, 1<<20) // 1MB buffer
	// Capture the current goroutine stack
	stackLen := runtime.Stack(buf, false)
	log.Printf("%v Stack:\n%s", flag, buf[:stackLen])
}

// bToMb returns the result.
func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}
