package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Xlsx struct {
	Name      string // 文件名（带文件扩展名）
	PathName  string // 文件完整路径
	FileName  string // 文件名
	Vertical  bool   // 纵向表
	Descs     []string
	Names     []string
	Types     []string
	Modes     []string
	RootField *FieldInfo
	Rows      [][]string
	Datas     []string
	Errors    []string // 错误信息
	TimeCost  int      // 耗时
}

/// methods
func (x *Xlsx) openExcel() error {
	f, err := excelize.OpenFile(x.PathName)
	if err != nil {
		x.appendError("xlsx文件打开失败")
		return err
	}
	defer f.Close()

	var vertical bool
	sheetIdx := f.GetSheetIndex("data")
	if sheetIdx == -1 {
		vertical = true
		sheetIdx = f.GetSheetIndex("vdata")
	}
	if sheetIdx == -1 {
		x.appendError("data/vdata sheet 不存在")
		return errors.New("data/vdata not exist")
	}

	sheetName := f.GetSheetName(sheetIdx)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		x.appendError("data/vdata sheet 打开错误")
		return err
	}
	x.Vertical = vertical
	if vertical {
		if len(rows[0]) < 5 {
			x.appendError("纵向表必须要5列")
			return errors.New("vdata error")
		}
		x.Rows = rotateRows(rows)
	} else {
		x.Rows = rows
	}
	return nil
}

func (x *Xlsx) appendError(errMsg string) {
	errCnt := len(x.Errors)
	if errCnt < MaxErrorCnt {
		if errCnt == MaxErrorCnt-1 {
			errMsg = "..."
		}
		x.Errors = append(x.Errors, errMsg)
	}
}

func (x *Xlsx) appendData(str string) {
	x.Datas = append(x.Datas, str)
}

func (x *Xlsx) trimData(str string) {
	x.Datas[len(x.Datas)-1] = str
}

func (x *Xlsx) clearData() {
	x.Datas = x.Datas[0:0]
}

func (x *Xlsx) getKeyField() *FieldInfo {
	return x.RootField.Fields[0]
}

func (x *Xlsx) checkKeyField() {
	keyField := x.RootField.Fields[0]
	if keyField.Index != 0 {
		x.appendError("Key 字段不能注释")
	} else if keyField.Name != "id" && !x.Vertical {
		x.appendError("Key 字段必须以 id 命名")
	}
	if !x.Vertical {
		if !(isNumberType(keyField.Type) || keyField.Type == "string") {
			x.appendError("Key 字段数据类型必须为定点整数或字符串")
		}
	}

	idMap := make(map[string]int)
	for line := 4; line < len(x.Rows); line++ {
		row := x.Rows[line]
		key := row[0]
		if strings.HasPrefix(key, "//") || key == "" {
			continue
		}
		idMap[key] += 1
	}
	for key, num := range idMap {
		if num > 1 {
			x.appendError(fmt.Sprintf("Id [%s] 重复 %d 次", key, num-1))
		}
	}
}

func (x *Xlsx) checkFields(f *FieldInfo) {
	if f.Index >= 0 {
		if !f.Parent.IsArray && len(f.Name) == 0 {
			x.appendError(fmt.Sprintf("字段名为空：(列%d)[%s@%s]", f.Index+1, f.Name, f.Desc))
		}
		if len(f.Fields) != f.FieldNum {
			x.appendError(fmt.Sprintf("字段元素个数不匹配：(列%d)[%s@%s]", f.Index+1, f.Name, f.Desc))
		}
		if !f.IsVaildType() {
			x.appendError(fmt.Sprintf("字段类型错误：(列%d)[%s@%s]", f.Index+1, f.Name, f.Desc))
		}
		if !f.IsVaildMode() {
			x.appendError(fmt.Sprintf("字段标签错误：(列%d)[%s@%s]", f.Index+1, f.Name, f.Desc))
		}
	}

	if len(f.Fields) > 0 {
		tmpMap := map[string]int{}
		for _, field := range f.Fields {
			if f.IsArray {
				if f.Type == "dict" {
					if field.Type != f.Type {
						x.appendError(fmt.Sprintf("结构体数组元素类型错误：(列%d)[%s@%s] ", field.Index+1, field.Name, field.Desc))
					}
				} else {
					if field.RawType != f.Type {
						x.appendError(fmt.Sprintf("数组元素类型错误：(列%d)[%s@%s] ", field.Index+1, field.Name, field.Desc))
					}
				}
			} else {
				index, ok := tmpMap[field.Name]
				if ok {
					x.appendError(fmt.Sprintf("字段名[%s@%s]冲突：(列%d)<->(列%d)", field.Name, field.Desc, index+1, field.Index+1))
				} else {
					tmpMap[field.Name] = field.Index
				}
			}
			x.checkFields(field)
		}
	}
}

