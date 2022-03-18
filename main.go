package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/xuri/excelize/v2"
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
	Flaghelp   bool
	Flagpath   string     // excel路径
	FlagServer FlagOutput // server 输出路径
	FlagClient FlagOutput
)

// global vars
var (
	LastModifyTime map[string]uint64 //文件的最后修改时间
	XlsxList       []*Xlsx
	Splitline      string
)

func init() {
	Splitline = fmt.Sprintf("%s+%s", strings.Repeat("-", 20), strings.Repeat("-", 50))
	flag.BoolVar(&Flaghelp, "help", false, "Excelparser help.")
	flag.StringVar(&Flagpath, "path", "", "excel input path.")
	flag.Var(&FlagClient, "client", "client export info")
	flag.Var(&FlagServer, "server", "server export info")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `ex version: 2022.0.0-M1
	Usage: excelparser [OPTIONS]
		eg.: excelparser --path=./xlsx --server=lua:./slua --client=lua:./clua
	Options:
`)
	flag.PrintDefaults()
}

func loadLastModTime() {
	LastModifyTime = make(map[string]uint64)
	file, err := os.Open("lastModTime.txt")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			s := strings.Split(line, ",")
			if len(s) == 2 {
				tm, _ := strconv.ParseUint(s[1], 10, 64)
				LastModifyTime[s[0]] = tm
			}
		}
	}
}

func createOutput(outpath string) {
	outDir, err := filepath.Abs(outpath)
	if err != nil {
		panic(err)
	}

	_, err = os.Stat(outDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func walkFunc(path string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() {
		// continue
		return nil
	}

	ok, mErr := filepath.Match("[^~$]*.xlsx", f.Name())
	if ok {
		modifyTime := uint64(f.ModTime().UnixNano() / 1000000)
		if lastTime, ok := LastModifyTime[f.Name()]; !ok || lastTime != modifyTime {
			// 检查文件名
			task := &Xlsx{
				PathName: path,
				FileName: getFileName(f.Name()),
				Errors:   make([]string, 0),
				TimeCost: 0,
			}
			XlsxList = append(XlsxList, task)
		}
		LastModifyTime[f.Name()] = modifyTime
		return nil
	}
	return mErr
}

func parseExcel(i interface{}) {
	xlsx := i.(*Xlsx)
	startTime := time.Now()
	f, err := excelize.OpenFile(xlsx.PathName)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	rows, err := f.GetRows("data")
	if err != nil {
		log.Println(err)
		return
	}

	xlsx.Rows = rows
	xlsx.Descs = rows[0]
	xlsx.Names = rows[1]
	xlsx.Types = rows[2]
	xlsx.Modes = rows[3]

	xlsx.parseHeader()
	if len(xlsx.Errors) == 0 {
		xlsx.Datas = make([]string, 0)

		var exporter iExporter
		if FlagClient.IsVaild() {
			exporter = NewExporter(xlsx, FlagClient.OutLang, "c")
			exporter.exportRows()
		}
		if FlagClient.IsVaild() {
			exporter = NewExporter(xlsx, FlagClient.OutLang, "s")
			exporter.exportRows()
		}
	}
	xlsx.TimeCost = getDurationMs(startTime)
	//ExportLua(xlsx)
}

func main() {
	flag.Parse()
	if Flaghelp || flag.NFlag() <= 0 {
		flag.Usage()
		return
	}

	xlsxPath, err := filepath.Abs(Flagpath)
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(xlsxPath)
	notExist := os.IsNotExist(err)
	if notExist {
		panic("excel path not exist.")
	}

	// output
	if !FlagClient.IsVaild() && !FlagServer.IsVaild() {
		panic("You must specify an output info.")
	} else {
		if FlagClient.IsVaild() {
			createOutput(FlagClient.OutPath)
		}
		if FlagServer.IsVaild() {
			createOutput(FlagServer.OutPath)
		}
	}

	// walk
	loadLastModTime()
	XlsxList = make([]*Xlsx, 0)
	err = filepath.Walk(xlsxPath, walkFunc)
	if err != nil {
		panic(err)
	}

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

	// result
	results := make([]string, 0)
	results = append(results, Splitline)
	results = append(results, fmt.Sprintf("%-20s| %s", "FileName", "Result"))
	for _, xlsx := range XlsxList {
		result := xlsx.printResult()
		results = append(results, result...)
	}
	results = append(results, Splitline)
	fmt.Println(strings.Join(results, "\n"))
	//fmt.Printf("running goroutines: %d\n", p.Running())
}
