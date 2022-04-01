package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type LuaFormater struct {
	*Xlsx
	mode string
}

// 类型检查(例如: int 类型的字段填了 string， 耗性能)
// 高级特性：id公式，数值范围检查，字段注释，配置行注释
func (l *LuaFormater) formatRows() {
	// 复用 datas
	l.clearData()

	// comment
	l.formatComments(l.RootField)

	// data
	l.appendData("\nreturn {\n")
	if l.Vertical {
		l.exportRow(l.RootField, 4, -1)
	} else {
		for line := 4; line < len(l.Rows); line++ {
			key := l.Rows[line][0]
			if strings.HasPrefix(key, "//") || key == "" {
				continue
			}
			l.exportRow(l.RootField, line, -1)
		}
		l.trimData("\n")
	}
	l.appendData("}\n")
	l.exportToFile()
}

/// comments
func (l *LuaFormater) formatComments(f *FieldInfo) {
	var idx int
	for _, field := range f.Fields {
		if field.isHitMode(l.mode) {
			l.formatComment(idx, field)
			idx++
		}
	}
}

func (l *LuaFormater) formatComment(idx int, f *FieldInfo) {
	var keyName string
	if f.Parent.IsArray {
		keyName = getIndent(f.Deepth) + "[" + strconv.Itoa(idx+1) + "]"
	} else {
		keyName = getIndent(f.Deepth) + f.Name
	}
	l.appendData(fmt.Sprintf("-- %-30s %-10s %s\n", keyName, f.RawType, f.Desc))

	// recursive
	if len(f.Fields) > 0 {
		l.formatComments(f)
	}
}

/// datas
func (l *LuaFormater) formatChildRow(f *FieldInfo, line int) {
	var idx int
	for _, field := range f.Fields {
		if field.isHitMode(l.mode) {
			l.exportRow(field, line, idx)
			idx++
		}
	}
	l.judgCompressTrim("", "\n")
}

func (l *LuaFormater) exportRow(f *FieldInfo, line, index int) {
	deepth := f.Deepth + 1
	indent := getIndent(deepth)

	row := l.Rows[line]
	if f.Index == -1 {
		if l.Vertical {
			l.formatChildRow(f, line)
		} else {
			// root, eg.: [1001] = {
			l.appendData(indent)
			l.appendData("[")
			l.appendData(row[0])
			l.appendData("] = {")
			l.judgCompressAppend("", "\n")
			l.formatChildRow(f, line)
			l.judgCompressAppend("", indent)
			l.appendData("}")
			l.appendData(",\n")
		}
	} else {
		var val string
		ok, val := f.getValue(row)
		if !ok {
			return
		}

		if f.Type == "json" {
			// json 格式化
			if err := l.formatJson(f, index+1, val); err != nil {
				l.sprintfError("[%s] json 格式错误：(行%d,列%d)[%s@%s]", l.mode, line+1, f.Index+1, f.Name, f.Desc)
			}
		} else {
			l.judgCompressAppend("", indent)
			if f.Parent.IsArray {
				l.appendData("[")
				l.appendData(strconv.Itoa(index + 1))
				l.appendData("] = ")
			} else {
				l.appendData(f.Name)
				l.appendData(" = ")
			}
			if len(f.Fields) > 0 {
				l.appendData("{")
				l.judgCompressAppend("", "\n")
				l.formatChildRow(f, line)
				l.judgCompressAppend("", indent)
				l.appendData("}")
			} else {
				l.appendData(val)
			}
			l.judgCompressAppend(",", ",\n")
		}
	}
}

/// json
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

func (l *LuaFormater) formatJsonValue(key interface{}, obj interface{}, deepth int) {
	indent := getIndent(deepth + 1)
	l.judgCompressAppend("", indent)
	l.appendData(l.formatJsonKey(key))
	l.appendData(" = ")

	switch obj := obj.(type) {
	case map[interface{}]interface{}:
		l.appendData("{")
		l.judgCompressAppend("", "\n")
		for k, v := range obj {
			l.formatJsonValue(k, v, deepth+1)
		}
		l.judgCompressTrim("", "\n")
		l.judgCompressAppend("", indent)
		l.appendData("}")
		l.judgCompressAppend(",", ",\n")
	case map[string]interface{}:
		l.appendData("{")
		l.judgCompressAppend("", "\n")
		for k, v := range obj {
			l.formatJsonValue(k, v, deepth+1)
		}
		l.judgCompressTrim("", "\n")
		l.judgCompressAppend("", indent)
		l.appendData("}")
		l.judgCompressAppend(",", ",\n")
	case []interface{}:
		l.appendData("{")
		l.judgCompressAppend("", "\n")
		for i, v := range obj {
			l.formatJsonValue(i+1, v, deepth+1)
		}
		l.judgCompressTrim("", "\n")
		l.judgCompressAppend("", indent)
		l.appendData("}")
		l.judgCompressAppend(",", ",\n")
	case string:
		l.appendData(formatString(obj))
		l.judgCompressAppend(",", ",\n")
	default:
		l.appendData(fmt.Sprintf("%v", obj))
		l.judgCompressAppend(",", ",\n")
	}
}

// https://github.com/ChimeraCoder/gojson/blob/master/json-to-struct.go
func (l *LuaFormater) formatJson(f *FieldInfo, index int, jsonStr string) error {
	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return err
	}

	var key interface{}
	if f.Parent.IsArray {
		key = index
	} else {
		key = f.Name
	}
	l.formatJsonValue(key, result, f.Deepth)
	return nil
}

/// export
func (l *LuaFormater) exportToFile() {
	var outpath string
	if l.mode == "c" {
		outpath = FlagClient.OutPath
	} else if l.mode == "s" {
		outpath = FlagServer.OutPath
	}
	fileName := fmt.Sprintf("%s/%s.lua", outpath, l.FileName)
	l.writeToFile(fileName)
}
