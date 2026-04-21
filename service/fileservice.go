package service

import (
	"context"
	"encoding/json"
	"excelparser/core"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const configFileName = ".excelparser.json"

const (
	ExportStatusIdle      = iota // 空闲
	ExportStatusExporting        // 导出中
	ExportStatusSuccess          // 导出成功
	ExportStatusFailed           // 导出失败
	ExportStatusSkipped          // 导出跳过（文件无变化）
)

// 配置结构
type AppConfig struct {
	ConfigPath string   `json:"config_path"`
	OutputPath string   `json:"output_path"`
	I18nPath   string   `json:"i18n_path"`
	I18nLang   string   `json:"i18n_lang"`
	ServerFmts []string `json:"server_fmts"`
	ClientFmts []string `json:"client_fmts"`
}

type XlsxListItem struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	NeedParse bool   `json:"need_parse"`
}

type ExportProgressEvent struct {
	Stage    string   `json:"stage"`    // 阶段：start, finish, error, done
	Name     string   `json:"name"`     // 文件名
	Path     string   `json:"path"`     // 文件完整路径
	Status   int      `json:"status"`   // 导出状态：0=空闲, 1=导出中, 2=成功, 3=失败, 4=跳过
	Message  string   `json:"message"`  // 结果消息，成功时可为空，失败时包含错误信息
	Messages []string `json:"messages"` // 所有错误消息列表
	Seq      int64    `json:"seq"`      // 事件序列号，用于前端排序
}

//#region MARK: 配置管理

// 加载配置
func loadConfig() *AppConfig {
	config := &AppConfig{}
	if data, err := os.ReadFile(configFileName); err == nil {
		json.Unmarshal(data, config)

		// 将配置应用到核心
		core.GFlags.Path = config.ConfigPath
		core.GFlags.Output = config.OutputPath
		core.GFlags.I18nPath = config.I18nPath
		core.GFlags.I18nLang = config.I18nLang
		core.GFlags.Server = config.ServerFmts
		core.GFlags.Client = config.ClientFmts
	}
	return config
}

// 保存配置
func saveConfig(config *AppConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFileName, data, 0644)
}

//#endregion

//#region MARK: 帮助函数

func validatePath(path string) (string, os.FileInfo, error) {
	if len(path) == 0 {
		return "", nil, fmt.Errorf("路径不能为空")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return "", nil, err
	}

	return absPath, stat, nil
}

func buildOpenFileCmd(path string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/C", "start", "", path)
		// 隐藏命令行窗口
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
		return cmd, nil
	case "darwin":
		return exec.Command("open", path), nil
	default:
		return exec.Command("xdg-open", path), nil
	}
}

func buildOpenDirectoryCmd(path string, isDir bool) (*exec.Cmd, error) {
	dirPath := path
	if !isDir {
		dirPath = filepath.Dir(path)
	}

	switch runtime.GOOS {
	case "windows":
		if isDir {
			return exec.Command("explorer", dirPath), nil
		}
		return exec.Command("explorer", "/select,"+path), nil
	case "darwin":
		if isDir {
			return exec.Command("open", dirPath), nil
		}
		return exec.Command("open", "-R", path), nil
	default:
		return exec.Command("xdg-open", dirPath), nil
	}
}

//#endregion

//#region MARK: CORE 交互

func reloadPath(path string) error {
	core.Walked = false
	core.GFlags.Path = path
	err := core.WalkPath()
	if err != nil {
		return err
	}
	return nil
}

//#endregion

//#region MARK: Service

type FileService struct {
	config *AppConfig
}

// 服务启动时加载配置
func (s *FileService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.config = loadConfig()
	return nil
}

func (f *FileService) SaveConfig(config *AppConfig) error {
	return saveConfig(config)
}

func (f *FileService) GetConfig() (*AppConfig, error) {
	return loadConfig(), nil
}

func (f *FileService) GetXlsxList(path string) ([]XlsxListItem, error) {
	if len(path) > 0 {
		if err := reloadPath(path); err != nil {
			return nil, err
		}
	}

	items := make([]XlsxListItem, 0, len(core.XlsxList))
	for _, x := range core.XlsxList {
		needParse := len(x.GetNeedParse()) > 0
		items = append(items, XlsxListItem{Name: x.Name, Path: x.PathName, NeedParse: needParse})
	}
	return items, nil
}

