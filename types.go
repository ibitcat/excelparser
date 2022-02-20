// types

package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type FieldInfo struct {
	Index   int          // 对应的excel列序号
	Desc    string       // 字段描述
	Name    string       // 字段名
	Type    string       // 字段数据类型
	RawType string       // 原始字段数据类型
	Mode    string       // 生成方式(s=server,c=client,b=both)
	Count   int          // 成员数量
	Array   []int        // 数组维度（1d, 2d）
	Parent  *FieldInfo   // 父字段
	Fields  []*FieldInfo // 成员
}

type Xlsx struct {
	Name      string
	Comments  []string
	Data      []string
	RootField *FieldInfo
}

// 解析头部
func parseHeader(descs, names, defs, modes []string) *FieldInfo {
	// 字段名是否重复
	// 是否有id
	keyName := names[0]
	if keyName != "id" {
		// key 必须以 id 命名
		fmt.Println(keyName)
	}
	for i, def := range defs {
		if len(strings.TrimSpace(def)) == 0 {
			panic(fmt.Sprintf("列[%d]数据类型为空", i))
		}
	}

	rootField := new(FieldInfo)
	rootField.Index = -1
	rootField.Fields = make([]*FieldInfo, 0, len(names))

	for idx := 0; idx < len(defs); {
		index := parseField(rootField, idx, descs, names, defs, modes)
		idx = index + 1
	}
	//fmt.Println(rootField)
	return rootField
}

/*
dict<2>
dict[2]
dict[2][3]
int[3]
*/
func parseField(parent *FieldInfo, index int, descs, names, defs, modes []string) int {
	rawType := strings.TrimSpace(defs[index])

	var fieldType string = rawType
	var fieldNum int
	var array []int
	if arrayBegin := strings.Index(rawType, "["); arrayBegin != -1 {
		// array
		fieldType = rawType[:arrayBegin]

		array = make([]int, 0)
		tmp := rawType[arrayBegin:]
		arrayBegin = strings.Index(tmp, "[")
		arrayEnd := strings.Index(tmp, "]")
		for arrayBegin != -1 && arrayEnd != -1 {
			dims, _ := strconv.Atoi(tmp[(arrayBegin + 1):arrayEnd])
			array = append(array, dims)
			tmp = tmp[arrayBegin+1:]

			arrayBegin = strings.Index(tmp, "[")
			arrayEnd = strings.Index(tmp, "]")
		}
	} else if dictBegin := strings.Index(rawType, "<"); dictBegin != -1 {
		// dict
		fieldType = rawType[:dictBegin]
		dictEnd := strings.Index(rawType, ">")
		fieldNum, _ = strconv.Atoi(rawType[(dictBegin + 1):dictEnd])
	}

	field := new(FieldInfo)
	field.Index = index
	field.Type = fieldType
	field.RawType = rawType
	field.Array = array
	field.Parent = parent
	if len(descs) > index {
		field.Desc = descs[index]
	}
	if len(names) > index {
		field.Name = names[index]
	}
	if len(modes) > index {
		field.Mode = modes[index]
	}
	parent.Fields = append(parent.Fields, field)
	if len(parent.Array) > 0 {
		if field.Type != parent.Type {
			panic(errors.New("类型不匹配"))
		}
	}

	// array
	if len(array) > 0 {
		count := 1
		for _, v := range array {
			count = v * count
		}
		field.Count = count

		for i := 0; i < count; i++ {
			index = parseField(field, index+1, descs, names, defs, modes)
		}
	} else {
		if fieldType == "dict" {
			field.Count = fieldNum
			for i := 0; i < fieldNum; i++ {
				index = parseField(field, index+1, descs, names, defs, modes)
			}
		}
	}
	return index
}

// id 必须是int
// id 冲突
// 字段名冲突（同层）
// 类型检查(例如: int 类型的字段填了 string， 耗性能)
// 高级特性：id公式，数值范围检查，字段注释，配置行注释
func (x *Xlsx) parseRows(rootField *FieldInfo, rows [][]string) {
	// comment
	x.Comments = make([]string, 0)
	for _, field := range rootField.Fields {
		x.parseComment(field, 0)
	}
	fmt.Println(strings.Join(x.Comments, "\n"))

	x.Data = make([]string, 0)
	for line := 4; line < len(rows); line++ {
		row := rows[line]
		if strings.HasPrefix(row[0], "//") || row[0] == "" {
			continue
		}
		x.parseRow(rootField, row, -1, 0)
	}
	fmt.Println(strings.Join(x.Data, "\n"))
}

func (x *Xlsx) parseComment(f *FieldInfo, deepth int) {
	x.Comments = append(x.Comments, fmt.Sprintf("-- %-30s %-10s %s", strings.Repeat(" ", deepth*2)+f.Name, f.Type, f.Desc))
	if len(f.Fields) > 0 {
		for _, field := range f.Fields {
			x.parseComment(field, deepth+1)
		}
	}
}

func (x *Xlsx) parseRow(f *FieldInfo, row []string, index, deepth int) {
	if f.Index == -1 {
		// root
		x.Data = append(x.Data, fmt.Sprintf("[%s] = {", row[0]))
	} else {
		if len(f.Fields) > 0 {
			// dict or array
			x.Data = append(x.Data, fmt.Sprintf("%s%s = {", strings.Repeat(" ", deepth*2), f.Name))
		} else {
			if f.Parent != nil && len(f.Parent.Array) > 0 {
				x.Data = append(x.Data, fmt.Sprintf("%s[%d] = %s,", strings.Repeat(" ", deepth*2), index, row[f.Index]))
			} else {
				x.Data = append(x.Data, fmt.Sprintf("%s%s = %s,", strings.Repeat(" ", deepth*2), f.Name, row[f.Index]))
			}

		}
	}
	for idx, field := range f.Fields {
		x.parseRow(field, row, idx+1, deepth+1)
	}
	if len(f.Fields) > 0 {
		x.Data = append(x.Data, fmt.Sprintf("%s},", strings.Repeat(" ", deepth*2)))
	}
}
