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
- [x] 配置错误详情输出
- [x] json 输出格式化
- [x] 支持生成标签(s=server, c=client, b=both)
- [ ] 数值类型范围检查
- [ ] id 公式检查


## 使用
```
执行：
excelparser.exe --path=./xlsx --server=json:./sjson --client=json:./cjson --indent

输出：
--------------------+--------------------------------------------------
FileName            | Result
--------------------+--------------------------------------------------
task                | 46   ms
--------------------+--------------------------------------------------
```

## 参数
- path，xlsx 配置文件目录
- server，指定 server 端生成约束，格式："配置格式:导出路径"，例如：--server=json:./sjson
- client, 指定 client 端生成约束，格式同上
- indent, 生成含有 json 类型的配置时，是否格式化(美化) json

## 示例

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