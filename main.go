//go:build !gui

package main

import (
	"excelparser/core"
	"flag"
	"fmt"
	"strings"
)

func main() {
	flag.Parse()
	if core.Flaghelp || flag.NFlag() <= 0 {
		flag.Usage()
		return
	}

	if core.FlagFiles != "" {
		for _, f := range strings.Split(core.FlagFiles, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				core.GFlags.Files = append(core.GFlags.Files, f)
			}
		}
	}

	core.Run(nil)
	fmt.Printf("Total Cost: %d ms\n", core.ExportCost)
	// fmt.Printf("running goroutines: %d\n", p.Running())
}
