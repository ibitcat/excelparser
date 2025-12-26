package core

import (
	"flag"
	"fmt"
	"os"
)

var Flaghelp bool

func init() {
	// flag
	flag.BoolVar(&Flaghelp, "help", false, "Excelparser help.")
	flag.BoolVar(&GFlags.Pretty, "indent", false, "Json indent flag.")
	flag.BoolVar(&GFlags.Force, "force", false, "Force export of all excel files.")
	flag.BoolVar(&GFlags.Compact, "compact", false, "Toggle compressed field content.")
	flag.StringVar(&GFlags.Path, "path", "", "Excel input path.")
	flag.StringVar(&GFlags.Client, "client", "", "Export client fields using the specified format.")
	flag.StringVar(&GFlags.Server, "server", "", "Export server fields using the specified format.")
	flag.StringVar(&GFlags.I18nPath, "i18n", "./locales", "I18n po file path.")
	flag.StringVar(&GFlags.I18nLang, "lang", "", "I18n language.")
	flag.StringVar(&GFlags.Output, "output", ".", "Export output path.")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `excelparser version: 2025.0.1
    Usage: excelparser [OPTIONS]
    eg.: excelparser.exe --path=./xlsx --server=lua  --client=lua
         excelparser.exe --path=./xlsx --server=json --client=json --indent
         excelparser.exe --path=./xlsx --server=lua  --client=json --indent
         excelparser.exe --path=./xlsx --server=lua  --indent --i18n=./i18n --lang=en
    Options:
`)
	flag.PrintDefaults()
}
