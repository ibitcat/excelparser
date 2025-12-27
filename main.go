//go:build !gui

package main

import (
	"excelparser/core"
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	if core.Flaghelp || flag.NFlag() <= 0 {
		flag.Usage()
		return
	}

	core.Run(nil)
	fmt.Printf("Total Cost: %d ms\n", core.ExportCost)
	// fmt.Printf("running goroutines: %d\n", p.Running())
}
