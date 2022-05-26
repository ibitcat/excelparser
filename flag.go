package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// types
type FlagOutput struct {
	OutLang string // 输出语言
	OutPath string // 输出路径
}

func (f *FlagOutput) String() string {
	return fmt.Sprintf("%s:%s", f.OutLang, f.OutPath)
}

func (f *FlagOutput) Set(value string) error {
	strs := strings.Split(value, ":")
	if len(strs) != 2 {
		return errors.New("导出参数错误")
	}

	f.OutLang = strs[0]
	f.OutPath = strs[1]
	return nil
}

func (f *FlagOutput) IsVaild() bool {
	return len(f.OutLang) > 0 && len(f.OutPath) > 0
}

// flags
var (
	Flaghelp     bool
	FlagIndent   bool       // json格式化
	FlagForce    bool       // 是否强制重新生成
	FlagDefault  bool       // 字段配空时使用默认值填充
	FlagCompress bool       // 是否压缩
	Flagpath     string     // excel路径
	FlagServer   FlagOutput // server 输出路径
	FlagClient   FlagOutput // client 输出信息
	FlagI18nPath string     // 国际化配置路径
	FlagI18nLang string     // 国际化语言
)

// global vars
var (
	LastModifyTime map[string]uint64 //文件的最后修改时间
	XlsxList       []*Xlsx
	Splitline      string
	CostFormat     string
	InfoFormat     string
	LoadingChan    chan struct{}
	MaxErrorCnt    int = 11
)

func init() {
	CostFormat = "%-30s| %-5dms"
	InfoFormat = "%-30s| %s"
	Splitline = fmt.Sprintf("%s+%s", strings.Repeat("-", 30), strings.Repeat("-", 70))

	// flag
	flag.BoolVar(&Flaghelp, "help", false, "Excelparser help.")
	flag.BoolVar(&FlagIndent, "indent", false, "Json indent flag.")
	flag.BoolVar(&FlagForce, "force", false, "Force export of all excel files.")
	flag.BoolVar(&FlagDefault, "default", true, "Fields are filled with default values when they are empty.")
	flag.BoolVar(&FlagCompress, "compress", false, "Toggle compressed field content.")
	flag.StringVar(&Flagpath, "path", "", "Excel input path.")
	flag.Var(&FlagClient, "client", "The client export information. Format like [file type]:[output path], eg. json:./outjson.")
	flag.Var(&FlagServer, "server", "The server export information. Like client flag.")
	flag.StringVar(&FlagI18nPath, "tpath", "", "I18n excel file path.")
	flag.StringVar(&FlagI18nLang, "lang", "", "I18n language.")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `excelparser version: 2022.0.1
    Usage: excelparser [OPTIONS]
    eg.: excelparser.exe --path=./xlsx --server=lua:./slua --client=lua:./clua
         excelparser.exe --path=./xlsx --server=json:./sjson --client=json:./cjson --indent
         excelparser.exe --path=./xlsx --server=lua:./slua --client=json:./cjson --indent
         excelparser.exe --path=./xlsx --server=lua:./slua --indent --tpath=./i18n --lang=cn
    Options:
`)
	flag.PrintDefaults()
}
