package ui

import "github.com/ying32/govcl/vcl"

//::private::
type TForm2Fields struct {
	// TODO
}

func (f *TForm2) OnFormCreate(sender vcl.IObject) {
	f.Label1.SetCaption("这是一个奇怪的工具\n\rbitcat © 2025")
}