func (x *Xlsx) parseHeader() {
	rootField := new(FieldInfo)
	rootField.Index = -1
	rootField.Fields = make([]*FieldInfo, 0, len(x.Names))
	x.RootField = rootField

	for idx := 0; idx < len(x.Types); idx++ {
		idx = x.parseField(rootField, idx)
	}

	// check
	x.checkKeyField()
	x.checkFields(rootField)
}

func (x *Xlsx) parseField(parent *FieldInfo, index int) int {
	if index >= len(x.Types) {
		return index
	}

	def := strings.TrimSpace(x.Types[index])
	field := new(FieldInfo)
	field.Index = index
	field.RawType = def
	field.Parent = parent
	field.Deepth = parent.Deepth + 1
	if len(x.Descs) > index {
		field.Desc = x.Descs[index]
	}
	if len(x.Names) > index {
		field.Name = x.Names[index]
	}

	// mode
	var mode string
	if len(x.Modes) > index {
		mode = x.Modes[index]
	}
	if parent.Index >= 0 && parent.Mode != "b" {
		// 继承父节点 mode
		mode = parent.Mode
	} else if strings.HasPrefix(field.Name, "//") {
		// 列被注释
		mode = ""
	}
	field.Mode = mode
	if len(mode) > 0 {
		parent.Fields = append(parent.Fields, field)
	}

	if arrayBegin := strings.LastIndex(def, "["); arrayBegin != -1 {
		// array
		field.Type = def[:arrayBegin]
		field.IsArray = true

		// sub array
		arrayEnd := strings.LastIndex(def, "]")
		fieldNum, _ := strconv.Atoi(def[(arrayBegin + 1):arrayEnd])
		field.FieldNum = fieldNum
		for i := 0; i < fieldNum; i++ {
			index = x.parseField(field, index+1)
		}
	} else {
		field.Type = def

		isDict := strings.HasPrefix(def, "dict")
		if isDict {
			dictBegin := strings.Index(def, "<")
			dictEnd := strings.Index(def, ">")
			field.Type = def[:dictBegin]
			fieldNum, _ := strconv.Atoi(def[(dictBegin + 1):dictEnd])
			field.FieldNum = fieldNum
			for i := 0; i < fieldNum; i++ {
				index = x.parseField(field, index+1)
			}
		}
	}
	return index
}

func (x *Xlsx) collectResult() []string {
	results := make([]string, 0)
	results = append(results, Splitline)

	errNum := len(x.Errors)
	if errNum == 0 {
		results = append(results, fmt.Sprintf(CostFormat, x.FileName, x.TimeCost))
	} else if errNum == 1 {
		results = append(results, fmt.Sprintf(InfoFormat, x.FileName, x.Errors[0]))
	} else {
		mid := int(math.Ceil(float64(errNum)/2)) - 1
		for i := 0; i < errNum; i++ {
			err := x.Errors[i]
			if mid == i {
				results = append(results, fmt.Sprintf(InfoFormat, x.FileName, err))
			} else {
				results = append(results, fmt.Sprintf(InfoFormat, "", err))
			}
		}
	}
	return results
}

func (x *Xlsx) writeToFile(outFileName string) {
	outFile, operr := os.OpenFile(outFileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if operr != nil {
		return
	}
	defer outFile.Close()

	outFile.WriteString(strings.Join(x.Datas, ""))
	outFile.Sync()
}
