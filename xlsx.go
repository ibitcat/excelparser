package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

var HeadLineNum = 4 // 配置表头行数
const (
	NameLine int = iota + 1
	TypeLine
	ModeLine
	DescLine
)

type ExportInfo struct {
	Mode     string `json:"mode"`
	Format   string `json:"format"`
	LastTime uint64 `json:"lasttime"`
}

type Xlsx struct {
	Name         string         // 文件名（带文件扩展名）
	FileName     string         // 文件名
	PathName     string         // 文件完整路径
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
	Exports      []*ExportInfo  // 导出信息
	LastModified uint64         // 最后修改时间
	TimeCost     int            // 耗时
}

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
	str := ternary(FlagCompact && !x.Vertical, "", "\n")
	x.appendData(str)
}

func (x *Xlsx) appendSpace() {
	str := ternary(FlagCompact && !x.Vertical, "", " ")
	x.appendData(str)
}

func (x *Xlsx) appendIndent(depth int) {
	str := ternary(FlagCompact && !x.Vertical, "", getIndent(depth))
	x.appendData(str)
}

func (x *Xlsx) appendComma() {
	str := ternary(FlagCompact && !x.Vertical, ",", ",\n")
	x.appendData(str)
}

func (x *Xlsx) replaceComma() {
	str := ternary(FlagCompact && !x.Vertical, "", "\n")
	x.Datas[len(x.Datas)-1] = str
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
	if parent != nil && parent.Kind == TStruct && len(field.Name) == 0 {
		x.sprintfCellError(NameLine, field.Index+1, "字段名称错误(字段名为空)")
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

func (x *Xlsx) canParse() bool {
	var cnt, num int
	for mode, format := range Mode2Format {
		if len(format) > 0 {
			cnt++
			if !FlagForce {
				for _, v := range x.Exports {
					if v.Mode == mode && v.Format == format && v.LastTime == x.LastModified {
						num++
						break
					}
				}
			}
		}
	}
	return cnt != num
}

func (x *Xlsx) updateExportInfo(mode, format string) {
	var e *ExportInfo
	for _, v := range x.Exports {
		if v.Mode == mode && v.Format == format {
			e = v
			break
		}
	}
	if e == nil {
		x.Exports = append(x.Exports, &ExportInfo{mode, format, x.LastModified})
	} else {
		e.LastTime = x.LastModified
	}
}

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

			num := 0
			keyField := x.RootField.Vals[0]
			for mode, format := range Mode2Format {
				if len(format) > 0 && keyField.isHitMode(mode) {
					num++
					formater := NewFormater(x, format, mode)
					formater.formatRows()

					// write
					if len(x.Errors) == 0 {
						x.updateExportInfo(mode, format)
						x.writeToFile(mode, format)
					}
				}
			}
			if num == 0 {
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
	outdir := FlagOutput + sep + mode + sep + format + sep
	outFileName := fmt.Sprintf("%s%s.%s", outdir, x.FileName, ext)
	err := os.MkdirAll(filepath.Dir(outFileName), os.ModeDir)
	if err == nil || os.IsExist(err) {
		outFile, operr := os.OpenFile(outFileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o666)
		if operr != nil {
			return
		}
		defer outFile.Close()

		outFile.WriteString(strings.Join(x.Datas, ""))
		outFile.Sync()
	}
}

func (x *Xlsx) collectResult(costFormat, infoFormat, splitline string) []string {
	results := make([]string, 0)
	results = append(results, splitline)

	errNum := len(x.Errors)
	if errNum == 0 {
		results = append(results, fmt.Sprintf(costFormat, x.FileName, x.TimeCost, len(x.Rows)))
	} else if errNum == 1 {
		results = append(results, fmt.Sprintf(infoFormat, x.FileName, x.Errors[0]))
	} else {
		mid := int(math.Ceil(float64(errNum)/2)) - 1
		for i := 0; i < errNum; i++ {
			err := x.Errors[i]
			if mid == i {
				results = append(results, fmt.Sprintf(infoFormat, x.FileName, err))
			} else {
				results = append(results, fmt.Sprintf(infoFormat, "", err))
			}
		}
	}
	return results
}
