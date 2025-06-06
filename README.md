# excelparser

golang 实现的 Excel 解析器。

## 特性

-   [x] 多协程加速生成
-   [x] 支持 lua 配置生成
-   [x] 支持 json 配置生成
-   [x] id 重复检查
-   [x] 字段名重复检查
-   [x] 行注释
-   [x] 列注释
-   [x] json 内容合法性检查
-   [x] json 输出格式化(json 格式缩进美化)
-   [x] 支持生成标签(s=server, c=client, x=不生成, 空=server 和 client 都生成)
-   [x] 字段数据类型检查(支持 `int`，`uint`,`float`， `bool`, `string`，`json`，`array`，`map`，`struct`)
-   [x] 配置错误详情输出
-   [x] 未修改的文件忽略生成(可以加速生成速度，不需要每次都全部生成一次)
-   [x] 支持纵向表
-   [x] 基础数据类型字段使用默认值填充字段
-   [x] 配置生成压缩
-   [x] 支持国际化翻译
-   [ ] 数值类型范围检查
-   [ ] id 公式检查

## 参数

-   path，xlsx 配置文件目录
-   output，生成文件的输出目录，默认为 `.`
-   server，指定 server 端生成格式，例如：--server=json
-   client, 指定 client 端生成约束，例如：--client=lua
-   indent, 生成含有 json 类型的配置时，是否格式化(美化) json（默认关闭）
-   force, 强制重新导出所有配置（默认关闭）
-   default, 若配置的基础数据类型字段（ `int`，`bool`, `string`，`float`）未配置时，使用默认值填充字段(例如：`0`, `false`, `0.0`, `0.0`)（默认开启）
-   compact, 生成的配置成行压缩，减少文件大小（默认关闭）
-   i18n，国际化翻译配置路径
-   lang，国际化翻译目标语言(en=英文;jp=日文;kr=韩文等)

**ps**：真正的输出路径格式为: `output/[server|client]/文件格式`，例如：./server/json 表示服务端 json 格式的输出目录。

## 使用

解析器只识别名为 `data` 或者 `vdata` 的工作表。

-   横向表：Excel Sheet 命名为 `data`，一般常用的配置方式，支持多行数据配置。
-   纵向表：Excel Sheet 命名为 `vdata`，一般用来配置全局字段表，只支持一行数据配置。

```
执行：
excelparser.exe --path=./xlsx --server=lua --client=json --indent --force
Progress:[██████████████████████████████████████████████████][100%]
------------------------------+----------------------------------------------------------------------
FileName                      | Result
------------------------------+----------------------------------------------------------------------
system                        | 42   ms
------------------------------+----------------------------------------------------------------------
task                          | 43   ms
------------------------------+----------------------------------------------------------------------
```

-   示例 1

```
server 生成lua配置并导出到 ./server/lua 目录中；client 生成lua配置并导出到 ./client/lua 目录中。
excelparser --path=./xlsx --server=lua --client=lua
```

-   示例 2:

```
server 生成lua配置并导出到 ./server/lua 目录中；client 生成json配置并导出到 ./cient/json 目录中，并格式化 json。
excelparser --path=./xlsx --server=lua --client=json --indent
```

-   示例 3:

```
server 生成json配置到 ./out/server/json 目录中，并使用 ./i18n 目录中的 en.xlsx 翻译文件来替换配置中的 i18n 类型配置值。
excelparser.exe --force=true --path=./xlsx --output=./out --server=json --i18n=./locales --lang=en_US
```

## 表头格式

### json

使用 json 类型时，可以指定在`<>`内指定 json 真正导出的数据结构，支持定长数组、变长数组、map(支持嵌套)，但不支持 any 和 struct(_不好描述结构体原型_)。另外，需要注意的是，变长数组只适用于 json 中，表头描述的类型只支持定长数组。

| id          | jsonval     |
| ----------- | ----------- |
| int         | json<[]int> |
|             | s           |
| 配置唯一 id | json 字符串 |
| 1001        | [1,2,3]     |

### 简单定长数组

| id          | list1    |              |         |          |
| ----------- | -------- | ------------ | ------- | -------- |
| int         | [3]any   | i18n         | int     | int      |
|             | s        |              |         |          |
| 配置唯一 id | 奖励道具 | 道具名       | 道具 id | 道具数量 |
| 1001        |          | 这是一个道具 | 2       | 3        |

### 二维数组

| id          | list2     |        |     |     |     |        |     |     |     |        |     |     |     |
| ----------- | --------- | ------ | --- | --- | --- | ------ | --- | --- | --- | ------ | --- | --- | --- |
| int         | [3][3]int | [3]int | int | int | int | [3]int | int | int | int | [3]int | int | int | int |
|             | s         |        |     |     |     |        |     |     |     |        |     |     |     |
| 配置唯一 id | 奖励列表  |        |     |     |     |        |     |     |     |        |     |     |     |
| 1001        |           |        | 1   | 2   | 3   |        | 11  | 22  | 33  |        | 11  | 22  | 33  |

### 简单 map

| id          | map1           |     |        |     |        |
| ----------- | -------------- | --- | ------ | --- | ------ |
| int         | map[int]string | int | string | int | string |
|             | s              |     |        |     |        |
| 配置唯一 id | 简单 map       |     |        |     |        |
| 1001        |                | 1   | aaa    | 2   | bbb    |

### 嵌套 map

| id          | map2                   |     |                |     |        |     |                |     |        |
| ----------- | ---------------------- | --- | -------------- | --- | ------ | --- | -------------- | --- | ------ |
| int         | map[int]map[int]string | int | map[int]string | int | string | int | map[int]string | int | string |
|             | s                      |     |                |     |        |     |                |     |        |
| 配置唯一 id | 嵌套 map               |     | 子 map1        |     |        |     | 子 map2        |     |        |
| 1001        |                        | 1   |                | 111 | bbb    | 2   |                | 111 | bbb    |

### 数组 map

| id          | map3           |     |        |     |     |     |     |        |     |     |     |
| ----------- | -------------- | --- | ------ | --- | --- | --- | --- | ------ | --- | --- | --- |
| int         | map[int][3]int | int | [3]int | int | int | int | int | [3]int | int | int | int |
|             | s              |     |        |     |     |     |     |        |     |     |     |
| 配置唯一 id | 数组 map       |     | 数组 1 |     |     |     |     | 数组 2 |     |     |     |
| 1001        |                | 101 |        | 11  | 22  | 33  | 102 |        | 11  | 22  | 33  |

### 结构体

| id          | s1               | s1.a   | s1.b   | s1.c   |     |     |     | s1.d      |        |     |     |     |        |     |     |     | s1.e   |
| ----------- | ---------------- | ------ | ------ | ------ | --- | --- | --- | --------- | ------ | --- | --- | --- | ------ | --- | --- | --- | ------ |
| int         | struct<TaskType> | int    | string | [3]int | int | int | int | [2][3]int | [3]int | int | int | int | [3]int | int | int | int | int    |
|             | s                |        |        |        |     |     |     |           |        |     |     |     |        |     |     |     |        |
| 配置唯一 id | 结构体           | 字段 a | 字段 b | 字段 c |     |     |     | 字段 d    |        |     |     |     |        |     |     |     | 字段 e |
| 1001        |                  | 111    | 2222   |        | 1   | 2   | 3   |           |        | 122 | 222 | 333 |        | 122 | 222 | 333 | 1001   |
