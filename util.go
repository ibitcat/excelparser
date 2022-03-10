package main

import (
	"fmt"
	"path/filepath"
)

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
	fmt.Println(pathname, filenameall, filesuffix, filename)
	return filename
}
