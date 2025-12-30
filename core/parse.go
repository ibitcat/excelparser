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
	xlsxPath, err := CheckPathValid(GFlags.Path)
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
			fname := strings.TrimPrefix(path, xlsxPath+string(filepath.Separator)) // eg.: tpl/item@道具.xlsx
			dirname := strings.TrimSuffix(fname, f.Name())                         // eg.: tpl/
			fileName := getFileName(f.Name())                                      // item@道具
			outName := strings.SplitN(fileName, "@", 2)[0]                         // item
			task := &Xlsx{
				Idx:          len(XlsxList),
				Name:         dirname + f.Name(), // eg.: tpl/item@道具.xlsx
				PathName:     path,               // eg.: D:/project/excelparser/tpl/item@道具.xlsx
				FileName:     dirname + fileName, // eg.: tpl/item@道具
				DirName:      filepath.Dir(path), // eg.: D:/project/excelparser/tpl/
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

func FindXlsxByName(name string) *Xlsx {
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
		x := FindXlsxByName(k)
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

	for event := range EventChan {
		if event.Status == "finish" {
			result := event.Xlsx.collectResult(costFormat, infoFormat, splitline)
			fmt.Println(strings.Join(result, "\n"))
		}
	}
	fmt.Println(splitline)
}

type ParseEvent struct {
	Xlsx   *Xlsx
	Status string // "start" / "finish"
}

type ParseHandler struct {
	OnEvent func(*ParseEvent) // 统一的事件回调
}

func Run(handler *ParseHandler) error {
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
	EventChan = make(chan *ParseEvent, xlsxCount*2) // *2 因为每个任务有 start 和 finish 两个事件

	// 启动监听协程
	if handler != nil && handler.OnEvent != nil {
		go func() {
			for event := range EventChan {
				handler.OnEvent(event)
			}
		}()
	} else {
		go ProcessMsg()
	}

	// parse
	startTime := time.Now()
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(10, func(i interface{}) {
		xlsx := i.(*Xlsx)
		EventChan <- &ParseEvent{Xlsx: xlsx, Status: "start"}
		StartParse(xlsx)
		EventChan <- &ParseEvent{Xlsx: xlsx, Status: "finish"}
		wg.Done()
	})
	defer p.Release()

	for _, xlsx := range XlsxList {
		wg.Add(1)
		_ = p.Invoke(xlsx)
	}
	wg.Wait()

	close(EventChan)

	if len(GFlags.I18nLang) > 0 {
		SaveI18nXlsx(GFlags.I18nPath, GFlags.I18nLang)
	}

	ExportCost = GetDurationMs(startTime)
	return nil
}
