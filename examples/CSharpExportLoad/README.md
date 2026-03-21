# CSharpExportLoad

控制台示例（非单元测试）：验证 `out/server/csharp` 里生成的 C# 类型能否用 [MessagePack-CSharp](https://github.com/MessagePack-CSharp/MessagePack-CSharp) 正确反序列化同目录导出的 `.bytes`。

## 运行

在仓库根目录执行：

```bash
dotnet run --project examples/CSharpExportLoad
```

可选：指定已包含 `*.bytes` 的目录（例如你复制后的路径）：

```bash
dotnet run --project examples/CSharpExportLoad -- "D:\path\to\folder"
```

## 说明

- 构建时会将 `out/server/csharp/*.bytes` 复制到输出目录的 `csharp-export`，与 `tests/CSharpBytesLoad` 行为一致。
- `tpl1.cs` 与 `template.cs` 均定义 `TaskType` / `SubType`，不能在同一项目同时编译；本示例只覆盖 `system`、`item`、`error`、`template`、`tpl2`。若需校验 `tpl1.bytes`，请单独建一个只引用 `tpl1.cs` 的小项目。
