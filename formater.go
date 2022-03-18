package main

type iFormater interface {
	formatRows()
	exportToFile()
}

func NewFormater(x *Xlsx, lang, mode string) iFormater {
	switch lang {
	case "lua":
		return &LuaFormater{Xlsx: x, mode: mode}
	case "json":
		return &JsonFormater{Xlsx: x, mode: mode}
	default:
		return nil
	}
}
