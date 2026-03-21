package ui

import (
	"encoding/json"
	"excelparser/core"
	"fmt"
	"os"

	"github.com/ying32/govcl/vcl"
	"github.com/ying32/govcl/vcl/rtl"
	"github.com/ying32/govcl/vcl/types"
)

// 配置结构
type AppConfig struct {
	ConfigPath string `json:"config_path"`
	OutputPath string `json:"output_path"`
	I18nPath   string `json:"i18n_path"`
}

var exporting bool = false

const configFileName = ".excelparser.config"

// 加载配置
func loadConfig() *AppConfig {
	config := &AppConfig{}
	if data, err := os.ReadFile(configFileName); err == nil {
		json.Unmarshal(data, config)
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

//::private::
type TForm1Fields struct {
	isInit    bool
	title     string
	aboutForm *TForm2
	config    *AppConfig
}

//#region 组件回调

func (f *TForm1) OnFormCreate(sender vcl.IObject) {
	// 一般在这里初始化自己的东东
	f.isInit = false
	f.title = "excelparser - Excel配置表解析工具"

	config := loadConfig()
	f.config = config
	f.SetCaption(f.title)
	f.SetBorderStyle(types.BsSingle) // 固定窗口大小
	f.initMenu()
	f.initSelectEdit()
	f.initListView()
	f.initAboutForm()
	f.initComboBox()
	if len(config.ConfigPath) > 0 {
		f.refreshListView(config.ConfigPath)
	}
	f.isInit = true
}

func (f *TForm1) OnSelectBtnClick(sender vcl.IObject) {
	btn := vcl.AsButton(sender)
	if f.SelectDirectoryDialog1.Execute() {
		selectedPath := f.SelectDirectoryDialog1.FileName()
		if _, err := core.CheckPathValid(selectedPath); err != nil {
			vcl.ShowMessage("⚠️ 所选路径不存在,请重新选择!")
			return
		}

		switch btn.Tag() {
		case 1:
			f.Edit1.SetText(selectedPath)
			f.refreshListView(selectedPath)
			f.config.ConfigPath = selectedPath
		case 2:
			f.Edit2.SetText(selectedPath)
			f.config.OutputPath = selectedPath
		case 3:
			f.Edit3.SetText(selectedPath)
			f.config.I18nPath = selectedPath
			f.refreshI18nComboBox(selectedPath)
		}

		saveConfig(f.config)
	}
}

func (f *TForm1) OnButton1Click(sender vcl.IObject) {
	f.OnSelectBtnClick(sender)
}

func (f *TForm1) OnButton2Click(sender vcl.IObject) {
	f.OnSelectBtnClick(sender)
}

func (f *TForm1) OnButton3Click(sender vcl.IObject) {
	f.OnSelectBtnClick(sender)
}

// 开始导出
func (f *TForm1) OnButton4Click(sender vcl.IObject) {
	f.StartExport()
}

func (f *TForm1) OnButton5Click(sender vcl.IObject) {
	f.refreshListView(f.Edit1.Text())
}

func (f *TForm1) OnComboBox1Change(sender vcl.IObject) {
	core.GFlags.Server = f.ComboBox1.Text()
}

func (f *TForm1) OnComboBox2Change(sender vcl.IObject) {
	core.GFlags.Client = f.ComboBox2.Text()
}

func (f *TForm1) OnCheckBoxAllChange(sender vcl.IObject) {
	core.GFlags.Force = f.CheckBoxAll.Checked()
}

func (f *TForm1) OnCheckBoxCompactChange(sender vcl.IObject) {
	if f.CheckBoxCompact.Checked() {
		f.CheckBoxPretty.SetChecked(false)
	}
}

func (f *TForm1) OnCheckBoxPrettyChange(sender vcl.IObject) {
	if f.CheckBoxPretty.Checked() {
		f.CheckBoxCompact.SetChecked(false)
	}
}

// 右键菜单 - 打开文件所在目录
func (f *TForm1) OnMenuItem1Click(sender vcl.IObject) {
	lv := f.ListView1
	if lv.Selected() == nil {
		return
	}

	item := lv.Selected()
	fileName := item.Caption()
	xlsx := core.FindXlsxByName(fileName)
	if xlsx == nil {
		return
	}
	rtl.SysOpen(xlsx.DirName)
}

// 右键菜单 - 打开文件
func (f *TForm1) OnMenuItem2Click(sender vcl.IObject) {
	lv := f.ListView1
	if lv.Selected() == nil {
		return
	}

	// 获取选中的文件
	item := lv.Selected()
	fileName := item.Caption()
	xlsx := core.FindXlsxByName(fileName)
	// fmt.Println(f.Edit1.Text(), fileName)
	if xlsx != nil {
		// 打开文件
		rtl.SysOpen(f.Edit1.Text() + "\\" + fileName)
	}
}

// ListView 双击事件 - 显示错误详情
func (f *TForm1) OnListViewDblClick(sender vcl.IObject) {
	lv := f.ListView1
	if lv.Selected() == nil {
		return
	}

	item := lv.Selected()
	fileName := item.Caption()
	xlsx := core.FindXlsxByName(fileName)
	if xlsx == nil {
		return
	}

	// 检查是否有错误
	if len(xlsx.Errors) > 0 {
		errorMsg := fmt.Sprintf("文件: %s\n\n错误详情:\n\n", fileName)
		for i, err := range xlsx.Errors {
			errorMsg += fmt.Sprintf("%d. %s\n", i+1, err)
		}
		vcl.ShowMessage(errorMsg)
	}
}

//#endregion

//#region 逻辑处理

// 初始化菜单
func (f *TForm1) initMenu() {
	menu := vcl.NewMenuItem(f)
	menu.SetCaption("文件(&F)")
	subMenu := vcl.NewMenuItem(f)
	subMenu.SetCaption("退出(&X)")
	menu.Add(subMenu)
	subMenu.SetOnClick(func(vcl.IObject) {
		f.Close()
	})
	f.MainMenu1.Items().Add(menu)

	menu = vcl.NewMenuItem(f)
	menu.SetCaption("关于(&A)")
	subMenu = vcl.NewMenuItem(f)
	subMenu.SetCaption("帮助(&H)")
	menu.Add(subMenu)
	subMenu.SetOnClick(func(vcl.IObject) {
		f.aboutForm.ShowModal()
	})
	f.MainMenu1.Items().Add(menu)
}

func (f *TForm1) initSelectEdit() {
	config := f.config

	if config.ConfigPath != "" {
		f.Edit1.SetText(config.ConfigPath)
	} else {
		f.Edit1.SetText("请选择配置路径")
	}

	if config.OutputPath != "" {
		f.Edit2.SetText(config.OutputPath)
	} else {
		f.Edit2.SetText("请选择导出路径")
	}

	if config.I18nPath != "" {
		f.Edit3.SetText(config.I18nPath)
	} else {
		f.Edit3.SetText("请选择翻译路径")
	}
}

// 初始化列表视图
func (f *TForm1) initListView() {
	lv1 := f.ListView1
	lv1.Columns().Clear()
	col := lv1.Columns().Add()
	col.SetCaption("文件名")
	col.SetAutoSize(true)
	col.SetWidth(lv1.ClientWidth() - 400)

	col = lv1.Columns().Add()
	col.SetCaption("文件状态")
	col.SetWidth(100)
	col.SetMaxWidth(100)

	col = lv1.Columns().Add()
	col.SetCaption("导出状态")
	col.SetWidth(100)
	col.SetMaxWidth(100)

	col = lv1.Columns().Add()
	col.SetCaption("导出结果")
	col.SetWidth(200)
	col.SetMaxWidth(200)

	lv1.SetDoubleBuffered(true)

	// 启用整行选择
	lv1.SetRowSelect(true)

	// 关联右键菜单
	lv1.SetPopupMenu(f.PopupMenu1)

	// 双击事件
	lv1.SetOnDblClick(f.OnListViewDblClick)
}

// 初始化关于窗口
func (f *TForm1) initAboutForm() {
	f.aboutForm = NewForm2(f)
	f.aboutForm.EnabledMaximize(false)
	f.aboutForm.EnabledMinimize(false)
}

func (f *TForm1) initComboBox() {
	f.ComboBox2.Items().Clear()
	f.ComboBox1.Items().Add("lua")
	f.ComboBox1.Items().Add("json")
	f.ComboBox1.Items().Add("csharp")
	f.ComboBox1.SetItemIndex(0)
	f.ComboBox1.SetStyle(types.CsDropDownList)
	f.ComboBox1.SetSelStart(int32(len(f.ComboBox1.Text())))

	f.ComboBox2.Items().Clear()
	f.ComboBox2.Items().Add("lua")
	f.ComboBox2.Items().Add("json")
	f.ComboBox2.Items().Add("csharp")
	f.ComboBox2.SetItemIndex(0)
	f.ComboBox2.SetStyle(types.CsDropDownList)
	f.ComboBox2.SetSelStart(int32(len(f.ComboBox2.Text())))

	f.ComboBox3.SetStyle(types.CsDropDownList)
	// 如果有配置的 i18n 路径，扫描文件夹列表
	if f.config.I18nPath != "" {
		f.refreshI18nComboBox(f.config.I18nPath)
	} else {
		f.ComboBox3.Items().Clear()
		f.ComboBox3.Items().Add("")
		f.ComboBox3.SetItemIndex(0)
	}
}

func (f *TForm1) refreshI18nComboBox(path string) {
	if path == "" {
		return
	}

	// 读取目录下的文件夹列表
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	f.ComboBox3.Items().Clear()
	for _, entry := range entries {
		if entry.IsDir() {
			f.ComboBox3.Items().Add(entry.Name())
		}
	}

	if f.ComboBox3.Items().Count() > 0 {
		f.ComboBox3.SetItemIndex(0)
	} else {
		f.ComboBox3.Items().Add("")
		f.ComboBox3.SetItemIndex(0)
	}
}

func (f *TForm1) refreshListView(path string) {
	// 检查path是否合法
	if path == "" {
		vcl.ShowMessage("⚠️ 请先选择有效的配置路径!")
		return
	}
	// 检查路径是否存在
	if _, err := core.CheckPathValid(path); err != nil {
		vcl.ShowMessage("⚠️ 所选路径不存在,请重新选择!")
		return
	}

	core.Walked = false
	core.GFlags.Path = path
	err := core.WalkPath()
	if err != nil {
		vcl.ShowMessage(err.Error())
		return
	}

	changeCount := 0
	lv1 := f.ListView1
	lv1.Items().BeginUpdate()
	lv1.Items().Clear()
	for _, xlsx := range core.XlsxList {
		item := lv1.Items().Add()
		item.SetCaption(xlsx.Name)

		// 文件状态
		if xlsx.CanParse() {
			changeCount++
			item.SubItems().Add("就绪")
		} else {
			item.SubItems().Add("-")
		}

		// 导出状态
		item.SubItems().Add("-")

		// 导出结果
		item.SubItems().Add("-")
	}
	lv1.Items().EndUpdate()

	f.StatusBar1.Panels().Items(0).SetText(fmt.Sprintf("文件数量：%d", int32(len(core.XlsxList))))
	f.StatusBar1.Panels().Items(1).SetText(fmt.Sprintf("有变化的文件数量：%d", changeCount))
}

func (f *TForm1) updateListViewItem(ch <-chan *core.Xlsx) {
	for xlsx := range ch {
		vcl.ThreadSync(func() {
			lv1 := f.ListView1
			item := lv1.Items().Item(int32(xlsx.Idx))
			item.SubItems().SetStrings(0, "导出中...")
		})
	}
}

func (f *TForm1) updateItemStatus(xlsx *core.Xlsx, idx int32, status string) {
	lv := f.ListView1
	item := lv.Items().Item(int32(xlsx.Idx))
	item.SubItems().SetStrings(idx, status)
}

func (f *TForm1) switchEnable(enable bool) {
	f.Button1.SetEnabled(enable)
	f.Button2.SetEnabled(enable)
	f.Button3.SetEnabled(enable)
	f.Button4.SetEnabled(enable)
	f.Button5.SetEnabled(enable)
	f.ComboBox1.SetEnabled(enable)
	f.ComboBox2.SetEnabled(enable)
	f.ComboBox3.SetEnabled(enable)
	f.CheckBoxAll.SetEnabled(enable)
	f.CheckBoxCompact.SetEnabled(enable)
	f.CheckBoxPretty.SetEnabled(enable)
}

func (f *TForm1) StartExport() {
	// ready := make(chan struct{})
	// go func() {
	// 	close(ready) // 关闭 channel 表示已进入协程
	// 	f.updateListViewItem()
	// }()
	// <-ready // 等待 channel 关闭

	fmt.Println("开始导出配置...")
	i18nPath := f.Edit3.Text()
	core.GFlags.Path = f.Edit1.Text()
	core.GFlags.Output = f.Edit2.Text()
	core.GFlags.Compact = f.CheckBoxCompact.Checked()
	core.GFlags.Pretty = f.CheckBoxPretty.Checked()
	core.GFlags.Force = f.CheckBoxAll.Checked()
	core.GFlags.Server = f.ComboBox1.Text()
	core.GFlags.Client = f.ComboBox2.Text()
	if _, err := core.CheckPathValid(core.GFlags.Path); err != nil {
		vcl.ShowMessage("⚠️ 配置路径无效,请重新选择!")
		return
	}
	if _, err := core.CheckPathValid(core.GFlags.Output); err != nil {
		vcl.ShowMessage("⚠️ 导出路径无效,请重新选择!")
		return
	}
	if i18nAbsPath, err := core.CheckPathValid(i18nPath); err == nil {
		core.GFlags.I18nPath = i18nAbsPath
		core.GFlags.I18nLang = f.ComboBox3.Text()
	}

	f.switchEnable(false)
	f.StatusBar1.Panels().Items(2).SetText("导出中...")

	go func() {
		if exporting {
			vcl.ShowMessage("⚠️ 正在导出中，请稍后...")
			return
		}
		exporting = true
		core.Run(&core.ParseHandler{
			OnEvent: func(event *core.ParseEvent) {
				vcl.ThreadSync(func() {
					switch event.Status {
					case "start":
						f.updateItemStatus(event.Xlsx, 1, "🔄 导出中...")
					case "finish":
						f.updateItemStatus(event.Xlsx, 1, "✓ 完成")
						if len(event.Xlsx.Errors) > 0 {
							f.updateItemStatus(event.Xlsx, 2, "❌ "+event.Xlsx.Errors[0])
						} else {
							f.updateItemStatus(event.Xlsx, 2, "✓ 成功")
						}
					}
				})
			},
		})
		exporting = false
		f.switchEnable(true)
		f.StatusBar1.Panels().Items(2).SetText(fmt.Sprintf("总耗时(ms)：%d", core.ExportCost))
	}()
}

//#endregion
