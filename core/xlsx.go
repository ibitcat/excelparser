package core

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// methods
func (x *Xlsx) appendError(errMsg string) {
	errCnt := len(x.Errors)
	if errCnt < MaxErrorCnt {
		if errCnt == MaxErrorCnt-1 {
			errMsg = "..."
		}
		x.Errors = append(x.Errors, errMsg)
	}
}

func (x *Xlsx) sprintfError(format string, a ...any) {
	x.appendError(fmt.Sprintf(format, a...))
}

func (x *Xlsx) sprintfCellError(row, col int, format string, a ...any) {
	if x.Vertical {
		x.sprintfError("[%s%d]%s", formatAxisX(row), col, fmt.Sprintf(format, a...))
	} else {
		x.sprintfError("[%s%d]%s", formatAxisX(col), row, fmt.Sprintf(format, a...))
	}
}

func (x *Xlsx) appendData(str string) {
	if len(str) > 0 {
		x.Datas = append(x.Datas, str)
	}
}

func (x *Xlsx) appendEOL() {
	str := ternary(GFlags.Compact && !x.Vertical, "", "\n")
	x.appendData(str)
}

func (x *Xlsx) appendSpace() {
	str := ternary(GFlags.Compact && !x.Vertical, "", " ")
	x.appendData(str)
}

func (x *Xlsx) appendIndent(depth int) {
	str := ternary(GFlags.Compact && !x.Vertical, "", getIndent(depth))
	x.appendData(str)
}

func (x *Xlsx) appendComma() {
	str := ternary(GFlags.Compact && !x.Vertical, ",", ",\n")
	x.appendData(str)
}

func (x *Xlsx) replaceComma() {
	tailIdx := len(x.Datas) - 1
	comma := x.Datas[tailIdx]
	if len(comma) > 0 && comma[:1] == "," {
		str := ternary(GFlags.Compact && !x.Vertical, "", "\n")
		x.Datas[tailIdx] = str
	}
}

func (x *Xlsx) replaceTail(str string) {
	x.Datas[len(x.Datas)-1] = str
}

func (x *Xlsx) clearData() {
	x.Datas = x.Datas[0:0]
}

func (x *Xlsx) createField(i int) *Field {
	f := new(Field)
	f.Index = i
	f.Xlsx = x
	if i < len(x.Names) {
		f.Rname = strings.TrimSpace(x.Names[i])
		f.Name = f.Rname
	}
	if i < len(x.Types) {
		typ := strings.TrimSpace(x.Types[i])
		f.Type = parseType(typ)
	}
	if i < len(x.Modes) {
		f.Mode = strings.TrimSpace(x.Modes[i])
	}
	if i < len(x.Descs) {
		f.Desc = strings.TrimSpace(x.Descs[i])
	}
	if len(f.Rname) > 0 {
		s := strings.Split(f.Rname, ".")
		if len(s) > 0 {
			f.Name = s[len(s)-1]
		}
	}
	return f
}

func (x *Xlsx) readSheetHead() [][]string {
	var cur int = 0
	var results [][]string
	if x.Vertical {
		// 纵向表
		cols, err := x.Excel.Cols(x.SheetName)
		if err != nil {
			return nil
		}
		for cols.Next() {
			cur++
			col, err := cols.Rows()
			if err != nil {
				break
			}
			results = append(results, col)
			if cur == HeadLineNum {
				break
			}
		}
	} else {
		// 横向表
		rows, err := x.Excel.Rows(x.SheetName)
		if err != nil {
			return nil
		}
		for rows.Next() {
			cur++
			row, err := rows.Columns()
			if err != nil {
				break
			}
			results = append(results, row)
			if cur == HeadLineNum {
				break
			}
		}
	}
	return results
}

