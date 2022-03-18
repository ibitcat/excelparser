// types

package main

import (
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
	Deepth  int          // 字段深度
	IsArray bool         // 是否是数组
	Parent  *FieldInfo   // 父字段
	Fields  []*FieldInfo // 成员
}

type Xlsx struct {
	PathName  string // 文件完整路径
	FileName  string // 文件名
	Descs     []string
	Names     []string
	Types     []string
	Modes     []string
	RootField *FieldInfo
	Rows      [][]string
	Datas     []string
	Errors    []string // 错误信息
	TimeCost  int      // 耗时
}

func (x *Xlsx) appendError(errMsg string) {
	x.Errors = append(x.Errors, errMsg)
}

func (x *Xlsx) appendData(str string) {
	x.Datas = append(x.Datas, str)
}

func (x *Xlsx) trimData(str string) {
	x.Datas[len(x.Datas)-1] = str
}

func (x *Xlsx) checkKeyField() {
	keyField := x.RootField.Fields[0]
	if keyField.Index != 0 {
		x.appendError("key 字段不能注释")
	} else if keyField.Name != "id" {
		x.appendError("key 必须以 id 命名")
	}
	if !(isNumberType(keyField.Type) || keyField.Type == "string") {
		x.appendError("key 字段数据类型必须为定点整数或字符串")
	}
}

func (x *Xlsx) checkDupField(parent *FieldInfo) {
	if len(parent.Fields) == 0 {
		return
	} else {
		tmpMap := map[string]int{}
		for i, field := range parent.Fields {
			if !parent.IsArray {
				index, ok := tmpMap[field.Name]
				if ok {
					x.appendError(fmt.Sprintf("字段名【%s】冲突：#%d<-->#%d ", field.Name, index, field.Index))
				} else {
					tmpMap[field.Name] = field.Index
				}
			} else {
				if parent.Type == "dict" {
					if field.Type != parent.Type {
						x.appendError(fmt.Sprintf("结构体数组元素类型错误：%s[%d] ", parent.Name, i))
					}
				} else {
					if field.RawType != parent.Type {
						x.appendError(fmt.Sprintf("数组元素类型错误：%s[%d] ", parent.Name, i))
					}
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
		idx = x.parseField(rootField, idx)
	}

	// check
	x.checkKeyField()
	x.checkDupField(rootField)
}

func (x *Xlsx) parseField(parent *FieldInfo, index int) int {
	if index >= len(x.Types) {
		return index
	}

	def := strings.TrimSpace(x.Types[index])
	field := new(FieldInfo)
	field.Index = index
	field.RawType = def
	field.Parent = parent
	field.Deepth = parent.Deepth + 1
	if len(x.Descs) > index {
		field.Desc = x.Descs[index]
	}
	if len(x.Names) > index {
		field.Name = x.Names[index]
	}

	// mode
	var mode string
	if len(x.Modes) > index {
		mode = x.Modes[index]
	}
	if parent.Index >= 0 && parent.Mode != "b" {
		// 继承父节点 mode
		mode = parent.Mode
	} else if strings.HasPrefix(field.Name, "//") {
		// 列被注释
		mode = ""
	}
	field.Mode = mode
	if len(mode) > 0 {
		parent.Fields = append(parent.Fields, field)
	}

	if arrayBegin := strings.LastIndex(def, "["); arrayBegin != -1 {
		// array
		field.Type = def[:arrayBegin]
		field.IsArray = true

		// sub array
		arrayEnd := strings.LastIndex(def, "]")
		fieldNum, _ := strconv.Atoi(def[(arrayBegin + 1):arrayEnd])
		for i := 0; i < fieldNum; i++ {
			index = x.parseField(field, index+1)
		}
	} else {
		field.Type = def

		isDict := strings.HasPrefix(def, "dict")
		if isDict {
			dictBegin := strings.Index(def, "<")
			dictEnd := strings.Index(def, ">")
			field.Type = def[:dictBegin]
			fieldNum, _ := strconv.Atoi(def[(dictBegin + 1):dictEnd])
			for i := 0; i < fieldNum; i++ {
				index = x.parseField(field, index+1)
			}
		}
	}
	return index
}

func (x *Xlsx) printResult() []string {
	results := make([]string, 0)
	results = append(results, Splitline)

	errNum := len(x.Errors)
	if errNum == 0 {
		results = append(results, fmt.Sprintf("%-20s| %-5dms", x.FileName, x.TimeCost))
	} else if errNum == 1 {
		results = append(results, fmt.Sprintf("%-20s| %s", x.FileName, x.Errors[0]))
	} else {
		mid := len(x.Errors) / 2
		for i, err := range x.Errors {
			if mid == (i + 1) {
				results = append(results, fmt.Sprintf("%-20s| %s", x.FileName, err))
			} else {
				results = append(results, fmt.Sprintf("%-20s| %s", "", err))
			}
		}
	}
	return results
}
