package core

import (
	"regexp"
	"strings"
	"sync"

	"github.com/ibitcat/gotext"
	"github.com/xuri/excelize/v2"
)

//#region constants

// 数据类型定义
const (
	TNone   int = -1   // 非法类型
	TAny    int = iota // any
	TInt               // 有符号整数
	TUint              // 无符号整数
	TFloat             // 浮点数
	TBool              // 布尔型
	TString            // 字符串
	TArray             // 数组
	TMap               // map
	TStruct            // 结构体
	TJson              // json
)

// 配置表头行定义
const (
	NameLine int = iota + 1
	TypeLine
	ModeLine
	DescLine
)

//#endregion

//#region structs

// 导出选项
type Flags struct {
	Pretty   bool   // json格式化
	Force    bool   // 是否强制重新生成
	Compact  bool   // 是否紧凑导出
	Path     string // excel路径
	Output   string // 导出路径
	Server   string // server 导出格式
	Client   string // client 导出格式
	I18nPath string // 国际化配置路径
	I18nLang string // 国际化语言
}

// 结构体定义
type ExportInfo struct {
	Mode     string `json:"mode"`
	Format   string `json:"format"`
	LastTime uint64 `json:"lasttime"`
}

// 类型定义
type Type struct {
	Kind   int              // 类型定义
	Cap    int              // 容量（for array）
	I18n   bool             // 是否有国际化字符串(for string,json)
	Aname  string           // alias type name(for 具名结构体)
	Ktype  *Type            // 键类型(for map)
	Vtype  *Type            // 值类型(for map,array,json)
	Ftypes map[string]*Type // 字段类型(for 匿名结构体)
}

// 字段定义
type Field struct {
	*Type           // 字段数据类型
	Parent *Field   // 父字段
	Xlsx   *Xlsx    // 所属excel
	Index  int      // 字段索引
	Desc   string   // 字段描述
	Rname  string   // 原始字段名
	Name   string   // 字段名
	Mode   string   // 生成方式(s=server,c=client,x=none)
	Keys   []*Field // 键元素列表
	Vals   []*Field // 值元素列表
}

// Excel配置表结构体
type Xlsx struct {
	Idx          int            // 索引
	Name         string         // 文件名（带文件扩展名）
	FileName     string         // 文件名
	PathName     string         // 文件完整路径
	OutName      string         // 输出文件名(item@道具.xlsx, 输出为 item)
	SheetName    string         // 工作表名
	Vertical     bool           // 纵向表
	Excel        *excelize.File // 打开的excel文件句柄
	Names        []string       // 字段名列表
	Types        []string       // 类型列表
	Modes        []string       // 导出模式列表
	Descs        []string       // 字段描述列表
	RootField    *Field         // 根字段
	Rows         [][]string     // 合法的配置行
	Datas        []string       // 导出数据缓存
	Errors       []string       // 错误信息
	Exports      []ExportInfo   // 导出信息
	LastModified uint64         // 最后修改时间
	TimeCost     int            // 耗时
}

// Lua格式化器
type LuaFormater struct {
	*Xlsx
	line int
	mode string
}

// JSON格式化器
type JsonFormater struct {
	*Xlsx
	line int
	mode string
}

//#endregion

//#region variables

var (
	GFlags      Flags
	HeadLineNum = 4                                                // 配置表头行数
	IndentStr   map[int]string                                     // 缩进字符串映射
	ArrayRe     = regexp.MustCompile(`^\[(\d*?)\](.+)`)            // 数组类型正则表达式
	MapRe       = regexp.MustCompile(`^map\[(.+?)\](.+)`)          // map类型正则表达式
	BasicTypes  = []string{"int", "uint", "bool", "string", "var"} // 基本类型列表
	I18nMap     sync.Map                                           // 国际化字符串映射
	I18nLocale  *gotext.Locale                                     // 国际化对象
	XlsxList    []*Xlsx                                            // Excel配置表列表
	EventChan   chan *ParseEvent                                   // 解析事件通道
	MaxErrorCnt = 6                                                // 每个文件最大错误数
	ExportYaml  = ".excelparser.temp"                              // 导出记录文件名
	ExportCost  = 0                                                // 总耗时
)

//#endregion

func init() {
	IndentStr = make(map[int]string)
	for i := range 10 {
		IndentStr[i] = strings.Repeat(" ", i*2)
	}
}
