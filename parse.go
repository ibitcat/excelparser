// types

package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type FieldInfo struct {
	Index     int          // 对应的excel列序号
	Desc      string       // 字段描述
	Name      string       // 字段名
	Type      string       // 字段数据类型
	RawType   string       // 原始字段数据类型
	Mode      string       // 生成方式(s=server,c=client,b=both)
	Commented bool         // 字段是否被注释
	Deepth    int          // 字段深度
	ArrayIdx  int          // 数组索引
	Parent    *FieldInfo   // 父字段
	Fields    []*FieldInfo // 成员
}

type Xlsx struct {
	PathName  string   // 文件完整路径
	FileName  string   // 文件名
	Errors    []string // 错误信息
	TimeCost  int64    // 耗时
	Comments  []string
	Data      []string
	RootField *FieldInfo
	Descs     []string
	Names     []string
	Types     []string
	Modes     []string
	Rows      [][]string
}

func (x *Xlsx) appendError(errMsg string) {
	x.Errors = append(x.Errors, errMsg)
}

func (x *Xlsx) checkKeyField() {
	keyField := x.RootField.Fields[0]
	if keyField.Name != "id" {
		x.appendError("key 必须以 id 命名")
	}
	if !(isNumberType(keyField.Type) || keyField.Type == "string") {
		x.appendError("key 字段数据类型必须为定点整数或字符串")
	}
	if keyField.Commented {
		x.appendError("key 字段不能注释")
	}
}

func (x *Xlsx) checkDupField(parent *FieldInfo) {
	if len(parent.Fields) == 0 {
		return
	} else {
		tmpMap := map[string]int{}
		for _, field := range parent.Fields {
			if field.ArrayIdx < 0 {
				index, ok := tmpMap[field.Name]
				if ok {
					x.appendError(fmt.Sprintf("字段名【%s】冲突：#%d<-->#%d ", field.Name, index, field.Index))
				} else {
					tmpMap[field.Name] = field.Index
				}
			}
			x.checkDupField(field)
		}
	}
}

func (x *Xlsx) parseHeader() {
	rootField := new(FieldInfo)
	rootField.Index = -1
	rootField.Fields = make([]*FieldInfo, 0, len(x.Names))
	x.RootField = rootField

	for idx := 0; idx < len(x.Types); idx++ {
		idx = x.parseField(rootField, idx, -1)
	}

	// check
	x.checkKeyField()
	x.checkDupField(rootField)
}

func (x *Xlsx) parseField(parent *FieldInfo, index, arrayIdx int) int {
	if index >= len(x.Types) {
		return index
	}

	def := strings.TrimSpace(x.Types[index])
	field := new(FieldInfo)
	field.Index = index
	field.RawType = def
	field.Parent = parent
	field.Deepth = parent.Deepth + 1
	field.ArrayIdx = arrayIdx
	if len(x.Descs) > index {
		field.Desc = x.Descs[index]
	}
	if len(x.Names) > index {
		field.Name = x.Names[index]
	}
	if len(x.Modes) > index {
		field.Mode = x.Modes[index]
	}
	if parent.Commented {
		field.Commented = parent.Commented
	} else {
		field.Commented = strings.HasPrefix(field.Name, "//")
	}
	parent.Fields = append(parent.Fields, field)

	if arrayBegin := strings.Index(def, "["); arrayBegin != -1 {
		// array
		field.Type = def[:arrayBegin]

		// sub array
		arrayEnd := strings.Index(def, "]")
		fieldNum, _ := strconv.Atoi(def[(arrayBegin + 1):arrayEnd])
		for i := 0; i < fieldNum; i++ {
			index = x.parseField(field, index+1, i)
		}
	} else {
		field.Type = def

		isDict := strings.HasPrefix(def, "dict")
		if isDict {
			dictBegin := strings.Index(def, "<")
			dictEnd := strings.Index(def, ">")
			fieldNum, _ := strconv.Atoi(def[(dictBegin + 1):dictEnd])
			for i := 0; i < fieldNum; i++ {
				index = x.parseField(field, index+1, -1)
			}
		}
	}
	return index
}

func (x *Xlsx) printResult() {
	for _, err := range x.Errors {
		log.Println(err)
	}
}
