package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var MaxFileLen int

func walkPath(xlsxPath string) {
	XlsxList = make([]*Xlsx, 0)
	OutNames := make(map[string]string) // 导出名冲突检查
	err := filepath.Walk(xlsxPath, func(path string, f os.FileInfo, err error) error {
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
			fname := strings.TrimPrefix(path, xlsxPath+string(filepath.Separator))
			dirname := strings.TrimSuffix(fname, f.Name())
			fileName := getFileName(f.Name())
			outName := strings.SplitN(fileName, "@", 2)[0]
			task := &Xlsx{
				Name:         dirname + f.Name(),
				PathName:     path,
				FileName:     dirname + fileName,
				OutName:      outName,
				Errors:       make([]string, 0),
				TimeCost:     0,
				LastModified: modifyTime,
				Exports:      make([]ExportInfo, 0),
			}
			if _, ok := OutNames[outName]; ok {
				return errors.New(outName + " 导出名冲突")
			}
			OutNames[outName] = task.Name
			MaxFileLen = max(MaxFileLen, len(task.FileName))
			XlsxList = append(XlsxList, task)
		}
		return mErr
	})
	if err != nil {
		panic(err)
	}
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
	data, err := os.ReadFile("export.yaml")
	if err != nil {
		return
	}
	m := make(map[string][]ExportInfo)
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return
	}
	for k, v := range m {
		x := findXlsx(k)
		if x != nil {
			x.Exports = v
		}
	}
}

func startParse(i any) {
	xlsx := i.(*Xlsx)
	if xlsx.canParse() {
		startTime := time.Now()
		xlsx.exportExcel()
		xlsx.TimeCost = getDurationMs(startTime)
	} else {
		xlsx.appendError("文件未变化")
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
	cellLen := max(MaxFileLen, len("FileName")) + 1
	costFormat := fmt.Sprintf("%%-%ds| cost:%%-5dms, line:%%-6d", cellLen)
	infoFormat := fmt.Sprintf("%%-%ds| %%s", cellLen)
	splitline := fmt.Sprintf("%s+%s", strings.Repeat("-", cellLen), strings.Repeat("-", 50))

	results := make([]string, 0)
	results = append(results, splitline)
	results = append(results, fmt.Sprintf(infoFormat, "FileName", "Result"))

	if len(XlsxList) > 0 {
		sort.Slice(XlsxList, func(i, j int) bool { return len(XlsxList[i].Errors) > len(XlsxList[j].Errors) })
		for _, xlsx := range XlsxList {
			result := xlsx.collectResult(costFormat, infoFormat, splitline)
			results = append(results, result...)
		}
	} else {
		results = append(results, splitline)
		results = append(results, fmt.Sprintf(infoFormat, "No files", "无需生成"))
	}
	results = append(results, splitline)
	fmt.Println(strings.Join(results, "\n"))
}

func saveConvTime() {
	file := "export.yaml"
	outFile, operr := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o666)
	if operr != nil {
		fmt.Println("创建[export.yaml]文件错误")
	}
	defer outFile.Close()

	m := make(map[string][]ExportInfo)
	for _, xlsx := range XlsxList {
		m[xlsx.Name] = xlsx.Exports
	}

	// save time
	d, err := yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	outFile.WriteString(string(d))
	outFile.Sync()
}
