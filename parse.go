// types

package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type FieldInfo struct {
	Index    int          // 对应的excel列序号
	Desc     string       // 字段描述
	Name     string       // 字段名
	Type     string       // 字段数据类型
	RawType  string       // 原始字段数据类型
	Mode     string       // 生成方式(s=server,c=client,b=both)
	Deepth   int          // 字段深度
	Array    bool         // 是否是数组字段
	ArrayIdx int          // 数组索引
	Parent   *FieldInfo   // 父字段
	Fields   []*FieldInfo // 成员
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

// 解析头部
func (x *Xlsx) parseHeader() {
	// 是否有id
	keyName := x.Names[0]
	if keyName != "id" {
		fmt.Println(keyName)
		panic("key 必须以 id 命名")
	}
	for i, def := range x.Types {
		if len(strings.TrimSpace(def)) == 0 {
			panic(fmt.Sprintf("列[%d]数据类型为空", i))
		}
	}

	rootField := new(FieldInfo)
	rootField.Index = -1
	rootField.Fields = make([]*FieldInfo, 0, len(x.Names))
	x.RootField = rootField

	for idx := 0; idx < len(x.Types); idx++ {
		idx = x.parseField(rootField, idx, -1)
	}

	// check
	// id 类型检查(int 和 string)
	// 字段名冲突
	fmt.Println(rootField, math.MaxInt)
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
	parent.Fields = append(parent.Fields, field)

	if arrayBegin := strings.Index(def, "["); arrayBegin != -1 {
		// array
		field.Type = def[:arrayBegin]
		field.Array = true

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
		if parent.Array {
			if field.Type != parent.Type {
				//panic(errors.New("类型不匹配"))
			}
		}
	}
	return index
}
