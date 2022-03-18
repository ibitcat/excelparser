package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
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
	FlagIndent bool       // json格式化
	FlagForce  bool       // 是否强制重新生成
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
	flag.BoolVar(&FlagIndent, "indent", false, "Json indent flag.")
	flag.BoolVar(&FlagForce, "force", false, "Force export all flag.")
	flag.StringVar(&Flagpath, "path", "", "Excel input path.")
	flag.Var(&FlagClient, "client", "Client export info.")
	flag.Var(&FlagServer, "server", "Server export info.")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `ex version: 2022.0.0-M1
    Usage: excelparser [OPTIONS]
    eg.: excelparser --path=./xlsx --server=lua:./slua --client=lua:./clua
         excelparser --path=./xlsx --server=json:./sjson --client=json:./cjson --indent
         excelparser --path=./xlsx --server=lua:./slua --client=json:./cjson --indent
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
		if lastTime, ok := LastModifyTime[f.Name()]; FlagForce || !ok || lastTime != modifyTime {
			// 检查文件名
			task := &Xlsx{
				Name:     f.Name(),
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
		xlsx.appendError("xlsx文件打开失败")
		return
	}
	defer f.Close()

	rows, err := f.GetRows("data")
	if err != nil {
		xlsx.appendError("data sheet 不存在")
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

		var formater iFormater
		keyField := xlsx.getKeyField()
		if FlagClient.IsVaild() && keyField.isHitMode("c") {
			formater = NewFormater(xlsx, FlagClient.OutLang, "c")
			formater.formatRows()
		}
		if FlagServer.IsVaild() && keyField.isHitMode("s") {
			formater = NewFormater(xlsx, FlagServer.OutLang, "s")
			formater.formatRows()
		}
	}
	xlsx.TimeCost = getDurationMs(startTime)
	//ExportLua(xlsx)
}

func saveConvTime() {
	file := "lastModTime.txt"
	outFile, operr := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if operr != nil {
		fmt.Println("创建[lastModTime.txt]文件错误")
	}
	defer outFile.Close()

	for _, xlsx := range XlsxList {
		if len(xlsx.Errors) > 0 {
			delete(LastModifyTime, xlsx.Name)
		}
	}

	// save time
	modTimes := make([]string, 0, len(LastModifyTime))
	for name, tm := range LastModifyTime {
		modTimes = append(modTimes, name+","+strconv.FormatUint(tm, 10))
	}
	outFile.WriteString(strings.Join(modTimes, "\n"))
	outFile.Sync()
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

	// result
	results := make([]string, 0)
	results = append(results, Splitline)
	results = append(results, fmt.Sprintf("%-20s| %s", "FileName", "Result"))
	if len(XlsxList) > 0 {
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

		// export results
		for _, xlsx := range XlsxList {
			result := xlsx.printResult()
			results = append(results, result...)
		}
	} else {
		results = append(results, Splitline)
		results = append(results, fmt.Sprintf("%-20s| %s", "No files", "无需生成"))
	}
	results = append(results, Splitline)
	fmt.Println(strings.Join(results, "\n"))

	saveConvTime()
	//fmt.Printf("running goroutines: %d\n", p.Running())
}
