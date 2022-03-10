package main

import (
	"fmt"
	"strconv"
	"strings"
)

// id 必须是int
// id 冲突
// 字段名冲突（同层）
// 类型检查(例如: int 类型的字段填了 string， 耗性能)
// 高级特性：id公式，数值范围检查，字段注释，配置行注释
func exportRows(x *Xlsx) {
	// comment
	x.Comments = make([]string, 0)
	for _, field := range x.RootField.Fields {
		exportComment(x, field)
	}
	fmt.Println(strings.Join(x.Comments, "\n"))

	// data
	x.Data = make([]string, 0)
	for line := 4; line < len(x.Rows); line++ {
		row := x.Rows[line]
		if strings.HasPrefix(row[0], "//") || row[0] == "" {
			continue
		}
		exportRow(x, x.RootField, row, -1, 0)
	}
	fmt.Println(strings.Join(x.Data, "\n"))
}

func exportComment(x *Xlsx, f *FieldInfo) {
	var comment string
	if f.Parent.Array {
		comment = fmt.Sprintf("-- %-30s %-10s %s", strings.Repeat(" ", f.Deepth*2)+"["+strconv.Itoa(f.ArrayIdx)+"]", f.RawType, f.Desc)
	} else {
		comment = fmt.Sprintf("-- %-30s %-10s %s", strings.Repeat(" ", f.Deepth*2)+f.Name, f.RawType, f.Desc)
	}
	x.Comments = append(x.Comments, comment)
	if len(f.Fields) > 0 {
		for _, field := range f.Fields {
			exportComment(x, field)
		}
	}
}

func exportRow(x *Xlsx, f *FieldInfo, row []string, index, deepth int) {
	if f.Index == -1 {
		// root
		x.Data = append(x.Data, fmt.Sprintf("[%s] = {", row[0]))
	} else {
		if len(f.Fields) > 0 {
			if f.Parent.Array {
				x.Data = append(x.Data, fmt.Sprintf("%s[%d] = {", strings.Repeat(" ", deepth*2), index))
			} else {
				x.Data = append(x.Data, fmt.Sprintf("%s%s = {", strings.Repeat(" ", deepth*2), f.Name))
			}
		} else {
			if f.Parent.Array {
				x.Data = append(x.Data, fmt.Sprintf("%s[%d] = %s,", strings.Repeat(" ", deepth*2), index, row[f.Index]))
			} else {
				x.Data = append(x.Data, fmt.Sprintf("%s%s = %s,", strings.Repeat(" ", deepth*2), f.Name, row[f.Index]))
			}
		}
	}
	for idx, field := range f.Fields {
		exportRow(x, field, row, idx+1, deepth+1)
	}
	if len(f.Fields) > 0 {
		x.Data = append(x.Data, fmt.Sprintf("%s},", strings.Repeat(" ", deepth*2)))
	}
}
