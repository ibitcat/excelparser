# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

基于 Go 的 Excel 配置表解析器，用于游戏开发。支持将 `.xlsx` 文件转换为 Lua 或 JSON 格式，并提供两种运行模式：

- **GUI 模式**（默认）：基于 Wails v3 + Vue 3 的桌面应用
- **CLI 模式**（build tag `cli`）：纯命令行工具

## 构建与运行

### 依赖工具

- Go 1.25+
- [Wails v3](https://v3.wails.io/)（GUI 开发）
- [Task](https://taskfile.dev/)（任务运行器）
- Node.js + npm（前端依赖）

### 常用命令

```bash
# GUI 开发模式（热重载）
task dev

# 构建 GUI 版本（Windows）
task build

# 构建 CLI 版本
go build -tags cli -o excelparser_cli.exe

# 运行 CLI
./excelparser_cli --path=./xlsx --server=lua --client=json --indent --force

# 前端依赖安装
cd frontend && npm install

# 生成 Wails 绑定（Go → TS 类型）
wails3 generate bindings
```

### CLI 参数

| 参数        | 说明                                                 |
| ----------- | ---------------------------------------------------- |
| `--path`    | xlsx 配置文件目录                                    |
| `--output`  | 输出目录（默认 `.`）                                 |
| `--server`  | 服务端格式，如 `--server=lua` 或 `--server=lua,json` |
| `--client`  | 客户端格式，如 `--client=json`                       |
| `--indent`  | JSON 美化输出                                        |
| `--force`   | 强制重新导出所有文件                                 |
| `--compact` | 紧凑模式（减少文件大小）                             |
| `--i18n`    | 国际化翻译路径                                       |
| `--lang`    | 目标语言（如 `en_US`）                               |

输出路径格式：`{output}/{server|client}/{format}/`

## 核心架构

### 数据流

1. `main.go`（GUI）或 `main_cli.go`（CLI）→ 调用 `core.Run()`
2. `core.Run()` → `WalkPath()` 扫描 xlsx 目录 → 构建 `XlsxList`
3. ants 协程池（10 goroutine）并发处理每个文件
4. 每个文件：`parseExcel` → `parseHeader` → `checkFields` → `checkRows` → `formatRows` → `writeToFile`
5. `export.yaml` 缓存上次导出时间，跳过未修改文件

### 关键组件

- **[core/xlsx.go](core/xlsx.go)**：`Xlsx` 核心结构体，持有文件句柄和解析状态；`parseField` 递归解析嵌套类型
- **[core/field.go](core/field.go)**：`Field` 递归字段树，表示 array/map/struct 嵌套
- **[core/type.go](core/type.go)**：类型系统，10 种基础类型 + 复合类型；`checkJsonVal` 递归验证 JSON 结构
- **[core/formater.go](core/formater.go)**：`iFormater` 策略接口，由 `LuaFormater` 和 `JsonFormater` 实现
- **[core/defines.go](core/defines.go)**：全局标志 `GFlags`、常量定义
- **[service/fileservice.go](service/fileservice.go)**：Wails 服务层，桥接 Go 后端与 Vue 前端
- **[frontend/src/](frontend/src/)**：Vue 3 前端，组件在 `components/` 目录

### GUI 架构（Wails v3）

- `main.go` 注册 `FileService` 为 Wails 服务
- `service.FileService` 方法自动绑定为前端可调用的 TypeScript API
- 导出进度通过 `export-progress` 事件从 Go 推送到 Vue 前端
- 配置持久化到 `.excelparser.json`（工作目录）
- 注意 frontend/bindings/ 目录下的自动生成文件，包含 Go → TS 类型定义，不要修改

## Excel 表头格式

每个 Excel 文件使用固定 **4 行表头**：

```
Row 1 (NameLine):  字段名
Row 2 (TypeLine):  字段类型
Row 3 (ModeLine):  导出模式 s=server, c=client, x=不导出, 空=both
Row 4 (DescLine):  字段描述
```

### Sheet 命名

- `data`：横向表（多行数据，第一列为 ID）
- `vdata`：纵向表（单行全局配置）

### 类型系统

- 基础类型：`int`、`uint`、`float`、`bool`、`string`、`json`、`i18n`
- 定长数组：`[3]int`，二维：`[2][3]int`
- Map：`map[int]string`，嵌套：`map[int]map[int]string`
- 结构体：`struct<TypeName>`（具名）或 `struct`（匿名）
- JSON 动态类型：`json:[]int`（尖括号内描述真实结构）

### 文件命名

`道具@item.xlsx` → 输出为 `item.lua` / `item.json`（`@` 前为显示名，后为导出文件名）

## 关键实现细节

### 字段解析（[core/xlsx.go](core/xlsx.go)）

`parseField` 使用合并单元格（`getMergeRangeX`）推断嵌套结构：

- 合并单元格跨度决定数组容量
- Map 的 key/value 按列对出现
- Struct 子字段用点号分隔（如 `s1.a`、`s1.b`）

### 数据缓存

`Datas []string` 积累输出片段，最终 `strings.Join` 一次性写文件。`appendData`/`replaceComma`/`replaceTail` 操作此缓存。

### 国际化

读取 `.po` 文件（gotext 库）→ 对标记 `I18n=true` 的字段替换值 → 记录引用位置用于 `.po` 更新。

## 扩展指南

### 新增输出格式

1. 在 [core/formater.go](core/formater.go) 实现 `iFormater` 接口
2. 在 `NewFormater` switch 添加 case
3. 在 `writeToFile` 添加文件扩展名映射

### 新增类型

1. 在 [core/type.go](core/type.go) 添加 `const T*` 常量
2. 在 `parseType` 添加解析逻辑
3. 在各 formatter 的 `formatData` 添加处理分支

## 常见陷阱

- **`HeadLineNum=4` 不可改**：整个解析系统依赖此常量
- **字段索引基于合并后的列**：复杂类型占多列，索引需累加
- **纵向表不支持 compact**：`vdata` sheet 的 `FlagCompact` 被忽略
- **错误上限 `MaxErrorCnt=6`**：防止错误日志爆炸
- **goroutine 顺序**：`wg.Wait()` 后才能 `close(FinishChan)`，顺序不能颠倒
- **GUI 与 CLI 互斥**：通过 build tag `cli` 区分，两者共享 `core` 包但入口不同
