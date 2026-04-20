package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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
			fname := strings.TrimPrefix(path, xlsxPath+string(filepath.Separator)) // eg.: tpl/D道具表@item.xlsx
			dirname := strings.TrimSuffix(fname, f.Name())                         // eg.: tpl/
			fileName := getFileName(f.Name())                                      // eg.: D道具表@item
			outName := fileName                                                    // eg.: item
			if s := strings.SplitN(outName, "@", 2); len(s) > 1 {
				outName = s[1]
			}

			task := &Xlsx{
				// Idx:          len(XlsxList),
				Name:         dirname + f.Name(), // eg.: tpl/D道具表@item.xlsx
				PathName:     path,               // eg.: D:/project/excelparser/tpl/D道具表@item.xlsx
				FileName:     dirname + fileName, // eg.: tpl/D道具表@item
				DirName:      filepath.Dir(path), // eg.: D:/project/excelparser/tpl/
				OutName:      outName,            // eg.: item
				Errors:       make([]string, 0),
				TimeCost:     0,
				LastModified: modifyTime,
				Exports:      make([]ExportInfo, 0),
			}
			if _, ok := OutNames[outName]; ok {
				return errors.New(outName + " 导出名冲突: " + task.Name + " 和 " + OutNames[outName])
			}
			OutNames[outName] = task.Name
			MaxFileLen = max(MaxFileLen, len(task.FileName))
			XlsxList = append(XlsxList, task)
		}
		return mErr
	})

	// 按 Name 排序 XlsxList
	sort.Slice(XlsxList, func(i, j int) bool {
		return XlsxList[i].Name < XlsxList[j].Name
	})
	// 重新设置 Idx 保证顺序正确
	for i, x := range XlsxList {
		x.Idx = i
	}

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
	// 清空 Errors，以免上次的错误影响本次结果
	xlsx.Errors = xlsx.Errors[:0]
	xlsx.Skipped = false
	needParse := xlsx.GetNeedParse()
	if len(needParse) == 0 {
		xlsx.Skipped = true
		xlsx.appendError("文件未变化")
		return
	}

	// 解析文件
	xlsx.exportExcel(needParse)
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

	// 过滤指定文件
	parseList := XlsxList
	if len(GFlags.Files) > 0 {
		parseList = make([]*Xlsx, 0, len(GFlags.Files))
		for _, x := range XlsxList {
			for _, f := range GFlags.Files {
				if x.Name == f || filepath.Base(x.PathName) == f {
					parseList = append(parseList, x)
					break
				}
			}
		}
		if len(parseList) == 0 {
			return errors.New("no matching .xlsx files found for --files")
		}
	}

	xlsxCount := len(parseList)
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

	for _, xlsx := range parseList {
		wg.Add(1)
		_ = p.Invoke(xlsx)
	}
	wg.Wait()

	// 生成/更新 GameTableProxy.cs（csharp 模式下）
	sep := string(filepath.Separator)
	for _, format := range GFlags.Server {
		if format == "csharp" {
			outdir := GFlags.Output + sep + "server" + sep + "csharp" + sep
			UpdateGameTableProxy(outdir, "server")
		}
	}
	for _, format := range GFlags.Client {
		if format == "csharp" {
			outdir := GFlags.Output + sep + "client" + sep + "csharp" + sep
			UpdateGameTableProxy(outdir, "client")
		}
	}

	close(EventChan)

	if len(GFlags.I18nLang) > 0 {
		SaveI18nXlsx(GFlags.I18nPath, GFlags.I18nLang)
	}

	ExportCost = GetDurationMs(startTime)
	return nil
}
