package main

import (
	"encoding/json"
	"strings"
)

const (
	TNone   int = -1   // 非法类型
	TAny    int = iota // any
	TInt               // 有符号整数
	TUint              // 无符号整数
	TFloat             // 浮点数
	TBool              // 布尔型
	TString            // 字符串
	TArray             // 数组
	TMap               // map
	TStruct            // 结构体
	TJson              // json
)

type Type struct {
	Kind   int              // 类型定义
	Cap    int              // 容量（for array）
	I18n   bool             // 是否有国际化字符串(for string,json)
	Aname  string           // alias type name(for 具名结构体)
	Ktype  *Type            // 键类型(for map)
	Vtype  *Type            // 值类型(for map,array,json)
	Ftypes map[string]*Type // 字段类型(for 匿名结构体)
}

// methods
func (t *Type) isVaild(inJson bool) bool {
	if t.Kind == TNone {
		return false
	}

	switch t.Kind {
	case TAny:
		if inJson {
			return false
		}
	case TArray:
		if t.Vtype == nil {
			return false
		}
		if !inJson && t.Cap < 0 {
			return false
		}
		return t.Vtype.isVaild(inJson)
	case TMap:
		if !t.Ktype.isBuiltin() {
			return false
		}
		ok := t.Ktype.isVaild(inJson)
		if ok {
			ok = t.Vtype.isVaild(inJson)
		}
		return ok
	case TJson:
		if t.Vtype != nil {
			return t.Vtype.isVaild(true)
		}
	}
	return true
}

func (t *Type) isAny() bool {
	return t.Kind == TAny
}

func (t *Type) isBuiltin() bool {
	switch t.Kind {
	case TInt, TUint, TFloat, TBool, TString:
		return true
	}
	return false
}

func (t *Type) isNumber() bool {
	switch t.Kind {
	case TInt, TUint, TFloat:
		return true
	}
	return false
}

func (t *Type) isInteger() bool {
	switch t.Kind {
	case TInt, TUint:
		return true
	}
	return false
}

func (t *Type) isI18nString() bool {
	return t.Kind == TString && t.I18n
}

// 是否是需要递归的类型
func (t *Type) isRecursice() bool {
	switch t.Kind {
	case TArray, TMap, TStruct:
		return true
	}
	return false
}

func (t *Type) defaultValue() string {
	switch t.Kind {
	case TInt, TUint:
		return "0"
	case TFloat:
		return "0.0"
	case TBool:
		return "false"
	case TString:
		return "\"\""
	default:
		return ""
	}
}

func (t *Type) formatValue(val string) string {
	val = strings.TrimSpace(val)
	if FlagDefault && len(val) == 0 {
		return t.defaultValue()
	} else {
		if t.Kind == TString {
			return formatString(val)
		} else if t.Kind == TBool {
			if val == "0" {
				return "false"
			} else {
				return "true"
			}
		} else {
			return val
		}
	}
}

func (t *Type) checkJsonObj(obj any) bool {
	switch t.Kind {
	case TArray:
		if array, ok := obj.([]any); !ok {
			return false
		} else {
			if t.Cap > 0 {
				if len(array) != t.Cap {
					return false
				}
			}

			for _, v := range array {
				if !t.Vtype.checkJsonObj(v) {
					return false
				}
			}
		}
	case TMap:
		if t.Ktype.Kind == TString {
			if m, ok := obj.(map[string]any); !ok {
				return false
			} else {
				for _, v := range m {
					if !t.Vtype.checkJsonObj(v) {
						return false
					}
				}
			}
		} else {
			if m, ok := obj.(map[any]any); !ok {
				return false
			} else {
				for k, v := range m {
					if !t.Ktype.checkJsonObj(k) {
						return false
					}
					if !t.Vtype.checkJsonObj(v) {
						return false
					}
				}
			}
		}
	case TInt, TUint:
		_, ok := obj.(float64)
		return ok
	case TBool:
		_, ok := obj.(bool)
		return ok
	case TString:
		_, ok := obj.(string)
		return ok
	case TStruct:
		if s, ok := obj.(map[string]any); !ok {
			return false
		} else {
			for k, ft := range t.Ftypes {
				vv, found := s[k]
				if !found {
					return false
				}
				if !ft.checkJsonObj(vv) {
					return false
				}
			}
		}
	}
	return true
}

func (t *Type) checkJsonVal(val string) bool {
	var result any
	err := json.Unmarshal([]byte(val), &result)
	if err == nil {
		return t.checkJsonObj(result)
	}
	return false
}

func (t *Type) jsonHasI18n() bool {
	switch t.Kind {
	case TArray, TMap:
		return t.Vtype.jsonHasI18n()
	case TString:
		return t.I18n
	}
	return false
}

func (t *Type) isI18nJson() bool {
	if t.Kind == TJson {
		vt := t.Vtype
		if vt != nil {
			return vt.jsonHasI18n()
		}
	}
	return false
}

func (t *Type) luaTypeName() string {
	switch t.Kind {
	case TInt:
		return "integer"
	case TUint:
		return "integer"
	case TFloat:
		return "number"
	case TBool:
		return "boolean"
	case TString:
		return "string"
	case TArray:
		// int[]
		return t.Vtype.luaTypeName() + "[]"
	case TMap:
		return "table" + "<" + t.Ktype.luaTypeName() + "," + t.Vtype.luaTypeName() + ">"
	case TStruct:
		return "table"
	case TJson:
		if t.Vtype != nil {
			return t.Vtype.luaTypeName()
		}
		return "table"
	}
	return "any"
}
