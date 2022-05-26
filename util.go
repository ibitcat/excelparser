package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var indentStr map[int]string

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

func isNumberType(def string) bool {
	return def == "int" || def == "uint"
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

func defaultValue(ftype string) string {
	switch ftype {
	case "int":
		return "0"
	case "float":
		return "0.0"
	case "bool":
		return "false"
	case "string":
		return "\"\""
	default:
		return ""
	}
}

func formatValue(f *FieldInfo, val string) string {
	val = strings.TrimSpace(val)
	if FlagDefault && len(val) == 0 {
		return defaultValue(f.Type)
	} else {
		if f.Type == "string" {
			return formatString(val)
		} else {
			return val
		}
	}
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

func ternaryString(b bool, trueStr, falseStr string) string {
	if b {
		return trueStr
	}
	return falseStr
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

func checkI18n(name string) (string, bool) {
	if len(name) > 0 && len(FlagI18nPath) > 0 {
		if name[0] == '#' {
			return name[1:], true
		}
		// s := strings.Split(name, "#")
		// if len(s) == 2 {
		// 	return s[0], true
		// }
	}
	return name, false
}
