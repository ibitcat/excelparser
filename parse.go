package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

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
		task := &Xlsx{
			Name:         f.Name(),
			PathName:     path,
			FileName:     getFileName(f.Name()),
			Errors:       make([]string, 0),
			TimeCost:     0,
			LastModified: modifyTime,
			Exports:      make([]*ExportInfo, 0),
		}
		XlsxList = append(XlsxList, task)
	}
	return mErr
}

func findXlsx(name string) *Xlsx {
	for _, x := range XlsxList {
		if x.Name == name {
			return x
		}
	}
	return nil
}

func loadExportLog() {
	file, err := os.Open("export.log")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			s := strings.SplitN(line, ",", 2)
			if len(s) == 2 {
				x := findXlsx(s[0])
				if x != nil {
					json.Unmarshal([]byte(s[1]), &x.Exports)
				}
			}
		}
	}
}

func startParse(i interface{}) {
	xlsx := i.(*Xlsx)
	if xlsx.canParse() {
		startTime := time.Now()
		xlsx.exportExcel()
		xlsx.TimeCost = getDurationMs(startTime)
	} else {
		xlsx.appendError("无需生成")
	}
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
	file := "export.log"
	outFile, operr := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o666)
	if operr != nil {
		fmt.Println("创建[export.log]文件错误")
	}
	defer outFile.Close()

	modTimes := make([]string, 0, len(XlsxList))
	for _, xlsx := range XlsxList {
		es, err := json.Marshal(xlsx.Exports)
		if err == nil {
			modTimes = append(modTimes, xlsx.Name+","+string(es))
		}
	}

	// save time
	outFile.WriteString(strings.Join(modTimes, "\n"))
	outFile.Sync()
}
