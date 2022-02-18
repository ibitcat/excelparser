// types

package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type FieldInfo struct {
	Index  int          // 对应的excel列序号
	Name   string       // 字段名
	Type   string       // 字段数据类型
	Mode   string       // 生成方式(s=server,c=client,b=both)
	Count  int          // 成员数量
	Array  []int        // 数组维度（1d, 2d）
	Fields []*FieldInfo // 成员
}

type Xlsx struct {
	Name      string
	Comments  []string
	RootField *FieldInfo
}

// 解析头部
func parseHeader(names []string, defs []string, modes []string) *FieldInfo {
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

	fmt.Println(len(names))
	for idx := 0; idx < len(defs); {
		index := parseField(rootField, idx, names, defs, modes)
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
func parseField(parent *FieldInfo, index int, names, defs, modes []string) int {
	def := strings.TrimSpace(defs[index])

	fieldType := def
	var fieldNum int
	var array []int
	if arrayBegin := strings.Index(def, "["); arrayBegin != -1 {
		fieldType = def[:arrayBegin]

		array = make([]int, 0)
		tmp := def[arrayBegin:]
		arrayBegin = strings.Index(tmp, "[")
		arrayEnd := strings.Index(tmp, "]")
		for arrayBegin != -1 && arrayEnd != -1 {
			dims, _ := strconv.Atoi(tmp[(arrayBegin + 1):arrayEnd])
			array = append(array, dims)
			tmp = tmp[arrayBegin+1:]

			arrayBegin = strings.Index(tmp, "[")
			arrayEnd = strings.Index(tmp, "]")
		}
	} else if dictBegin := strings.Index(def, "<"); dictBegin != -1 {
		fieldType = def[:dictBegin]
		dictEnd := strings.Index(def, ">")
		fieldNum, _ = strconv.Atoi(def[(dictBegin + 1):dictEnd])
	}

	field := new(FieldInfo)
	field.Index = index
	field.Type = fieldType
	field.Array = array
	if len(names) > index {
		field.Name = names[index]
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
			index = parseField(field, index+1, names, defs, modes)
		}
	} else {
		if fieldType == "dict" {
			field.Count = fieldNum
			for i := 0; i < fieldNum; i++ {
				index = parseField(field, index+1, names, defs, modes)
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

	// for line := 4; line < len(rows); line++ {
	// 	row := rows[line]
	// 	if strings.HasPrefix(row[0], "//") || row[0] == "" {
	// 		continue
	// 	}

	// 	strs := make([]string, 0)
	// 	for _, field := range rootField.Fields {
	// 		strs = append(strs, fmt.Sprintf("%s = %s", field.Name, row[field.Index]))
	// 	}
	// 	fmt.Println(strs)
	// }
}

func (x *Xlsx) parseComment(f *FieldInfo, deepth int) {
	x.Comments = append(x.Comments, fmt.Sprintf("--%s %s %s", strings.Repeat(" ", deepth*4), f.Name, f.Type))
	if len(f.Fields) > 0 {
		for _, field := range f.Fields {
			x.parseComment(field, deepth+1)
		}
	}
}

func parseRow(f *FieldInfo, row []string) {

}
