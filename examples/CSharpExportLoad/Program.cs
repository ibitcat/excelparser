using MessagePack;

namespace ExcelParser.Examples.CSharpExportLoad;

/// <summary>
/// 演示：用 MessagePack-CSharp 加载 excelparser 导出的 .bytes（与 out/server/csharp 中生成类对应）。
/// 运行前请先在工具里导出 C# 与 bytes，保证仓库内 out/server/csharp 已更新。
/// </summary>
internal static class Program
{
    private static string ExportDir =>
        Path.Combine(AppContext.BaseDirectory, "csharp-export");

    private static int Main(string[] args)
    {
        var dir = args.Length > 0 ? args[0] : ExportDir;
        if (!Directory.Exists(dir))
        {
            Console.Error.WriteLine($"未找到导出目录: {dir}");
            Console.Error.WriteLine("若未传参，默认使用输出目录下的 csharp-export（dotnet run 会先复制 *.bytes）。");
            return 1;
        }

        var failed = false;

        // --- 纵向表：基础类型 ---
        failed |= !Run("system.bytes (纵向表 + 基础类型)", () =>
        {
            var data = ReadBytes(dir, "system");
            var row = MessagePackSerializer.Deserialize<TSystem>(data);
            Console.WriteLine($"  key1(int)={row.key1}, key2(string)={row.key2}, key3(bool)={row.key3}");
            Console.WriteLine($"  key4(uint)={row.key4}, key5(float)={row.key5}");
            Console.WriteLine($"  key6(int[])=[{string.Join(",", row.key6 ?? [])}]");
            Console.WriteLine($"  key7(int[][]).Length={row.key7?.Length ?? 0}");
        });

        // --- 横向表：简单结构 ---
        failed |= !Run("item.bytes (横向表 + 简单字段)", () =>
        {
            var data = ReadBytes(dir, "item");
            var map = MessagePackSerializer.Deserialize<Dictionary<int, TItem>>(data);
            Console.WriteLine($"  行数={map.Count}, [1001].name={map[1001].name}");
        });

        failed |= !Run("error.bytes (横向表 + 基础验证)", () =>
        {
            var data = ReadBytes(dir, "error");
            var map = MessagePackSerializer.Deserialize<Dictionary<int, TError>>(data);
            Console.WriteLine($"  行数={map.Count}, [1001].msg={map[1001].msg}");
        });

        // --- 复杂结构：嵌套struct + JSON类型定义 + Dictionary ---
        failed |= !Run("template.bytes (嵌套结构体)", () =>
        {
            var data = ReadBytes(dir, "template");
            var map = MessagePackSerializer.Deserialize<Dictionary<int, TTemplate>>(data);
            var row = map[1001];

            // 嵌套结构体
            Console.WriteLine($"  s1.a={row.s1.a}, s1.b={row.s1.b}, s1.g={row.s1.g}");
            Console.WriteLine($"  s1.c(定长数组)=[{string.Join(",", row.s1.c ?? [])}]");
            Console.WriteLine($"  s1.d(二维数组)=[{string.Join("; ", (row.s1.d ?? []).Select(a => $"[{string.Join(",", a)}]"))}]");

            // 自引用结构体
            Console.WriteLine($"  s1.x(自引用).a={row.s1.x?.a}");

            // 子结构体数组
            Console.WriteLine($"  s1.f(SubType[]).Length={row.s1.f?.Length}, [0]={{a={row.s1.f?[0].a}, b={row.s1.f?[0].b}}}");

            // 子结构体字典
            var hEntry = row.s1.h?.GetValueOrDefault(1);
            Console.WriteLine($"  s1.h(Dict<int,SubType>)[1]={{a={hEntry?.a}, b={hEntry?.b}}}");
        });

        failed |= !Run("template.bytes (JSON 带类型定义)", () =>
        {
            var data = ReadBytes(dir, "template");
            var map = MessagePackSerializer.Deserialize<Dictionary<int, TTemplate>>(data);
            var row = map[1001];

            // json:{age=int,sites=[]{name=string,url=string}} → TTemplateJsonval
            Console.WriteLine($"  jsonval.age(int)={row.jsonval?.age}");
            var site = row.jsonval?.sites?.FirstOrDefault();
            Console.WriteLine($"  jsonval.sites[0]={{name={site?.name}, url={site?.url}}}");

            // json:[]i18n → string[]
            Console.WriteLine($"  i18njson(string[])=[{string.Join(",", row.i18njson ?? [])}]");
        });

        failed |= !Run("template.bytes (Dictionary 类型)", () =>
        {
            var data = ReadBytes(dir, "template");
            var map = MessagePackSerializer.Deserialize<Dictionary<int, TTemplate>>(data);
            var row = map[1001];

            // map[int]string
            Console.WriteLine($"  map1[1]={row.map1?[1]}, map1[2]={row.map1?[2]}");

            // map[int]map[int]string
            Console.WriteLine($"  map2[1][111]={row.map2?[1]?[111]}");

            // map[int][]int
            var arr = row.map3?.GetValueOrDefault(101);
            Console.WriteLine($"  map3[101]=[{string.Join(",", arr ?? [])}]");
        });

        // --- JSON 二维数组 ---
        failed |= !Run("tpl2.bytes (JSON 二维int数组)", () =>
        {
            var data = ReadBytes(dir, "tpl2");
            var map = MessagePackSerializer.Deserialize<Dictionary<int, TTpl2>>(data);
            var row = map[1001];
            Console.WriteLine($"  id={row.id}, type={row.type}");
            Console.WriteLine($"  conditions(int[][]).Length={row.conditions?.Length}");
            if (row.conditions is { Length: > 0 })
                Console.WriteLine($"  conditions[0]=[{string.Join(",", row.conditions[0])}]");
        });

        if (failed)
        {
            Console.Error.WriteLine("\n部分文件加载失败，请查看上方异常。");
            return 1;
        }

        Console.WriteLine("\n全部 .bytes 已按生成类型反序列化成功。");
        return 0;
    }

    private static byte[] ReadBytes(string dir, string baseName)
    {
        var path = Path.Combine(dir, baseName + ".bytes");
        if (!File.Exists(path))
            throw new FileNotFoundException("缺少二进制文件（请先在 excelparser 中导出）", path);
        return File.ReadAllBytes(path);
    }

    /// <returns>是否成功</returns>
    private static bool Run(string label, Action action)
    {
        Console.WriteLine(label);
        try
        {
            action();
            return true;
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine($"  失败: {ex.Message}");
            return false;
        }
    }
}
