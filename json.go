package main

import (
	"fmt"
	"strings"
)

type JsonFormater struct {
	*Xlsx
	mode string
}

// id 冲突
// 类型检查(例如: int 类型的字段填了 string， 耗性能)
// 高级特性：id公式，数值范围检查，字段注释，配置行注释
func (j *JsonFormater) formatRows() {
	// 复用 datas
	j.clearData()

	// data
	j.appendData("[\n")
	for line := 4; line < len(j.Rows); line++ {
		row := j.Rows[line]
		if strings.HasPrefix(row[0], "//") || row[0] == "" {
			continue
		}
		j.formatRow(j.RootField, row, -1)
	}
	j.trimData("}\n")
	j.appendData("]\n")
	j.exportToFile()
}

/// datas
func (l *JsonFormater) formatChildRow(f *FieldInfo, row []string) {
	var idx int
	for _, field := range f.Fields {
		if field.Mode == l.mode || field.Mode == "b" {
			l.formatRow(field, row, idx)
			idx++
		}
	}
	l.trimData("\n")
}

func (l *JsonFormater) formatRow(f *FieldInfo, row []string, index int) {
	deepth := f.Deepth + 1
	indent := getIndent(deepth)

	if f.Index == -1 {
		// root, eg.: [1001] = {
		l.appendData(indent)
		l.appendData("{\n")
		l.formatChildRow(f, row)
		l.appendData(indent)
		l.appendData("},\n")
	} else {
		l.appendData(indent)
		if !f.Parent.IsArray {
			l.appendData("\"" + f.Name + "\":")
		}
		if len(f.Fields) > 0 {
			if f.IsArray {
				l.appendData("[\n")
			} else {
				l.appendData("{\n")
			}
			l.formatChildRow(f, row)
			l.appendData(indent)
			if f.IsArray {
				l.appendData("]")
			} else {
				l.appendData("}")
			}
			l.appendData(",\n")
		} else {
			val := formatValue(f, row[f.Index])
			l.appendData(val)
			l.appendData(",\n")
		}
	}
}

/// export
func (j *JsonFormater) exportToFile() {
	var outpath string
	if j.mode == "c" {
		outpath = FlagClient.OutPath
	} else if j.mode == "s" {
		outpath = FlagServer.OutPath
	}
	fileName := fmt.Sprintf("%s/%s.json", outpath, j.FileName)
	j.writeToFile(fileName)
}
