# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

基于 Go 的 Excel 配置表解析器，用于游戏开发。将 `.xlsx` 文件转换为 Lua、JSON 或 C#（MessagePack 二进制）格式，支持服务端/客户端差异化导出和国际化。

## 构建与运行

```bash
# CLI 构建
go build -o excelparser

# GUI 构建（仅 Windows）
build_gui.cmd

# 运行示例
./excelparser --path=./xlsx --server=lua --client=json --indent
./excelparser --path=./xlsx --server=csharp --output=./out
./excelparser --force=true --path=./xlsx --output=./out --server=json --i18n=./locales --lang=en_US
```

无测试套件——通过对 `xlsx/` 示例文件运行并检查 `out/` 输出来手动测试。

## 架构

### 数据流

```
main.go → WalkPath() → XlsxList[]
       → startParse() [ants 库，10 协程并发处理]
           → parseExcel() → parseHeader() → checkFields() → checkRows()
           → NewFormater() → formatRows() → writeToFile()
       → SaveExportTime() → .excelparser.temp（YAML 缓存）
```

### 包结构

- **`core/`** — 所有解析逻辑：
  - `defines.go` — 类型常量（`TAny`…`TJson`）、全局变量（`GFlags`、`XlsxList`、`I18nMap`）
  - `flag.go` — CLI 参数；版本 2025.0.1
  - `xlsx.go` — 核心 `Xlsx` 结构体；`parseField()`（递归，第 182 行）、`checkRows()`、`writeToFile()`
  - `field.go` — `Field` 递归树；`isHitMode()`、`checkRow()`
  - `type.go` — `Type` 结构体；`parseType()`、`checkJsonVal()`、`defaultValue()`
  - `parse.go` — 编排逻辑；`WalkPath()`、`Run()`、事件系统
  - `formater.go` — `iFormater` 策略接口 + `NewFormater()` 工厂方法
  - `lua.go`、`json.go`、`csharp.go` — 各格式实现
  - `util.go` — `parseCompositeType()`、单元格坐标辅助函数、`ternary[T]()`
  - `i18n.go` — `.po` 文件加载与字符串替换
- **`ui/`** — Windows GUI（govcl 框架）；构建标签 `gui`
- **`main.go`** / **`main_gui.go`** — CLI 和 GUI 入口

### 关键数据结构

**`Type`**：递归类型节点 — `Kind int`、`Cap int`（数组容量）、`Ktype`/`Vtype *Type`（map/数组）、`Ftypes map[string]*Type`（结构体字段）、`I18n bool`、`Aname string`（结构体别名）。

**`Field`**：列树节点 — 嵌入 `*Type`，包含 `Parent *Field`、`Index int`、`Mode string`（`s`/`c`/`x`/`""`）、`Keys`/`Vals []*Field`。

**`Xlsx`**：单文件状态 — 持有 `*excelize.File`、已解析的表头行、`RootField *Field`、`Rows [][]string`、`Datas []string`（输出缓冲区）、`Errors []string`。

## Excel 格式约定

**固定 4 行表头（HeadLineNum=4，不可修改）：**
```
第 1 行：字段名
第 2 行：类型
第 3 行：导出模式：s=server，c=client，x=不导出，空=两端都导出
第 4 行：字段描述
```

**Sheet 命名：** `data`（横向表，多行数据，第一列为 ID），`vdata`（纵向表，全局配置，只有一行数据）。

**文件命名：** `道具@item.xlsx` → 导出为 `item.*`（@ 前为显示名，@ 后为导出名）。

**输出路径：** `{output}/{server|client}/{format}/`，例如 `./out/server/lua/`。

**类型语法：**
- 定长数组：`[3]int`，二维：`[2][3]int`
- Map：`map[int]string`，嵌套：`map[int]map[int]string`
- 具名结构体：`struct<TaskType>`，结构体子字段用点号分隔：`s1.a`、`s1.b`
- 带类型描述的 JSON：`json:[]int`、`json:map[int]string`
- 国际化字符串：`i18n`

## 关键实现细节

- **`parseField()` 递归**（`xlsx.go:182`）：利用合并单元格跨度（`getMergeRangeX`）推断嵌套结构——跨度宽度决定数组容量，map 的 key/value 按列配对。
- **输出缓冲**：`Datas []string` 积累输出片段；`appendData`/`replaceComma`/`replaceTail` 操作缓冲区；最后通过 `strings.Join` 一次性写入（`xlsx.go:540`）。
- **缓存机制**：`.excelparser.temp` YAML 记录每个文件的导出时间戳；未修改的文件跳过处理（除非使用 `--force`）。
- **错误累积上限**：每个文件最多报告 `MaxErrorCnt=6` 个错误，错误信息带单元格坐标如 `[B5]`。
- **Compact 模式**：仅适用于横向表 — `ternary(FlagCompact && !x.Vertical, ...)`。
- **goroutine 池**：`wg.Wait()` 必须在 `close(FinishChan)` 之前完成。

## 新增导出格式

1. 在新文件（`core/myformat.go`）中实现 `iFormater` 接口
2. 在 `core/formater.go` 的 `NewFormater()` 中添加 case
3. 在 `core/xlsx.go` 的 `writeToFile()` 中添加文件扩展名映射

## 新增类型

1. 在 `core/defines.go` 中添加 `const T*`
2. 在 `core/type.go` 的 `parseType()` 中添加解析逻辑
3. 在各格式实现的 `formatData()` 中添加处理分支
