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
	for line := 4; line < len(l.Rows); line++ {
		row := l.Rows[line]
		if strings.HasPrefix(row[0], "//") || row[0] == "" {
			continue
		}
		l.exportRow(l.RootField, row, -1)
	}
	l.trimData("}\n")
	l.appendData("}\n")
	l.exportToFile()
}

/// comments
func (l *LuaFormater) formatComments(f *FieldInfo) {
	var idx int
	for _, field := range f.Fields {
		if field.Mode == l.mode || field.Mode == "b" {
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
func (l *LuaFormater) formatChildRow(f *FieldInfo, row []string) {
	var idx int
	for _, field := range f.Fields {
		if field.Mode == l.mode || field.Mode == "b" {
			l.exportRow(field, row, idx)
			idx++
		}
	}
	l.trimData("\n")
}

func (l *LuaFormater) exportRow(f *FieldInfo, row []string, index int) {
	deepth := f.Deepth + 1
	indent := getIndent(deepth)

	if f.Index == -1 {
		// root, eg.: [1001] = {
		l.appendData(indent)
		l.appendData("[")
		l.appendData(row[0])
		l.appendData("] = {\n")
		l.formatChildRow(f, row)
		l.appendData(indent)
		l.appendData("},\n")
	} else {
		if f.Type == "json" {
			// json 格式化
			val := formatValue(f, row[f.Index])
			l.formatJson(f, index+1, val)
		} else {
			l.appendData(indent)
			if f.Parent.IsArray {
				l.appendData("[")
				l.appendData(strconv.Itoa(index + 1))
				l.appendData("] = ")
			} else {
				l.appendData(f.Name)
				l.appendData(" = ")
			}
			if len(f.Fields) > 0 {
				l.appendData("{\n")
				l.formatChildRow(f, row)
				l.appendData(indent)
				l.appendData("}")
				l.appendData(",\n")
			} else {
				val := formatValue(f, row[f.Index])
				l.appendData(val)
				l.appendData(",\n")
			}
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
	l.appendData(indent)
	l.appendData(l.formatJsonKey(key))
	l.appendData(" = ")

	switch obj := obj.(type) {
	case map[interface{}]interface{}:
		l.appendData("{\n")
		for k, v := range obj {
			l.formatJsonValue(k, v, deepth+1)
		}
		l.trimData("\n")
		l.appendData(indent)
		l.appendData("}")
		l.appendData(",\n")
	case map[string]interface{}:
		l.appendData("{\n")
		for k, v := range obj {
			l.formatJsonValue(k, v, deepth+1)
		}
		l.trimData("\n")
		l.appendData(indent)
		l.appendData("}")
		l.appendData(",\n")
	case []interface{}:
		l.appendData("{\n")
		for i, v := range obj {
			l.formatJsonValue(i+1, v, deepth+1)
		}
		l.trimData("\n")
		l.appendData(indent)
		l.appendData("}")
		l.appendData(",\n")
	case string:
		l.appendData(formatString(obj))
		l.appendData(",\n")
	default:
		l.appendData(fmt.Sprintf("%v", obj))
		l.appendData(",\n")
	}
}

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
