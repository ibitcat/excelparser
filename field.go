package main

import "strings"

type FieldInfo struct {
	Index    int          // 对应的excel列序号
	Desc     string       // 字段描述
	Name     string       // 字段名
	Type     string       // 字段数据类型
	RawType  string       // 原始字段数据类型
	Mode     string       // 生成方式(s=server,c=client,b=both)
	Deepth   int          // 字段深度
	IsArray  bool         // 是否是数组
	Parent   *FieldInfo   // 父字段
	Fields   []*FieldInfo // 成员
	FieldNum int          // 成员数
}

/// methods
func (f *FieldInfo) isHitMode(tMode string) bool {
	if f.Mode == tMode || f.Mode == "b" {
		return true
	}
	return false
}

func (f *FieldInfo) IsVaildType() bool {
	def := f.Type
	if len(f.Fields) > 0 {
		// map or array
		if f.IsArray {
			arrayBegin := strings.Index(f.RawType, "[")
			def = f.RawType[:arrayBegin]
		} else {
			if f.Type == "dict" {
				return true
			}
		}
	}

	switch def {
	case "int", "float", "bool", "string", "json", "dict":
		return true
	}
	return false
}

func (f *FieldInfo) IsVaildMode() bool {
	switch f.Mode {
	case "", "b", "s", "c":
		return true
	default:
		return false
	}
}

func (f *FieldInfo) getValue(row []string) (bool, string) {
	var val string
	if f.Index >= len(row) {
		if FlagDefault {
			val = defaultValue(f.Type)
		} else {
			return false, val
		}
	} else {
		val = formatValue(f, row[f.Index])
	}

	if len(f.Fields) == 0 && len(val) == 0 {
		return false, val
	}
	return true, val
}
