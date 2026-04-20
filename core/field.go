package core

import (
	"strconv"
)

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
	errStr := "配置字段值错误"
	switch f.Kind {
	case TArray:
		for _, v := range f.Vals {
			if !v.checkRow(row, line, x) {
				errStr = "数组元素配置值错误"
				ok = false
			}
		}
	case TMap:
		for _, k := range f.Keys {
			if !k.checkRow(row, line, x) {
				errStr = "映射键配置值错误,key: " + k.Name
				ok = false
			}
		}
		for _, v := range f.Vals {
			if !v.checkRow(row, line, x) {
				errStr = "映射值配置值错误,val: " + v.Name
				ok = false
			}
		}
	case TStruct:
		for _, v := range f.Vals {
			if !v.checkRow(row, line, x) {
				errStr = "结构体字段配置值错误,field: " + v.Name
				ok = false
			}
		}
	case TJson:
		if f.Vtype != nil && len(val) > 0 {
			ok = f.Vtype.checkJsonVal(val)
		}
	case TUint:
		if len(val) > 0 {
			_, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				errStr = "无效的无符号整数值: " + err.Error()
				ok = false
			}
		}
	case TInt:
		if len(val) > 0 {
			_, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				errStr = "无效的整数值: " + err.Error()
				ok = false
			}
		}
	case TFloat:
		if len(val) > 0 {
			_, err := strconv.ParseFloat(val, 64)
			if err != nil {
				errStr = "无效的浮点数值: " + err.Error()
				ok = false
			}
		}
	case TBool:
		if len(val) > 0 {
			if !(val == "0" || val == "1" || val == "true" || val == "false") {
				errStr = "无效的布尔值: " + val
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
		x.sprintfCellError(line, f.Index+1, errStr)
	}
	return ok
}
