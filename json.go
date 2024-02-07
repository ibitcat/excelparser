package main

import (
	"bytes"
	"encoding/json"
	"strings"
)

type JsonFormater struct {
	*Xlsx
	mode string
}

func (j *JsonFormater) formatRows() {
	// 复用 datas
	j.clearData()

	// data
	if j.Vertical {
		for _, col := range j.Rows {
			j.formatData(j.RootField, col, 0)
		}
	} else {
		j.appendData("[\n")
		for _, row := range j.Rows {
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
				var result interface{}
				json.Unmarshal([]byte(s), &result)
				j.updateI18nJson(field.Vtype, result)
				bytes, err := json.Marshal(result)
				if err == nil {
					s = string(bytes)
				}
			}
			var out bytes.Buffer
			if FlagCompact {
				json.Compact(&out, []byte(s))
				j.appendData(out.String())
			} else if FlagIndent {
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

func (j *JsonFormater) updateI18nJson(t *Type, obj interface{}) {
	var vt *Type
	if t != nil {
		vt = t.Vtype
	}
	if vt == nil {
		return
	}

	switch val := obj.(type) {
	case map[interface{}]interface{}:
		for k, v := range val {
			if vt.isI18nString() {
				val[k] = getI18nString(v.(string))
			} else {
				j.updateI18nJson(vt, v)
			}
		}
	case map[string]interface{}:
		for k, v := range val {
			if vt.isI18nString() {
				val[k] = getI18nString(v.(string))
			} else {
				j.updateI18nJson(vt, v)
			}
		}
	case []interface{}:
		for i, v := range val {
			if vt.isI18nString() {
				val[i] = getI18nString(v.(string))
			} else {
				j.updateI18nJson(vt, v)
			}
		}
	}
}