func (x *Xlsx) parseField(parent *Field, sindex int) int {
	i := sindex
	pkind := parent.Kind
	for i < len(x.Types) {
		if pkind == TNone {
			break
		} else if pkind == TArray {
			// list
			v := x.createField(i)
			if len(parent.Vals) == parent.Cap {
				break
			}
			pv := parent.Vtype
			if !(pv.isAny() || v.Kind == pv.Kind) {
				break
			}

			i++
			v.Parent = parent
			parent.Vals = append(parent.Vals, v)
			if v.isRecursice() {
				i += x.parseField(v, i)
			}
		} else if pkind == TMap {
			// kv 是否匹配
			k := x.createField(i)
			v := x.createField(i + 1)
			if len(k.Name) > 0 {
				break
			}
			pk := parent.Ktype
			pv := parent.Vtype
			if !pk.isBuiltin() {
				break
			}
			if pk.Kind != k.Kind {
				break
			}
			if !(pv.isAny() || v.Kind == pv.Kind) {
				break
			}
			i += 2
			k.Parent = parent
			v.Parent = parent
			parent.Keys = append(parent.Keys, k)
			parent.Vals = append(parent.Vals, v)

			if v.isRecursice() {
				i += x.parseField(v, i)
			}
		} else if pkind == TStruct {
			f := x.createField(i)
			if parent.Index >= 0 {
				ckey, found := strings.CutPrefix(f.Rname, parent.Rname)
				if !found {
					break
				}
				if len(strings.Split(ckey, ".")) != 2 {
					break
				}
			}

			i++
			f.Parent = parent
			if parent.Ftypes == nil {
				parent.Ftypes = make(map[string]*Type)
			}
			parent.Ftypes[f.Name] = f.Type
			parent.Vals = append(parent.Vals, f)
			if f.isRecursice() {
				i += x.parseField(f, i)
			}
		} else if pkind == TJson {
			break
		} else {
			f := x.createField(i)
			i++
			f.Parent = parent
			parent.Vals = append(parent.Vals, f)

			if f.isRecursice() {
				i += x.parseField(f, i)
			}
		}
	}
	return i - sindex
}

func (x *Xlsx) getMergeRangeX() [][]int {
	mergeCells, _ := x.Excel.GetMergeCells(x.SheetName)
	rangeX := make([][]int, 0, len(mergeCells))
	for _, mergeCell := range mergeCells {
		startx, starty := splitAxis(mergeCell.GetStartAxis())
		endx, endy := splitAxis(mergeCell.GetEndAxis())
		if starty == 1 && endy == 1 {
			rangeX = append(rangeX, []int{startx, endx})
		}
		if x.Vertical {
			if max(2, startx) < min(5, endx) {
				x.appendError("第2~5行不能有合并单元格")
				return nil
			}
		} else {
			if max(2, starty) < min(5, endy) {
				x.appendError("第2~5行不能有合并单元格")
				return nil
			}
		}
	}
	return rangeX
}

func (x *Xlsx) parseHeader() {
	f := new(Field)
	f.Index = -1
	f.Xlsx = x
	f.Type = &Type{Kind: TStruct}
	x.RootField = f

	fieldNum := len(x.Types)
	for i := 0; i < fieldNum; {
		i += x.parseField(x.RootField, i)
	}
}

func (x *Xlsx) checkField(field *Field) {
	if !field.isVaild(false) {
		x.sprintfCellError(TypeLine, field.Index+1, "字段类型错误(类型不合法)")
	}
	if !field.isVaildMode() {
		x.sprintfCellError(ModeLine, field.Index+1, "导出模式错误")
	}
	if field.Kind == TMap && len(field.Keys) != len(field.Vals) {
		x.sprintfCellError(TypeLine, field.Index+1, "字段类型错误(map键值对不匹配)")
	}

	parent := field.Parent
	if parent != nil {
		if field.Kind == TStruct && (parent.Kind == TArray || parent.Kind == TMap) {
			if (parent.Name + "[]") != field.Name {
				x.sprintfCellError(NameLine, field.Index+1, "字段名称错误(字段名%s应为%s[])", field.Name, parent.Name)
			}
		}

		if parent.Kind == TStruct && len(field.Name) == 0 {
			x.sprintfCellError(NameLine, field.Index+1, "字段名称错误(字段名为空)")
		}
	}

	if len(field.Keys) > 0 {
		for _, k := range field.Keys {
			x.checkField(k)
		}
	}
	if len(field.Vals) > 0 {
		keyMap := map[string]int{}
		for _, v := range field.Vals {
			if field.Kind == TStruct {
				_, ok := keyMap[v.Name]
				if ok {
					x.sprintfCellError(NameLine, v.Index+1, "字段名称错误(字段名%s冲突)", v.Name)
				} else {
					keyMap[v.Name] = v.Index
				}
			}
			x.checkField(v)
		}
	}
}

