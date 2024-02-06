package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// flags
var (
	Flaghelp     bool
	FlagIndent   bool   // json格式化
	FlagForce    bool   // 是否强制重新生成
	FlagDefault  bool   // 字段配空时使用默认值填充
	FlagCompact  bool   // 是否紧凑导出
	Flagpath     string // excel路径
	FlagOutput   string // 导出路径
	FlagServer   string // server 导出格式
	FlagClient   string // client 导出格式
	FlagI18nPath string // 国际化配置路径
	FlagI18nLang string // 国际化语言
)

// global vars
var (
	XlsxList    []*Xlsx
	Splitline   string
	CostFormat  string
	InfoFormat  string
	LoadingChan chan struct{}
	MaxErrorCnt int = 11
	Mode2Format map[string]string
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
	flag.BoolVar(&FlagCompact, "compact", false, "Toggle compressed field content.")
	flag.StringVar(&Flagpath, "path", "", "Excel input path.")
	flag.StringVar(&FlagClient, "client", "", "Export client fields using the specified format.")
	flag.StringVar(&FlagServer, "server", "", "Export server fields using the specified format.")
	flag.StringVar(&FlagI18nPath, "i18n", "", "I18n excel file path.")
	flag.StringVar(&FlagI18nLang, "lang", "", "I18n language.")
	flag.StringVar(&FlagOutput, "output", ".", "Export output path.")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `excelparser version: 2022.0.1
    Usage: excelparser [OPTIONS]
    eg.: excelparser.exe --path=./xlsx --server=lua  --client=lua
         excelparser.exe --path=./xlsx --server=json --client=json --indent
         excelparser.exe --path=./xlsx --server=lua  --client=json --indent
         excelparser.exe --path=./xlsx --server=lua  --indent --i18n=./i18n --lang=en
    Options:
`)
	flag.PrintDefaults()
}