func (f *FileService) OpenFile(path string) error {
	absPath, stat, err := validatePath(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("目标不是文件: %s", absPath)
	}

	cmd, err := buildOpenFileCmd(absPath)
	if err != nil {
		return err
	}

	return cmd.Start()
}

func (f *FileService) OpenFileDirectory(path string) error {
	absPath, stat, err := validatePath(path)
	if err != nil {
		return err
	}

	cmd, err := buildOpenDirectoryCmd(absPath, stat.IsDir())
	if err != nil {
		return err
	}

	return cmd.Start()
}

// 选择目录并更新核心路径
func (f *FileService) SelectDirectory(pathType int, title string) (string, error) {
	path, err := application.Get().Dialog.OpenFile().
		SetTitle(title).
		CanChooseDirectories(true).
		CanChooseFiles(false).
		CanCreateDirectories(true).
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}

	switch pathType {
	case 1:
		// 配置路径
		if err := reloadPath(path); err != nil {
			return "", err
		}
	case 2:
		// 输出路径
		core.GFlags.Output = path
	case 3:
		// i18n路径
		core.GFlags.I18nPath = path
	}

	return path, nil
}

// 设置导出选项
func (f *FileService) SetExportFlag(flagType int32, flagVal bool) {
	switch flagType {
	case 1:
		// 是否使用紧凑格式
		core.GFlags.Compact = flagVal
	case 2:
		// 是否使用格式化的JSON输出
		core.GFlags.Pretty = flagVal
	case 3:
		// 是否强制重新生成
		core.GFlags.Force = flagVal
	}
}

// 设置导出格式
func (f *FileService) SetExportFormat(target string, formats []string) {
	switch target {
	case "server":
		core.GFlags.Server = formats
	case "client":
		core.GFlags.Client = formats
	}
}

// 设置i18n语言
func (f *FileService) SetI18nLang(lang string) {
	core.GFlags.I18nLang = lang
}

// 开始导出
func (f *FileService) StartExport() error {
	var eventSeq int64
	emitProgress := func(payload ExportProgressEvent) {
		eventSeq++
		payload.Seq = eventSeq
		application.Get().Event.Emit("export-progress", payload)
	}

	err := core.Run(&core.ParseHandler{
		OnEvent: func(event *core.ParseEvent) {
			if event == nil || event.Xlsx == nil {
				return
			}

			// fmt.Println(event.Xlsx.Name, event.Status) // 调试输出
			switch event.Status {
			case "start":
				emitProgress(ExportProgressEvent{
					Stage:   "start",
					Name:    event.Xlsx.Name,
					Path:    event.Xlsx.PathName,
					Status:  ExportStatusExporting,
					Message: "-",
				})
			case "finish":
				status := ExportStatusSuccess
				message := ""
				var messages []string
				if len(event.Xlsx.Errors) > 0 {
					if event.Xlsx.Skipped {
						status = ExportStatusSkipped
					} else {
						status = ExportStatusFailed
					}
					message = event.Xlsx.Errors[0]
					messages = event.Xlsx.Errors
				}

				emitProgress(ExportProgressEvent{
					Stage:    "finish",
					Name:     event.Xlsx.Name,
					Path:     event.Xlsx.PathName,
					Status:   status,
					Message:  message,
					Messages: messages,
				})
			}
		},
	})
	if err != nil {
		emitProgress(ExportProgressEvent{
			Stage:   "error",
			Status:  ExportStatusFailed,
			Message: err.Error(),
		})
		return err
	}

	emitProgress(ExportProgressEvent{Stage: "done"})
	return nil
}

// 获取翻译列表
func (f *FileService) GetTranslationList() ([]string, error) {
	path := core.GFlags.I18nPath
	if len(path) == 0 {
		return nil, fmt.Errorf("i18n路径未设置")
	}

	// 读取目录下的文件夹列表
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	langs := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			langs = append(langs, entry.Name())
		}
	}
	return langs, nil
}

//#endregion
