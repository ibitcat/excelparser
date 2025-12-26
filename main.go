//go:build !gui

package main

import (
	"excelparser/core"
	"flag"
	"fmt"
	"time"
)

func main() {
	flag.Parse()
	if core.Flaghelp || flag.NFlag() <= 0 {
		flag.Usage()
		return
	}

	startTime := time.Now()
	core.Run()
	fmt.Printf("Total Cost: %d ms\n", core.GetDurationMs(startTime))
	// fmt.Printf("running goroutines: %d\n", p.Running())
}
