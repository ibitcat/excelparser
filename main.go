package main

import (
	"flag"
	"path/filepath"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

func main() {
	flag.Parse()
	if Flaghelp || flag.NFlag() <= 0 {
		flag.Usage()
		return
	}

	// xlsx path
	err, xlsxPath := checkPathVaild(Flagpath)
	if err != nil {
		panic(err)
	}

	// i18n output path
	if len(FlagI18nPath) > 0 {
		openI18nXlsx(FlagI18nPath, FlagI18nLang)
	}

	// walk
	loadLastModTime()
	XlsxList = make([]*Xlsx, 0)
	err = filepath.Walk(xlsxPath, walkFunc)
	if err != nil {
		panic(err)
	}

	xlsxCount := len(XlsxList)
	if xlsxCount > 0 {
		LoadingChan = make(chan struct{}, xlsxCount)
		go processMsg()

		// parse
		var wg sync.WaitGroup
		p, _ := ants.NewPoolWithFunc(10, func(i interface{}) {
			parseExcel(i)
			wg.Done()
		})
		defer p.Release()

		for _, xlsx := range XlsxList {
			wg.Add(1)
			_ = p.Invoke(xlsx)
		}
		wg.Wait()

		// 注意 channel range 需要close channel https://segmentfault.com/a/1190000040399883?utm_source=sf-similar-article
		close(LoadingChan)
		saveConvTime()

		time.Sleep(time.Millisecond * 100)
		printResult()

		if len(FlagI18nPath) > 0 {
			saveI18nXlsx(FlagI18nPath, FlagI18nLang)
		}
	} else {
		printResult()
	}
	// fmt.Printf("running goroutines: %d\n", p.Running())
}
