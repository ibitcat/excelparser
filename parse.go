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
	err := xlsx.openExcel()
	if err == nil {
		rows := xlsx.Rows
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
	}
	xlsx.TimeCost = getDurationMs(startTime)
	LoadingChan <- struct{}{}
	//ExportLua(xlsx)
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
