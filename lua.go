package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// id 冲突
// 类型检查(例如: int 类型的字段填了 string， 耗性能)
// 高级特性：id公式，数值范围检查，字段注释，配置行注释
func exportRows(x *Xlsx) {
	cap := len(x.Types) * (len(x.Rows) - 2)
	fmt.Println(cap, len(x.Types))
	x.Data = make([]string, 0, cap)
	// comment
	for _, field := range x.RootField.Fields {
		exportComment(x, field)
	}

	// data
	x.Data = append(x.Data, "\nreturn {")
	for line := 4; line < len(x.Rows); line++ {
		row := x.Rows[line]
		if strings.HasPrefix(row[0], "//") || row[0] == "" {
			continue
		}
		exportRow(x, x.RootField, row, -1, 0)
	}
	x.Data = append(x.Data, "}")
}

func exportComment(x *Xlsx, f *FieldInfo) {
	var keyName string
	if f.ArrayIdx < 0 {
		keyName = strings.Repeat(" ", f.Deepth*2) + f.Name
	} else {
		keyName = strings.Repeat(" ", f.Deepth*2) + "[" + strconv.Itoa(f.ArrayIdx) + "]"
	}
	x.Data = append(x.Data, fmt.Sprintf("-- %-30s %-10s %s", keyName, f.RawType, f.Desc))

	// recursive
	if len(f.Fields) > 0 {
		for _, field := range f.Fields {
			exportComment(x, field)
		}
	}
}

func formatValue(f *FieldInfo, val string) string {
	if f.Type == "string" {
		val = strings.Replace(val, "\"", "\\\"", -1)
		return fmt.Sprintf("\"%s\"", val)
	} else {
		return val
	}
}

func exportRow(x *Xlsx, f *FieldInfo, row []string, index, deepth int) {
	var str string
	indent := getIndent(deepth)

	if f.Index == -1 {
		// root
		str = fmt.Sprintf("[%s] = {", row[0])
	} else {
		if len(f.Fields) > 0 {
			if f.ArrayIdx < 0 {
				str = fmt.Sprintf("%s%s = {", indent, f.Name)
			} else {
				str = fmt.Sprintf("%s[%d] = {", indent, index)
			}
		} else {
			val := formatValue(f, row[f.Index])
			if f.ArrayIdx < 0 {
				str = fmt.Sprintf("%s%s = %s,", indent, f.Name, val)
			} else {
				str = fmt.Sprintf("%s[%d] = %s,", indent, index, val)
			}
		}
	}
	x.Data = append(x.Data, str)
	for idx, field := range f.Fields {
		exportRow(x, field, row, idx+1, deepth+1)
	}
	if len(f.Fields) > 0 {
		x.Data = append(x.Data, fmt.Sprintf("%s},", indent))
	}
}

func writeToFile(x *Xlsx) {
	file := fmt.Sprintf("%s/%s.lua", out, x.FileName)
	outFile, operr := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if operr != nil {
		return
	}
	defer outFile.Close()

	outFile.WriteString(strings.Join(x.Data, "\n"))
	outFile.Sync()
}
