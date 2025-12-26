//go:build gui

package main

import (
	"excelparser/ui"

	_ "github.com/ying32/govcl/pkgs/winappres"
	"github.com/ying32/govcl/vcl"
)

func main() {
	vcl.Application.Initialize()
	vcl.Application.SetMainFormOnTaskBar(true)

	f1 := ui.NewForm1(vcl.Application)
	vcl.Application.CreateForm(&f1)
	vcl.Application.Run()
}