func (x *Xlsx) checkFields() {
	x.checkField(x.RootField)

	// key field
	keyField := x.RootField.Vals[0]
	if len(keyField.Mode) != 0 {
		x.appendError("key 字段的导出模式错误")
	}
	if !x.Vertical {
		// 横向表
		if keyField.Name != "id" {
			x.appendError("Key 字段必须以 id 命名")
		}
		if !keyField.isInteger() {
			x.appendError("横向表 Key 字段类型必须为整数")
		}
	}
}

func (x *Xlsx) checkRows() {
	line := 0
	x.Rows = make([][]string, 0, 64)
	if x.Vertical {
		cols, _ := x.Excel.Cols(x.SheetName)
		for cols.Next() {
			line++
			if line > HeadLineNum {
				col, err := cols.Rows()
				if err != nil {
					break
				}

				if x.RootField.checkRow(col, line, x) {
					x.Rows = append(x.Rows, col)
				}
				break
			}
		}
	} else {
		idMap := make(map[string]int)
		rows, _ := x.Excel.Rows(x.SheetName)
		for rows.Next() {
			line++
			if line > HeadLineNum {
				row, err := rows.Columns()
				if err != nil {
					break
				}
				if len(row) == 0 {
					break
				}

				key := row[0]
				if strings.HasPrefix(key, "//") || key == "" {
					continue
				}

				num, ok := idMap[key]
				if ok {
					x.sprintfCellError(line, 1, "Id [%s] 重复 %d 次", key, num-1)
				} else {
					idMap[key] += 1
				}

				if x.RootField.checkRow(row, line, x) {
					x.Rows = append(x.Rows, row)
				}
			}
		}
	}
}

// 是否文件已修改
func (x *Xlsx) isModified(mode, format string) bool {
	if GFlags.Force {
		return true
	}

	for _, v := range x.Exports {
		if v.Mode == mode && v.Format == format && v.LastTime == x.LastModified {
			// 文件未修改
			return false
		}
	}
	return true
}

func (x *Xlsx) CanParse() bool {
	if len(GFlags.Server) > 0 && x.isModified("server", GFlags.Server) {
		return true
	}
	if len(GFlags.Client) > 0 && x.isModified("client", GFlags.Client) {
		return true
	}
	return false
}

func (x *Xlsx) updateExportInfo(mode, format string) {
	var e *ExportInfo
	for i := 0; i < len(x.Exports); i++ {
		p := &x.Exports[i]
		if p.Mode == mode && p.Format == format {
			e = p
			break
		}
	}
	if e != nil {
		e.LastTime = x.LastModified
	} else {
		x.Exports = append(x.Exports, ExportInfo{mode, format, x.LastModified})
	}
}

// 解析excel表头并静态检查表数据
func (x *Xlsx) parseExcel() bool {
	var vertical bool
	f := x.Excel
	sheetIdx, _ := f.GetSheetIndex("data")
	if sheetIdx == -1 {
		vertical = true
		sheetIdx, _ = f.GetSheetIndex("vdata")
	}
	if sheetIdx == -1 {
		x.appendError("data/vdata sheet 不存在")
		return false
	}

	x.Vertical = vertical
	x.SheetName = f.GetSheetName(sheetIdx)
	heads := x.readSheetHead()
	if len(heads) < HeadLineNum {
		x.appendError("配置表头格式错误(不足4行)")
		return false
	}

	x.Names = heads[NameLine-1] // 字段名行
	x.Types = heads[TypeLine-1] // 字段类型行
	x.Modes = heads[ModeLine-1] // 导出模式行
	x.Descs = heads[DescLine-1] // 字段描述行
	x.parseHeader()
	x.checkFields()
	x.checkRows()
	return true
}

