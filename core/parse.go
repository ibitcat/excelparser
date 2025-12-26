package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ibitcat/gotext"
	"github.com/panjf2000/ants/v2"
	"gopkg.in/yaml.v3"
)

var (
	MaxFileLen int  // 最大文件名长度
	Walked     bool = false
)

func WalkPath() error {
	xlsxPath, err := CheckPathVaild(GFlags.Path)
	if err != nil {
		return err
	}
	if Walked {
		return nil
	}

	XlsxList = make([]*Xlsx, 0)
	OutNames := make(map[string]string) // 导出名冲突检查
	err = filepath.Walk(xlsxPath, func(path string, f os.FileInfo, err error) error {
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

	Walked = true
	LoadExportTime()
	return err
}

func findXlsx(name string) *Xlsx {
	for _, x := range XlsxList {
		if x.Name == name {
			return x
		}
	}
	return nil
}

func LoadExportTime() {
	data, err := os.ReadFile(ExportYaml)
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

func SaveExportTime() {
	outFile, operr := os.OpenFile(ExportYaml, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o666)
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

func StartParse(xlsx *Xlsx) {
	if xlsx.CanParse() {
		startTime := time.Now()
		xlsx.exportExcel()
		xlsx.TimeCost = GetDurationMs(startTime)
	} else {
		xlsx.appendError("文件未变化")
	}
}

func ProcessMsg() {
	cellLen := max(MaxFileLen, len("FileName")) + 1
	costFormat := fmt.Sprintf("%%-%ds| cost:%%-5dms, line:%%-6d", cellLen)
	infoFormat := fmt.Sprintf("%%-%ds| %%s", cellLen)
	splitline := fmt.Sprintf("%s+%s", strings.Repeat("-", cellLen), strings.Repeat("-", 50))

	// header
	fmt.Println(splitline)
	fmt.Printf(infoFormat, "FileName", "Result\n")

	for xlsx := range LoadingChan {
		result := xlsx.collectResult(costFormat, infoFormat, splitline)
		fmt.Println(strings.Join(result, "\n"))
		// percent := float32(count) / float32(total)
		// fmt.Printf("\rProgress:[%-50s][%d%%]", strings.Repeat("█", int(percent*50)), int(percent*100))
	}
	fmt.Println(splitline)
}

func Run() error {
	// i18n output path
	if len(GFlags.I18nLang) > 0 {
		I18nLocale = gotext.NewLocale(GFlags.I18nPath, GFlags.I18nLang)
		I18nLocale.AddDomain("default")
		I18nLocale.ClearAllRefs()
	}

	err := WalkPath()
	if err != nil {
		return err
	}

	xlsxCount := len(XlsxList)
	if xlsxCount == 0 {
		return errors.New("no valid .xlsx files found")
	}

	defer SaveExportTime()
	LoadingChan = make(chan *Xlsx, xlsxCount)
	go ProcessMsg()

	// parse
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(10, func(i interface{}) {
		xlsx := i.(*Xlsx)
		StartParse(xlsx)
		LoadingChan <- xlsx
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

	if len(GFlags.I18nLang) > 0 {
		SaveI18nXlsx(GFlags.I18nPath, GFlags.I18nLang)
	}

	return nil
}
