package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	indentStr  map[int]string
	ArrayRe    = regexp.MustCompile(`^\[(\d*?)\](.+)`)
	MapRe      = regexp.MustCompile(`^map\[(.+?)\](.+)`)
	RawRe      = regexp.MustCompile(`(\w+)<(.+)>`)
	BasicTypes = []string{"int", "uint", "bool", "string", "var"}
)

func init() {
	indentStr = make(map[int]string)
	for i := 0; i < 10; i++ {
		indentStr[i] = strings.Repeat(" ", i*2)
	}
}

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
	if indent, ok := indentStr[num]; ok {
		return indent
	} else {
		return strings.Repeat(" ", num*2)
	}
}

func formatString(val string) string {
	val = strings.Replace(val, "\"", "\\\"", -1)
	return fmt.Sprintf("\"%s\"", val)
}

func getDurationMs(t time.Time) int {
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

func checkPathVaild(path string) (error, string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err, ""
	}
	_, err = os.Stat(absPath)
	notExist := os.IsNotExist(err)
	if notExist {
		return err, ""
	}
	return nil, absPath
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

	if len(typ) > 0 && typ[:1] == "[" {
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
		s := MapRe.FindStringSubmatch(typ)
		if len(s) == 3 {
			t.Kind = TMap
			t.Ktype = parseType(s[1])
			t.Vtype = parseType(s[2])
		}
	} else if len(typ) >= 6 && typ[:6] == "struct" {
		t.Kind = TStruct
		s := RawRe.FindStringSubmatch(typ)
		if len(s) == 3 {
			// 结构体类型别名
			t.Aname = s[2]
		}
	} else if len(typ) >= 4 && typ[:4] == "json" {
		t.Kind = TJson
		s := RawRe.FindStringSubmatch(typ)
		if len(s) == 3 {
			t.Vtype = parseType(s[2])
		}
	} else {
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
		}
	}
	return t
}