func (x *Xlsx) exportModeExcel(mode, format string) bool {
	if len(mode) == 0 || len(format) == 0 {
		return false
	}
	keyField := x.RootField.Vals[0]
	if !keyField.isHitMode(mode) {
		// key字段mode不匹配
		return false
	}
	formater := NewFormater(x, format, mode)
	formater.formatRows()

	// write
	if len(x.Errors) == 0 {
		x.updateExportInfo(mode, format)
		x.writeToFile(mode, format)
	}
	return true
}

func (x *Xlsx) exportExcel() {
	f, err := excelize.OpenFile(x.PathName)
	if err == nil {
		defer func() {
			x.Excel = nil
			f.Close()
		}()

		x.Excel = f
		ok := x.parseExcel()
		if ok && len(x.Errors) == 0 && len(x.Rows) > 0 {
			x.Datas = make([]string, 0)

			sok := x.exportModeExcel("server", GFlags.Server)
			cok := x.exportModeExcel("client", GFlags.Client)
			if !sok && !cok {
				x.appendError("无需生成")
			}
		}
	} else {
		x.appendError("xlsx文件打开失败")
	}
}

func (x *Xlsx) writeToFile(mode, format string) {
	ext := ""
	switch format {
	case "lua":
		ext = "lua"
	case "json":
		ext = "json"
	}
	sep := string(filepath.Separator)
	// linux: out/server/json
	// windows: out\server\json
	outdir := GFlags.Output + sep + mode + sep + format + sep
	outFileName := fmt.Sprintf("%s%s.%s", outdir, x.OutName, ext)
	err := os.MkdirAll(filepath.Dir(outFileName), 0o755)
	if err == nil || os.IsExist(err) {
		outFile, operr := os.OpenFile(outFileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o666)
		if operr != nil {
			return
		}
		defer outFile.Close()

		outFile.WriteString(strings.Join(x.Datas, ""))
		outFile.Sync()
	} else {
		x.appendError(fmt.Sprintf("输出目录[%s]创建失败: %v", outdir, err))
	}
}

func (x *Xlsx) collectResult(costFormat, infoFormat, splitline string) []string {
	results := make([]string, 0)
	results = append(results, splitline)

	errNum := len(x.Errors)
	switch errNum {
	case 0:
		results = append(results, fmt.Sprintf(costFormat, x.OutName, x.TimeCost, len(x.Rows)))
	case 1:
		results = append(results, fmt.Sprintf(infoFormat, x.OutName, x.Errors[0]))
	default:
		mid := int(math.Ceil(float64(errNum)/2)) - 1
		for i := range errNum {
			err := x.Errors[i]
			if mid == i {
				results = append(results, fmt.Sprintf(infoFormat, x.OutName, err))
			} else {
				results = append(results, fmt.Sprintf(infoFormat, "", err))
			}
		}
	}
	return results
}

func (x *Xlsx) formatLuaComment(mode string) string {
	clsName := "T" + toTitle(x.OutName)
	comments := make([]string, 0, len(x.RootField.Vals)+3)
	comments = append(comments, "---"+x.FileName)
	comments = append(comments, "---@class "+clsName)
	for _, v := range x.RootField.Vals {
		if v.isHitMode(mode) {
			comments = append(comments, "---@field "+v.Name+" "+v.luaTypeName()+" "+v.Desc)
		}
	}
	if x.Vertical {
		comments = append(comments, "\n---@type "+clsName)
	} else {
		comments = append(comments, "\n---@type table<integer, "+clsName+">")
	}
	return strings.Join(comments, "\n")
}
