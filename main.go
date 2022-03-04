package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/panjf2000/ants/v2"
	"github.com/xuri/excelize/v2"
)

// flags
var (
	XlsxPath   string // xlsx配置路径
	ParseTag   string // 解析标签
	PtputPath  string // 输出路径
	OutputMode string // 输出模式（lua、json）
)

// global vars
var (
	LastModifyTime map[string]uint64 //文件的最后修改时间
)

func init() {
	flag.BoolVar(&help, "help", false, "Excelparser help.")
	flag.StringVar(&dbUser, "u", "", "User for login if not current user.")
	flag.StringVar(&dbPass, "p", "", "Password to use when connecting to server.")
	flag.StringVar(&dbHost, "h", "localhost", "Connect to host.")
	flag.IntVar(&dbPort, "P", 3306, "Port number to use for connection.")
	flag.StringVar(&dbName, "d", "", "Database to diff.")
	flag.StringVar(&dbFile, "f", "", "Read this sql file to update database.")
	flag.StringVar(&dbChar, "default-character-set=name", "utf8", "Set the default character set.")
	flag.BoolVar(&onlyCk, "only-check", false, "Only check diff.")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `mysqldiff version: 0.1.0
	Usage: mysqldiff [OPTIONS]
		eg.: mysqldiff -u root -p 123456 -h 127.0.0.1 -P 3306 -d database -f filename.sql
	Options:
	`)
	flag.PrintDefaults()
}

func loadLastModTime() {
	LastModifyTime = make(map[string]uint64)
	file, err := os.Open(curAbsRoot + "\\lastModTime.txt")
	if err != nil {
		if runtime.GOOS == "windows" {
			os.RemoveAll(luaAbsRoot)
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			s := strings.Split(line, "|")
			if len(s) == 2 {
				tm, _ := strconv.ParseUint(s[1], 10, 64)
				LastModifyTime[s[0]] = tm
			}
		}
	}
}

var sum int32

func myFunc(i interface{}) {
	n := i.(int32)
	atomic.AddInt32(&sum, n)
	//fmt.Printf("run with %d\n", n)
}
func main() {
	// 参数
	flag.StringVar(&xlsxPath, "i", "xlsx", "输入路径")
	flag.Parse()

	// walk
	var err error
	xlsxPath, err = filepath.Abs(xlsxPath)
	if err != nil {
		panic(err)
	}

	// 解析
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(10, func(i interface{}) {
		myFunc(i)
		wg.Done()
	})
	defer p.Release()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		_ = p.Invoke(int32(i))
	}
	wg.Wait()

	fmt.Printf("running goroutines: %d\n", p.Running())
	fmt.Printf("finish all tasks, result is %d\n", sum)

	f, err := excelize.OpenFile("task.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}

	xlsx := &Xlsx{
		Name:  "task",
		Descs: rows[0],
		Names: rows[1],
		Types: rows[2],
		Modes: rows[3],
		Rows:  rows,
	}
	xlsx.parseHeader()
	//xlsx.parseRows(rows)
	ExportLua(xlsx)

	// 错误输出
}
