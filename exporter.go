// 数据导出

package main

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

var funcs = template.FuncMap{
	"join":   strings.Join,
	"indent": Indent,
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
	"plus": func(a, b int) int {
		return a + b
	},
	"setRootRow": func(f *FieldInfo, row []string) string {
		f.Row = row
		return ""
	},
}

func Indent(deepth int, s string) string {
	return strings.Repeat(" ", deepth*4) + s
}

func ExportLua(x *Xlsx) {
	tmplBytes, _ := ioutil.ReadFile("lua.tmpl")
	tmpl, err := template.New("luaExport").Funcs(funcs).Parse(string(tmplBytes))
	if err != nil {
		panic(err)
	}

	outFile, operr := os.OpenFile("output.lua", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if operr != nil {
		panic(operr)
	}
	defer outFile.Close()

	err = tmpl.Execute(outFile, x)
	if err != nil {
		panic(err)
	}
}
