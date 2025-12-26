package core

import (
	"bytes"
	"encoding/json"
	"strings"
)

func (j *JsonFormater) formatRows() {
	// 复用 datas
	j.line = 0
	j.clearData()

	// data
	if j.Vertical {
		for _, col := range j.Rows {
			j.line++
			j.formatData(j.RootField, col, 0)
		}
	} else {
		j.appendData("[\n")
		for _, row := range j.Rows {
			j.line++
			key := row[0]
			if strings.HasPrefix(key, "//") || key == "" {
				continue
			}
			j.appendIndent(1)
			j.formatData(j.RootField, row, 1)
			j.appendData(",\n")
		}
		j.replaceTail("\n")
		j.appendData("]")
	}
}

// datas
func (j *JsonFormater) formatData(field *Field, row []string, depth int) {
	fkind := field.Kind
	switch fkind {
	case TArray:
		j.appendData("[")
		j.appendEOL()
		for _, f := range field.Vals {
			j.appendIndent(depth + 1)
			j.formatData(f, row, depth+1)
			j.appendComma()
		}
		j.replaceComma()
		j.appendIndent(depth)
		j.appendData("]")
	case TMap:
		j.appendData("{")
		j.appendEOL()
		for i, k := range field.Keys {
			j.appendIndent(depth + 1)
			j.appendData("\"")
			j.appendData(row[k.Index])
			j.appendData("\":")

			v := field.Vals[i]
			j.formatData(v, row, depth+1)
			j.appendComma()
		}
		j.replaceComma()
		j.appendIndent(depth)
		j.appendData("}")
	case TStruct:
		j.appendData("{")
		j.appendEOL()
		for _, f := range field.Vals {
			if f.isHitMode(j.mode) {
				j.appendIndent(depth + 1)
				j.appendData("\"")
				j.appendData(f.Name)
				j.appendData("\":")
				j.formatData(f, row, depth+1)
				j.appendComma()
			}
		}
		j.replaceComma()
		j.appendIndent(depth)
		j.appendData("}")
	case TJson:
		if len(row) > field.Index {
			s := row[field.Index]
			if field.I18n {
				var result any
				json.Unmarshal([]byte(s), &result)
				j.updateI18nJson(field, field.Vtype, result)
				bytes, err := json.Marshal(result)
				if err == nil {
					s = string(bytes)
				}
			}
			var out bytes.Buffer
			if GFlags.Compact {
				json.Compact(&out, []byte(s))
				j.appendData(out.String())
			} else if GFlags.Pretty {
				json.Indent(&out, []byte(s), getIndent(depth), "  ")
				j.appendData(out.String())
			} else {
				j.appendData(s)
			}

		} else {
			j.appendData("null")
		}
	default:
		s := ""
		if len(row) > field.Index {
			s = row[field.Index]
		}
		j.appendData(field.formatValue(s))
	}
}

func (j *JsonFormater) updateI18nJson(field *Field, t *Type, obj any) {
	var vt *Type
	if t != nil {
		vt = t.Vtype
	}
	if vt == nil {
		return
	}

	switch val := obj.(type) {
	case map[any]any:
		for k, v := range val {
			if vt.isI18nString() {
				val[k] = getI18nString(v.(string), field, j.line+HeadLineNum)
			} else {
				j.updateI18nJson(field, vt, v)
			}
		}
	case map[string]any:
		for k, v := range val {
			if vt.isI18nString() {
				val[k] = getI18nString(v.(string), field, j.line+HeadLineNum)
			} else {
				j.updateI18nJson(field, vt, v)
			}
		}
	case []any:
		for i, v := range val {
			if vt.isI18nString() {
				val[i] = getI18nString(v.(string), field, j.line+HeadLineNum)
			} else {
				j.updateI18nJson(field, vt, v)
			}
		}
	}
}
