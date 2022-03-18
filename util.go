package main

import (
	"fmt"
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
// https://github.com/ChimeraCoder/gojson/blob/master/json-to-struct.go
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

func formatValue(f *FieldInfo, val string) string {
	if f.Type == "string" {
		return formatString(val)
	} else {
		return val
	}
}

func getDurationMs(t time.Time) int {
	return int(time.Since(t).Nanoseconds() / 1e6)
}
