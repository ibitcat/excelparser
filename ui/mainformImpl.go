package ui

import (
	"excelparser/core"

	"github.com/ying32/govcl/vcl"
)

//::private::
type TForm1Fields struct {
	isInit    bool
	title     string
	aboutForm *TForm2
}

//#region 组件回调

func (f *TForm1) OnFormCreate(sender vcl.IObject) {
	// 一般在这里初始化自己的东东
	f.isInit = false
	f.title = "excelparser - Excel配置表解析工具"

	f.SetCaption(f.title)
	f.initMenu()
	f.initSelectEdit()
	f.initSelectBtn()
	f.initListView()
	f.initAboutForm()
	f.initComboBox()
	f.isInit = true
}

func (f *TForm1) OnSelectBtnClick(sender vcl.IObject) {
	btn := vcl.AsButton(sender)
	if f.SelectDirectoryDialog1.Execute() {
		selectedPath := f.SelectDirectoryDialog1.FileName()
		switch btn.Tag() {
		case 1:
			f.Edit1.SetText(selectedPath)
			f.refreshListView(selectedPath)
		case 2:
			f.Edit2.SetText(selectedPath)
		case 3:
			f.Edit3.SetText(selectedPath)
		}
	}
}

func (f *TForm1) OnPanel1Click(sender vcl.IObject) {
}

func (f *TForm1) OnComboBox1Change(sender vcl.IObject) {
	core.GFlags.Server = f.ComboBox1.Text()
}

func (f *TForm1) OnComboBox2Change(sender vcl.IObject) {
	core.GFlags.Client = f.ComboBox2.Text()
}

func (f *TForm1) OnCheckBox3Change(sender vcl.IObject) {
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

// 初始化选择路径编辑框
func (f *TForm1) initSelectEdit() {
	f.Edit1.SetText("请选择配置路径")
	f.Edit2.SetText("请选择导出路径")
	f.Edit3.SetText("请选择翻译路径")
}

// 初始化选择按钮
func (f *TForm1) initSelectBtn() {
	f.Button1.SetOnClick(f.OnSelectBtnClick)
	f.Button2.SetOnClick(f.OnSelectBtnClick)
	f.Button3.SetOnClick(f.OnSelectBtnClick)
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
	f.ComboBox1.SetItemIndex(0)

	f.ComboBox2.Items().Clear()
	f.ComboBox2.Items().Add("lua")
	f.ComboBox2.Items().Add("json")
	f.ComboBox2.SetItemIndex(0)

	core.GFlags.Server = f.ComboBox1.Text()
	core.GFlags.Server = f.ComboBox1.Text()
}

func (f *TForm1) refreshListView(path string) {
	core.GFlags.Path = path

	err := core.WalkPath()
	if err != nil {
		vcl.ShowMessage(err.Error())
		return
	}

	lv1 := f.ListView1
	lv1.Items().BeginUpdate()
	lv1.Items().Clear()
	for _, xlsx := range core.XlsxList {
		item := lv1.Items().Add()
		item.SetCaption(xlsx.Name)

		// 文件状态
		if xlsx.CanParse() {
			item.SubItems().Add("就绪")
		} else {
			item.SubItems().Add("-")
		}

		// 导出状态

		// 导出结果
	}
	lv1.Items().EndUpdate()
}

//#endregion

func (f *TForm1) OnCheckBox4Change(sender vcl.IObject) {
}
