package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

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

func exportExcel(format, mode string, xlsx *Xlsx) {
	var formater iFormater
	keyField := xlsx.RootField.Vals[0]
	if len(format) > 0 && keyField.isHitMode(mode) && len(xlsx.Rows) > 0 {
		formater = NewFormater(xlsx, format, mode)
		formater.formatRows()

		// write
		if len(xlsx.Errors) == 0 {
			xlsx.writeToFile(mode, format)
		}
	}
}

func parseExcel(i interface{}) {
	xlsx := i.(*Xlsx)

	startTime := time.Now()
	f, err := excelize.OpenFile(xlsx.PathName)
	if err == nil {
		defer func() {
			xlsx.Excel = nil
			f.Close()
		}()

		xlsx.Excel = f
		ok := xlsx.parseExcel()
		if ok && len(xlsx.Errors) == 0 {
			xlsx.Datas = make([]string, 0)
			exportExcel(FlagClient, "client", xlsx)
			exportExcel(FlagServer, "server", xlsx)
		}
	} else {
		xlsx.appendError("xlsx文件打开失败")
	}
	xlsx.TimeCost = getDurationMs(startTime)
	LoadingChan <- struct{}{}
}

func processMsg() {
	count := 0
	total := len(XlsxList)
	for range LoadingChan {
		count++
		percent := float32(count) / float32(total)
		fmt.Printf("\rProgress:[%-50s][%d%%]", strings.Repeat("█", int(percent*50)), int(percent*100))
	}
	fmt.Println()
}

func printResult() {
	results := make([]string, 0)
	results = append(results, Splitline)
	results = append(results, fmt.Sprintf(InfoFormat, "FileName", "Result"))

	if len(XlsxList) > 0 {
		sort.Slice(XlsxList, func(i, j int) bool { return len(XlsxList[i].Errors) > len(XlsxList[j].Errors) })
		for _, xlsx := range XlsxList {
			result := xlsx.collectResult()
			results = append(results, result...)
		}
	} else {
		results = append(results, Splitline)
		results = append(results, fmt.Sprintf(InfoFormat, "No files", "无需生成"))
	}
	results = append(results, Splitline)
	fmt.Println(strings.Join(results, "\n"))
}

func saveConvTime() {
	file := "lastModTime.txt"
	outFile, operr := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o666)
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
