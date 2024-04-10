package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type LuaFormater struct {
	*Xlsx
	line int
	mode string
}

// 类型检查(例如: int 类型的字段填了 string， 耗性能)
// 高级特性：id公式，数值范围检查，字段注释，配置行注释
func (l *LuaFormater) formatRows() {
	// 复用 datas
	l.line = 0
	l.clearData()

	// data
	if l.Vertical {
		l.appendData("return ")
		for _, col := range l.Rows {
			l.line++
			l.formatData(l.RootField, col, 0)
		}
	} else {
		l.appendData("return {\n")
		for _, row := range l.Rows {
			l.line++
			key := row[0]
			if strings.HasPrefix(key, "//") || key == "" {
				continue
			}
			l.appendIndent(1)
			l.appendData("[")
			l.appendData(row[0])
			l.appendData("]")
			l.appendSpace()
			l.appendData("=")
			l.appendSpace()
			l.formatData(l.RootField, row, 1)
			l.appendData(",\n")
		}
		l.replaceTail("\n")
		l.appendData("}")
	}
}

func (l *LuaFormater) formatData(field *Field, row []string, depth int) {
	fkind := field.Kind
	switch fkind {
	case TArray:
		l.appendData("{")
		l.appendEOL()
		for i, f := range field.Vals {
			l.appendIndent(depth + 1)
			l.appendData("[")
			l.appendData(strconv.Itoa(i + 1))
			l.appendData("]")
			l.appendSpace()
			l.appendData("=")
			l.appendSpace()
			l.formatData(f, row, depth+1)
			l.appendComma()
		}
		l.replaceComma()
		l.appendIndent(depth)
		l.appendData("}")
	case TMap:
		l.appendData("{")
		l.appendEOL()
		for i, k := range field.Keys {
			l.appendIndent(depth + 1)
			if k.isNumber() {
				l.appendData("[")
				l.appendData(row[k.Index])
				l.appendData("]")
				l.appendSpace()
				l.appendData("=")
				l.appendSpace()
			} else {
				l.appendData(row[k.Index])
				l.appendSpace()
				l.appendData("=")
				l.appendSpace()
			}

			v := field.Vals[i]
			l.formatData(v, row, depth+1)
			l.appendComma()
		}
		l.replaceComma()
		l.appendIndent(depth)
		l.appendData("}")
	case TStruct:
		l.appendData("{")
		l.appendEOL()
		for _, f := range field.Vals {
			if f.isHitMode(l.mode) {
				l.appendIndent(depth + 1)
				l.appendData(f.Name)
				l.appendSpace()
				l.appendData("=")
				l.appendSpace()
				l.formatData(f, row, depth+1)
				l.appendComma()
			}
		}
		l.replaceComma()
		l.appendIndent(depth)
		l.appendData("}")
	case TJson:
		s := ""
		if len(row) > field.Index {
			s = row[field.Index]
		}

		// https://github.com/ChimeraCoder/gojson/blob/master/json-to-struct.go
		var result interface{}
		err := json.Unmarshal([]byte(s), &result)
		if err == nil {
			l.formatJsonValue(field, field.Vtype, result, depth)
		} else {
			l.appendData("nil")
		}
	default:
		s := ""
		if len(row) > field.Index {
			s = row[field.Index]
		}
		l.appendData(field.formatValue(s))
	}
}

// json
func (l *LuaFormater) formatJsonKey(key interface{}) string {
	var keystr string
	switch key.(type) {
	case int, uint:
		keystr = fmt.Sprintf("[%v]", key)
	default:
		keystr = fmt.Sprintf("%v", key)
	}
	return keystr
}

func (l *LuaFormater) formatJsonValue(field *Field, t *Type, obj interface{}, depth int) {
	var vt *Type
	if t != nil {
		vt = t.Vtype
	}
	switch val := obj.(type) {
	case map[interface{}]interface{}:
		l.appendData("{")
		l.appendEOL()
		for k, v := range val {
			l.appendIndent(depth + 1)
			l.appendData(l.formatJsonKey(k))
			l.appendSpace()
			l.appendData("=")
			l.appendSpace()
			l.formatJsonValue(field, vt, v, depth+1)
			l.appendComma()
		}
		l.replaceComma()
		l.appendIndent(depth)
		l.appendData("}")
	case map[string]interface{}:
		l.appendData("{")
		l.appendEOL()
		for k, v := range val {
			l.appendIndent(depth + 1)
			l.appendData(l.formatJsonKey(k))
			l.appendSpace()
			l.appendData("=")
			l.appendSpace()
			l.formatJsonValue(field, vt, v, depth+1)
			l.appendComma()
		}
		l.replaceComma()
		l.appendIndent(depth)
		l.appendData("}")
	case []interface{}:
		l.appendData("{")
		l.appendEOL()
		for i, v := range val {
			l.appendIndent(depth + 1)
			l.appendData(l.formatJsonKey(i + 1))
			l.appendSpace()
			l.appendData("=")
			l.appendSpace()
			l.formatJsonValue(field, vt, v, depth+1)
			l.appendComma()
		}
		l.replaceComma()
		l.appendIndent(depth)
		l.appendData("}")
	case string:
		if t != nil && t.isI18nString() {
			val = getI18nString(val, field, l.line+HeadLineNum)
		}
		l.appendData(formatString(val))
	default:
		l.appendData(fmt.Sprintf("%v", val))
	}
}
