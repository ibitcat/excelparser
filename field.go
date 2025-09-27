package main

import (
	"strconv"
)

type Field struct {
	*Type           // 字段数据类型
	Parent *Field   // 父字段
	Xlsx   *Xlsx    // 所属excel
	Index  int      // 字段索引
	Desc   string   // 字段描述
	Rname  string   // 原始字段名
	Name   string   // 字段名
	Mode   string   // 生成方式(s=server,c=client,x=none)
	Keys   []*Field // 键元素列表
	Vals   []*Field // 值元素列表
}

// methods
func (f *Field) isHitMode(tMode string) bool {
	if len(f.Mode) == 0 {
		return true
	}
	if tMode == "server" && f.Mode == "s" {
		return true
	}
	if tMode == "client" && f.Mode == "c" {
		return true
	}
	return false
}

func (f *Field) isVaildMode() bool {
	switch f.Mode {
	case "", "x", "s", "c":
		return true
	default:
		return false
	}
}

func (f *Field) checkRow(row []string, line int, x *Xlsx) bool {
	var val string
	if f.Index >= 0 && len(row) > f.Index {
		val = row[f.Index]
	}

	ok := true
	switch f.Kind {
	case TArray:
		for _, v := range f.Vals {
			if !v.checkRow(row, line, x) {
				ok = false
			}
		}
	case TMap:
		for _, k := range f.Keys {
			if !k.checkRow(row, line, x) {
				ok = false
			}
		}
		for _, v := range f.Vals {
			if !v.checkRow(row, line, x) {
				ok = false
			}
		}
	case TStruct:
		for _, v := range f.Vals {
			if !v.checkRow(row, line, x) {
				ok = false
			}
		}
	case TJson:
		if f.Vtype != nil {
			ok = f.Vtype.checkJsonVal(val)
		}
	case TUint:
		if len(val) > 0 && !FlagDefault {
			_, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				ok = false
			}
		}
	case TInt:
		if len(val) > 0 && !FlagDefault {
			_, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				ok = false
			}
		}
	case TFloat:
		if len(val) > 0 && !FlagDefault {
			_, err := strconv.ParseFloat(val, 64)
			if err != nil {
				ok = false
			}
		}
	case TBool:
		if len(val) > 0 && !FlagDefault {
			if !(val == "0" || val == "1" || val == "true" || val == "false") {
				ok = false
			}
		}
	case TString:
		if ok && f.isI18nString() && len(val) > 0 {
			i18nStr := getI18nString(val, f, line)
			if len(i18nStr) > 0 {
				row[f.Index] = i18nStr
			}
		}
	}

	if !ok && (f.isBuiltin() || f.Kind == TJson) {
		x.sprintfCellError(line, f.Index+1, "配置值错误")
	}
	return ok
}
