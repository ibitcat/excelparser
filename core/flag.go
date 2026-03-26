package core

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var Flaghelp bool

// 字符串切片类型，用于接收命令行参数中的逗号分隔列表
type StringFlagSlice []string

// 实现 flag.Value 接口
func (s *StringFlagSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *StringFlagSlice) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

func init() {
	// flag
	flag.BoolVar(&Flaghelp, "help", false, "Excelparser help.")
	flag.BoolVar(&GFlags.Pretty, "indent", false, "Json indent flag.")
	flag.BoolVar(&GFlags.Force, "force", false, "Force export of all excel files.")
	flag.BoolVar(&GFlags.Compact, "compact", false, "Toggle compressed field content.")
	flag.StringVar(&GFlags.Path, "path", "", "Excel input path.")
	flag.Var((*StringFlagSlice)(&GFlags.Client), "client", "Export client fields using the specified format, separated by comma. eg: lua,json")
	flag.Var((*StringFlagSlice)(&GFlags.Server), "server", "Export server fields using the specified format, separated by comma. eg: lua,json")
	flag.StringVar(&GFlags.I18nPath, "i18n", "./locales", "I18n po file path.")
	flag.StringVar(&GFlags.I18nLang, "lang", "", "I18n language.")
	flag.StringVar(&GFlags.Output, "output", ".", "Export output path.")
	flag.Var((*StringFlagSlice)(&GFlags.Files), "files", "Specify excel files to export, separated by comma. eg: item@道具.xlsx,hero@英雄.xlsx")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `excelparser version: 2025.0.1
    Usage: excelparser [OPTIONS]
    eg.: excelparser.exe --path=./xlsx --server=lua    --client=lua --output=./out
         excelparser.exe --path=./xlsx --server=json   --client=json --indent
         excelparser.exe --path=./xlsx --server=lua    --client=json --indent
         excelparser.exe --path=./xlsx --server=csharp --client=csharp --output=./out
         excelparser.exe --path=./xlsx --server=lua    --indent --i18n=./i18n --lang=en
    Formats: lua, json, csharp (MessagePack binary + C# class)
    Options:
`)
	flag.PrintDefaults()
}
