package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/ibitcat/gotext"
	"github.com/panjf2000/ants/v2"
)

func main() {
	flag.Parse()
	if Flaghelp || flag.NFlag() <= 0 {
		flag.Usage()
		return
	}

	// xlsx path
	xlsxPath, err := checkPathVaild(Flagpath)
	if err != nil {
		panic(err)
	}

	// i18n output path
	if len(FlagI18nLang) > 0 {
		I18nLocale = gotext.NewLocale(FlagI18nPath, FlagI18nLang)
		I18nLocale.AddDomain("default")
		I18nLocale.ClearAllRefs()
	}

	Mode2Format = make(map[string]string)
	Mode2Format["server"] = FlagServer
	Mode2Format["client"] = FlagClient

	startTime := time.Now()
	// walk
	walkPath(xlsxPath)
	loadExportLog()

	xlsxCount := len(XlsxList)
	if xlsxCount > 0 {
		LoadingChan = make(chan struct{}, xlsxCount)
		go processMsg()

		// parse
		var wg sync.WaitGroup
		p, _ := ants.NewPoolWithFunc(10, func(i interface{}) {
			startParse(i)
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
	fmt.Printf("Total Cost: %d ms\n", getDurationMs(startTime))
	// fmt.Printf("running goroutines: %d\n", p.Running())
}
