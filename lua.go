package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// id 冲突
// 类型检查(例如: int 类型的字段填了 string， 耗性能)
// 高级特性：id公式，数值范围检查，字段注释，配置行注释
func exportRows(x *Xlsx, mode string) {
	//cap := len(x.Types) * (len(x.Rows) - 2)
	//fmt.Println(cap, len(x.Types))
	x.Datas = x.Datas[0:0]

	// comment
	exportComments(x, x.RootField, mode)

	// data
	x.appendData("\nreturn {\n")
	for line := 4; line < len(x.Rows); line++ {
		row := x.Rows[line]
		if strings.HasPrefix(row[0], "//") || row[0] == "" {
			continue
		}
		exportRow(x, x.RootField, row, -1, mode)
	}
	x.trimData("}\n")
	x.appendData("}\n")

	writeToFile(x, mode)
}

func exportComments(x *Xlsx, f *FieldInfo, mode string) {
	var idx int
	for _, field := range f.Fields {
		if field.Mode == mode || field.Mode == "b" {
			exportComment(x, idx, field, mode)
			idx++
		}
	}
}

func exportComment(x *Xlsx, idx int, f *FieldInfo, mode string) {
	var keyName string
	if f.Parent.IsArray {
		keyName = getIndent(f.Deepth) + "[" + strconv.Itoa(idx+1) + "]"
	} else {
		keyName = getIndent(f.Deepth) + f.Name
	}
	x.appendData(fmt.Sprintf("-- %-30s %-10s %s\n", keyName, f.RawType, f.Desc))

	// recursive
	if len(f.Fields) > 0 {
		exportComments(x, f, mode)
	}
}

func exportChildRow(x *Xlsx, f *FieldInfo, row []string, mode string) {
	var idx int
	for _, field := range f.Fields {
		if field.Mode == mode || field.Mode == "b" {
			exportRow(x, field, row, idx, mode)
			idx++
		}
	}
	x.trimData("\n")
}

func exportRow(x *Xlsx, f *FieldInfo, row []string, index int, mode string) {
	deepth := f.Deepth + 1
	indent := getIndent(deepth)

	if f.Index == -1 {
		// root, eg.: [1001] = {
		x.appendData(indent)
		x.appendData("[")
		x.appendData(row[0])
		x.appendData("] = {\n")
		exportChildRow(x, f, row, mode)
		x.appendData(indent)
		x.appendData("},\n")
	} else {
		if f.Type == "json" {
			// json 格式化
			val := formatValue(f, row[f.Index])
			formatJson(x, f, index+1, val)
		} else {
			x.appendData(indent)
			if f.Parent.IsArray {
				x.appendData("[")
				x.appendData(strconv.Itoa(index + 1))
				x.appendData("] = ")
			} else {
				x.appendData(f.Name)
				x.appendData(" = ")
			}
			if len(f.Fields) > 0 {
				x.appendData("{\n")
				exportChildRow(x, f, row, mode)
				x.appendData(indent)
				x.appendData("}")
				x.appendData(",\n")
			} else {
				val := formatValue(f, row[f.Index])
				x.appendData(val)
				x.appendData(",\n")
			}
		}
	}
}

func formatString(val string) string {
	val = strings.Replace(val, "\"", "\\\"", -1)
	return fmt.Sprintf("\"%s\"", val)
}

func formatValue(f *FieldInfo, val string) string {
	if f.Type == "string" {
		return formatString(val)
	} else {
		return val
	}
}

func formatJsonKey(key interface{}) string {
	var keystr string
	switch key.(type) {
	case int, uint:
		keystr = fmt.Sprintf("[%v]", key)
	default:
		keystr = fmt.Sprintf("%v", key)
	}
	return keystr
}

func formatJsonValue(x *Xlsx, key interface{}, obj interface{}, deepth int) {
	indent := getIndent(deepth + 1)
	x.appendData(indent)
	x.appendData(formatJsonKey(key))
	x.appendData(" = ")

	switch obj.(type) {
	case map[interface{}]interface{}:
		x.appendData("{\n")
		for k, v := range obj.(map[interface{}]interface{}) {
			formatJsonValue(x, k, v, deepth+1)
		}
		x.trimData("\n")
		x.appendData(indent)
		x.appendData("}")
		x.appendData(",\n")
	case map[string]interface{}:
		x.appendData("{\n")
		for k, v := range obj.(map[string]interface{}) {
			formatJsonValue(x, k, v, deepth+1)
		}
		x.trimData("\n")
		x.appendData(indent)
		x.appendData("}")
		x.appendData(",\n")
	case []interface{}:
		x.appendData("{\n")
		for i, v := range obj.([]interface{}) {
			formatJsonValue(x, i+1, v, deepth+1)
		}
		x.trimData("\n")
		x.appendData(indent)
		x.appendData("}")
		x.appendData(",\n")
	case string:
		x.appendData(formatString(obj.(string)))
		x.appendData(",\n")
	default:
		x.appendData(fmt.Sprintf("%v", obj))
		x.appendData(",\n")
	}
}

func formatJson(x *Xlsx, f *FieldInfo, index int, jsonStr string) error {
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
	formatJsonValue(x, key, result, f.Deepth)
	return nil
}

func writeToFile(x *Xlsx, mode string) {
	var outpath string
	if mode == "c" {
		outpath = FlagClient
	} else if mode == "s" {
		outpath = FlagServer
	}
	file := fmt.Sprintf("%s/%s.lua", outpath, x.FileName)
	outFile, operr := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if operr != nil {
		return
	}
	defer outFile.Close()

	outFile.WriteString(strings.Join(x.Datas, ""))
	outFile.Sync()
}
