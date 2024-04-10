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
	Kind  int    // 类型定义
	Cap   int    // 容量（for array）
	I18n  bool   // 是否有国际化字符串(for string,json)
	Aname string // alias type name(for struct)
	Ktype *Type  // 键类型(for map)
	Vtype *Type  // 值类型
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
	case TStruct:
		if inJson {
			// json 不支持结构体
			return false
		}
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

func (t *Type) checkJsonObj(obj interface{}) bool {
	switch t.Kind {
	case TArray:
		if array, ok := obj.([]interface{}); !ok {
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
			if m, ok := obj.(map[string]interface{}); !ok {
				return false
			} else {
				for _, v := range m {
					if !t.Vtype.checkJsonObj(v) {
						return false
					}
				}
			}
		} else {
			if m, ok := obj.(map[interface{}]interface{}); !ok {
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
	}
	return true
}

func (t *Type) checkJsonVal(val string) bool {
	var result interface{}
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
