package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/xuri/excelize/v2"
)

// flags
var (
	Flaghelp   bool
	Flagpath   string // excel路径
	FlagLang   string // 输出语言（lua、json）
	FlagClient string // client 输出路径
	FlagServer string // server 输出路径
)

// global vars
var (
	LastModifyTime map[string]uint64 //文件的最后修改时间
	XlsxList       []*Xlsx
)

func init() {
	flag.BoolVar(&Flaghelp, "help", false, "Excelparser help.")
	flag.StringVar(&Flagpath, "path", "", "excel input path.")
	flag.StringVar(&FlagClient, "client", "", "client export path.")
	flag.StringVar(&FlagServer, "server", "", "server export path.")
	flag.StringVar(&FlagLang, "lang", "", "target language for parsing.")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `ex version: 2022.0.0-M1
	Usage: excelparser [OPTIONS]
		eg.: excelparser --path=./xlsx --lang=lua server=./slua --client=./clua
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
		if len(FlagClient) > 0 {
			exportRows(xlsx, "c")
		}
		if len(FlagServer) > 0 {
			exportRows(xlsx, "s")
		}
	}
	xlsx.printResult()
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
	fmt.Println(len(FlagClient), len(FlagServer))
	if len(FlagClient) == 0 && len(FlagServer) == 0 {
		panic("You must specify an output path.")
	} else {
		if len(FlagClient) > 0 {
			createOutput(FlagClient)
		}
		if len(FlagServer) > 0 {
			createOutput(FlagServer)
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

	fmt.Printf("running goroutines: %d\n", p.Running())
}
