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
	help      bool
	excelpath string // excel路径
	out       string // 输出路径
	tag       string // 解析标签
	mode      string // 输出模式（lua、json）
)

// global vars
var (
	LastModifyTime map[string]uint64 //文件的最后修改时间
	XlsxList       []*Xlsx
)

func init() {
	flag.BoolVar(&help, "help", false, "Excelparser help.")
	flag.StringVar(&excelpath, "path", "", "excel input path.")
	flag.StringVar(&out, "out", "", "export output path.")
	flag.StringVar(&tag, "tag", "", "parse tag.")
	flag.StringVar(&mode, "mode", "", "parse mode.")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `ex version: 2022.0.0-M1
	Usage: excelparser [OPTIONS]
		eg.: excelparser --path=../xlsx out=../lua --tag=c --mode=lua
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

func createOutput() {
	outDir, err := filepath.Abs(out)
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
		exportRows(xlsx)
		writeToFile(xlsx)
	}
	xlsx.printResult()
	//ExportLua(xlsx)
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	xlsxPath, err := filepath.Abs(excelpath)
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(xlsxPath)
	notExist := os.IsNotExist(err)
	if notExist {
		panic("excel path not exist.")
	}

	// output
	createOutput()

	// walk
	loadLastModTime()
	XlsxList = make([]*Xlsx, 0)
	err = filepath.Walk(xlsxPath, walkFunc)
	if err != nil {
		panic(err)
	}

	// exporter
	// TODO

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
