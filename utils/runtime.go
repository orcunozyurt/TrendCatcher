package utils

import (
	"fmt"
	"runtime"
)

// ConfigRuntime sets GOMAXPROCS to number of CPUs
func ConfigRuntime() {
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	fmt.Printf("Running with %d CPUs\n", nuCPU)
}
