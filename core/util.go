// 辅助函数

package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// json decode
func isAsciiStr(str string) bool {
	runes := []rune(str)
	return len(runes) == len(str)
}

func getFileName(pathname string) string {
	filenameall := filepath.Base(pathname)
	filesuffix := filepath.Ext(filenameall)
	filename := filenameall[0 : len(filenameall)-len(filesuffix)]
	return filename
}

func getIndent(num int) string {
	if num < 0 {
		num = 0
	}
	if indent, ok := IndentStr[num]; ok {
		return indent
	} else {
		return strings.Repeat(" ", num*2)
	}
}

func formatString(val string) string {
	val = strings.ReplaceAll(val, "\"", "\\\"")
	return fmt.Sprintf("\"%s\"", val)
}

func GetDurationMs(t time.Time) int {
	return int(time.Since(t).Nanoseconds() / 1e6)
}

func rotateRows(rows [][]string) [][]string {
	ret := make([][]string, 0)
	for i := 0; i < len(rows[0]); i++ {
		row := make([]string, 0)
		row = append(row, rows[0][i])
		ret = append(ret, row)
	}
	for i := 1; i < len(rows); i++ {
		for j := 0; j < len(rows[i]); j++ {
			ret[j] = append(ret[j], rows[i][j])
		}
	}
	return ret
}

func ternary[T any](condition bool, ifOutput T, elseOutput T) T {
	if condition {
		return ifOutput
	}

	return elseOutput
}

func CheckPathVaild(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(absPath)
	notExist := os.IsNotExist(err)
	if notExist {
		return "", err
	}
	return absPath, nil
}

// 分割cell坐标
// example: "E3" -> "E",3
func splitAxis(axis string) (int, int) {
	var celly int = 0
	var cellx int = 0
	pos := 0
	for i, c := range axis {
		if c >= '0' && c <= '9' {
			if pos == 0 {
				pos = i
			}
			celly = celly*10 + int(c-'0')
		}
	}

	if pos > 0 {
		cellx = parseAxisX(axis[:pos])
	}
	return cellx, celly
}

func parseAxisX(s string) int {
	num := 0
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			num = num*26 + int(c-'A') + 1
		}
	}
	return num
}

func formatAxisX(x int) string {
	c := []byte{}
	for {
		r1 := (x-1)%26 + 1
		c = append(c, 0)
		copy(c[1:], c[0:])
		c[0] = 'A' + byte(r1-1)
		x = x - r1
		if x <= 0 {
			break
		}
		x = x / 26
	}
	return string(c)
}

func parseType(typ string) *Type {
	t := new(Type)
	t.Kind = TNone
	t.Cap = -1

	switch typ {
	case "int":
		t.Kind = TInt
	case "uint":
		t.Kind = TUint
	case "float":
		t.Kind = TFloat
	case "bool":
		t.Kind = TBool
	case "string":
		t.Kind = TString
	case "any":
		t.Kind = TAny
	case "i18n":
		t.Kind = TString
		t.I18n = true
	default:
		parseCompositeType(typ, t)
	}
	return t
}

// splitStructFields 将结构体字符串按字段分割，正确处理嵌套的括号和数组
// 输入格式：{field1=type1,field2=type2,...}
// 返回：[]string{"field1=type1", "field2=type2", ...}
func splitStructFields(s string) []string {
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return nil
	}

	// 去掉外层大括号
	s = s[1 : len(s)-1]

	var result []string
	var current strings.Builder
	depth := 0        // 跟踪大括号嵌套深度
	bracketDepth := 0 // 跟踪方括号嵌套深度

	for i := range len(s) {
		char := s[i]
		switch char {
		case '{':
			depth++
			current.WriteByte(char)
		case '}':
			depth--
			current.WriteByte(char)
		case '[':
			bracketDepth++
			current.WriteByte(char)
		case ']':
			bracketDepth--
			current.WriteByte(char)
		case ',':
			if depth == 0 && bracketDepth == 0 {
				// 只在最外层处理逗号
				if current.Len() > 0 {
					result = append(result, strings.TrimSpace(current.String()))
					current.Reset()
				}
			} else {
				current.WriteByte(char)
			}
		default:
			current.WriteByte(char)
		}
	}

	// 添加最后一个字段
	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

// 解析复合类型
func parseCompositeType(typ string, t *Type) {
	if len(typ) > 0 && typ[:1] == "[" {
		// array
		s := ArrayRe.FindStringSubmatch(typ)
		if len(s) == 3 {
			t.Kind = TArray
			if len(s[1]) > 0 {
				cap, _ := strconv.Atoi(s[1])
				t.Cap = cap
				if cap == 0 {
					// 数组长度不能为0
					t.Kind = -1
				}
			}
			t.Vtype = parseType(s[2])
		}
	} else if len(typ) >= 3 && typ[:3] == "map" {
		// map
		s := MapRe.FindStringSubmatch(typ)
		if len(s) == 3 {
			t.Kind = TMap
			t.Ktype = parseType(s[1])
			t.Vtype = parseType(s[2])
		}
	} else if len(typ) >= 6 && typ[:6] == "struct" {
		// 具名结构体
		t.Kind = TStruct
		s := strings.SplitN(typ, "#", 2)
		if len(s) == 2 {
			// 结构体类型别名
			t.Aname = s[1]
		}
	} else if len(typ) >= 4 && typ[:4] == "json" {
		// json
		// eg.: json 原始json
		// eg.: json:map[string]int json 的值是 map[string]int
		// eg.: json:{sites=[]{name=string,url=string}} json 的值是一个结构体
		t.Kind = TJson
		s := strings.SplitN(typ, ":", 2)
		if len(s) == 2 {
			// json 的类型别名
			t.Vtype = parseType(s[1])
		}
		t.I18n = t.isI18nJson()
	} else if len(typ) > 0 && typ[0] == '{' && typ[len(typ)-1] == '}' {
		// 匿名结构体
		// eg.: {sites=[]{name=string,url=string},age=int}
		t.Kind = TStruct
		t.Ftypes = make(map[string]*Type)

		// 使用 splitStructFields 正确处理嵌套结构
		parts := splitStructFields(typ)
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				fname := strings.TrimSpace(kv[0])
				ftype := strings.TrimSpace(kv[1])
				if len(fname) > 0 && len(ftype) > 0 {
					t.Ftypes[fname] = parseType(ftype)
				}
			}
		}
	}
}

func toTitle(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
