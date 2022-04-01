# excelparser
golang 实现的 Excel 解析器。

## 特性
- [x] 多协程加速生成
- [x] 支持 lua 配置生成
- [x] 支持 json 配置生成
- [x] id 重复检查
- [x] 字段名重复检查
- [x] 行注释
- [x] 列注释
- [x] json 内容合法性检查
- [x] json 输出格式化(json格式缩进美化)
- [x] 支持生成标签(s=server, c=client, b=both)
- [x] 字段数据类型检查(支持 `int`，`bool`, `string`，`float`，`json`，`dict`，`一维数组`，`多维数组`)
- [x] 配置错误详情输出
- [x] 未修改的文件忽略生成(可以加速生成速度，不需要每次都全部生成一次)
- [x] 支持纵向表
- [x] 基础数据类型字段使用默认值填充字段
- [x] 配置生成压缩
- [ ] 数值类型范围检查
- [ ] id 公式检查

## 参数
- path，xlsx 配置文件目录
- server，指定 server 端生成约束，格式："配置格式:导出路径"，例如：--server=json:./sjson
- client, 指定 client 端生成约束，格式同上
- indent, 生成含有 json 类型的配置时，是否格式化(美化) json（默认关闭）
- force, 强制重新导出所有配置（默认关闭）
- default, 若配置的基础数据类型字段（ `int`，`bool`, `string`，`float`）未配置时，使用默认值填充字段(例如：`0`, `false`, `0.0`, `0.0`)（默认开启）
- compress, 生成的配置成行压缩，减少文件大小（默认关闭）

## 使用
解析器只识别名为 `data` 或者 `vdata` 的工作表。

- 横向表：Excel Sheet 命名为 `data`，一般常用的配置方式，支持多行数据配置。
- 纵向表：Excel Sheet 命名为 `vdata`，一般用来配置全局字段表，只支持一行数据配置。

```
执行：
excelparser.exe --path=./xlsx --server=lua:./slua --client=json:./cjson --indent --force
Progress:[██████████████████████████████████████████████████][100%]
------------------------------+----------------------------------------------------------------------
FileName                      | Result
------------------------------+----------------------------------------------------------------------
system                        | 42   ms
------------------------------+----------------------------------------------------------------------
task                          | 43   ms
------------------------------+----------------------------------------------------------------------
```

- 示例1
```
server 生成lua配置并导出到 ./slua 目录中；client 生成lua配置并导出到 ./clua 目录中
excelparser --path=./xlsx --server=lua:./slua --client=lua:./clua
```

- 示例2:
```
server 生成lua配置并导出到 ./slua 目录中；client 生成json配置并导出到 ./cjson 目录中，并格式化 json
excelparser --path=./xlsx --server=lua:./slua --client=json:./cjson --indent
```
