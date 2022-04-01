package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type JsonFormater struct {
	*Xlsx
	mode string
}

func (j *JsonFormater) formatRows() {
	// 复用 datas
	j.clearData()

	// data
	if j.Vertical {
		j.appendData("{\n")
		j.formatRow(j.RootField, 4, -1)
		j.appendData("}\n")
	} else {
		j.appendData("[\n")
		for line := 4; line < len(j.Rows); line++ {
			key := j.Rows[line][0]
			if strings.HasPrefix(key, "//") || key == "" {
				continue
			}
			j.formatRow(j.RootField, line, -1)
		}
		j.trimData("\n")
		j.appendData("]\n")
	}
	j.exportToFile()
}

/// datas
func (j *JsonFormater) formatChildRow(f *FieldInfo, line int) {
	var idx int
	for _, field := range f.Fields {
		if field.isHitMode(j.mode) {
			j.formatRow(field, line, idx)
			idx++
		}
	}
	j.judgCompressTrim("", "\n")
}

func (j *JsonFormater) formatRow(f *FieldInfo, line, index int) {
	deepth := f.Deepth + 1
	indent := getIndent(deepth)

	if f.Index == -1 {
		// root, eg.: [1001] = {
		if j.Vertical {
			j.formatChildRow(f, line)
		} else {
			j.appendData(indent)
			j.appendData("{")
			j.judgCompressAppend("", "\n")
			j.formatChildRow(f, line)
			j.judgCompressAppend("", indent)
			j.appendData("}")
			j.appendData(",\n")
		}
	} else {
		row := j.Rows[line]
		ok, val := f.getValue(row)
		if !ok {
			return
		}

		j.judgCompressAppend("", indent)
		if !f.Parent.IsArray {
			j.appendData("\"" + f.Name + "\":")
		}
		if len(f.Fields) > 0 {
			if f.IsArray {
				j.judgCompressAppend("[", "[\n")
			} else {
				j.judgCompressAppend("{", "{\n")
			}
			j.formatChildRow(f, line)
			j.judgCompressAppend("", indent)
			if f.IsArray {
				j.appendData("]")
			} else {
				j.appendData("}")
			}
			j.judgCompressAppend(",", ",\n")
		} else {
			if f.Type == "json" {
				if !json.Valid([]byte(val)) {
					j.sprintfError("[%s] json 格式错误：(行%d,列%d)[%s@%s]", j.mode, line+1, f.Index+1, f.Name, f.Desc)
					j.appendData("null")
				} else {
					var out bytes.Buffer
					if FlagCompress {
						json.Compact(&out, []byte(val))
						j.appendData(out.String())
					} else if FlagIndent {
						json.Indent(&out, []byte(val), indent, "  ")
						j.appendData(out.String())
					} else {
						j.appendData(val)
					}
				}
			} else {
				j.appendData(val)
			}
			j.judgCompressAppend(",", ",\n")
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
