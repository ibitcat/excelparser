# excelparser AI Coding Instructions

## 项目概述

基于 Go 的 Excel 配置表解析器,支持将 `.xlsx` 文件转换为 Lua 或 JSON 格式。用于游戏开发配置管道,支持服务端/客户端差异化导出和国际化。

## 核心架构

### 数据流

1. `main.go` → 遍历 xlsx 目录 (`walkPath`) → 构建 `XlsxList`
2. 并发池 (ants) 10 协程处理每个文件 (`startParse`)
3. 解析流程: `parseExcel` → `parseHeader` → `checkFields` → `checkRows` → `formatRows` → `writeToFile`
4. 缓存机制: `export.yaml` 记录每个文件的最后导出时间,避免重复处理

### 关键组件

-   **Xlsx** ([xlsx.go](xlsx.go)): 核心数据结构,持有 Excel 文件句柄和解析状态
-   **Field** ([field.go](field.go)): 递归字段树,表示嵌套类型 (array/map/struct)
-   **Type** ([type.go](type.go)): 类型系统 (10 种基础类型 + 复合类型)
-   **Formatter** ([formater.go](formater.go)): 策略模式接口,由 `LuaFormater` 和 `JsonFormater` 实现

## 关键约定

### Excel 表头格式 (固定 4 行)

```
Row 1: 字段名 (NameLine)
Row 2: 字段类型 (TypeLine)
Row 3: 导出模式 (ModeLine): s=server, c=client, x=不导出, 空=both
Row 4: 字段描述 (DescLine)
```

### Sheet 命名规则

-   `data`: 横向表 (多行配置,第一列为 ID)
-   `vdata`: 纵向表 (全局配置,单行数据)

### 类型表示

-   定长数组: `[3]int`
-   二维数组: `[2][3]int`
-   Map: `map[int]string`, `map[int]map[int]string` (支持嵌套)
-   结构体: `struct<TaskType>` (具名) 或 `struct` (匿名)
-   JSON 动态类型: `json<[]int>` (在尖括号中描述 JSON 真实结构)
-   i18n 字符串: `i18n` (标记需要翻译的字段)

### 文件命名

-   Excel: `道具@item.xlsx` → 输出为 `item.lua` / `item.json` (@ 前为显示名,后为导出名)
-   输出路径: `{output}/{mode}/{format}/` (例如 `./server/lua/`)

## 开发工作流

### 构建与运行

```bash
go build -o excelparser
./excelparser --path=./xlsx --server=lua --client=json --indent
```

### 常用命令

```bash
# 强制重新导出所有文件
--force

# JSON 格式美化
--indent

# 紧凑模式(减少文件大小)
--compact

# 国际化导出
--i18n=./locales --lang=en_US
```

### 测试场景

-   使用 `xlsx/tpl/` 中的模板测试新特性
-   修改后检查 `export.yaml` 时间戳变化
-   验证错误处理: 故意在 Excel 中制造类型错误/ID 重复

## 关键实现细节

### 字段解析递归 ([xlsx.go:182](xlsx.go#L182))

`parseField` 使用合并单元格 (`getMergeRangeX`) 推断嵌套结构:

-   合并单元格跨度决定数组容量
-   Map 的 key/value 按列对出现
-   Struct 子字段名用点号分隔 (例如 `s1.a`, `s1.b`)

### 数据缓存 ([xlsx.go:71-110](xlsx.go#L71-L110))

`Datas []string` 积累输出片段而非直接写文件:

-   `appendData` / `replaceComma` / `replaceTail` 操作缓存
-   最后一次性 `strings.Join` 写入 ([xlsx.go:540](xlsx.go#L540))

### JSON 类型验证 ([type.go](type.go))

`checkJsonVal` 解析 JSON 字符串并递归验证 `Vtype` 结构

### 国际化流程

1. 读取 `.po` 文件 (gotext 库)
2. 遍历字段标记 `I18n=true` 的字段
3. 替换时记录引用位置 (`AddRefs`) 用于 .po 文件更新

## 常见陷阱

-   **不要修改 `HeadLineNum=4`**: 整个解析依赖此常量
-   **字段索引基于合并后的列**: 复杂类型占多列,索引需递增
-   **纵向表不支持 compact**: `ternary(FlagCompact && !x.Vertical, ...)` 检查
-   **错误累积上限**: `MaxErrorCnt=6` 防止日志爆炸
-   **goroutine 池必须等待**: `wg.Wait()` + `close(FinishChan)` 顺序不能颠倒

## 扩展点

### 新增输出格式

1. 在 [formater.go](formater.go) 实现 `iFormater` 接口
2. 在 `NewFormater` switch 添加 case
3. 在 `writeToFile` 添加文件扩展名映射

### 新增类型

1. 在 [type.go](type.go) 添加 `const T*`
2. 在 `parseType` 添加解析逻辑
3. 在 formatters 的 `formatData` 添加处理分支
