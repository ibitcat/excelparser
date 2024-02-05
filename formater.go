package main

type iFormater interface {
	formatRows()
}

func NewFormater(x *Xlsx, format, mode string) iFormater {
	switch format {
	case "lua":
		return &LuaFormater{Xlsx: x, mode: mode}
	case "json":
		return &JsonFormater{Xlsx: x, mode: mode}
	default:
		return nil
	}
}
