package main

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/xuri/excelize/v2"
)

var I18nMap sync.Map

func openI18nXlsx(path, lang string) error {
	fileName := path + "/" + lang + ".xlsx"
	_, err := os.Stat(fileName)
	if err == nil {
		f, err := excelize.OpenFile(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		rows, _ := f.GetRows("Sheet1")
		for _, row := range rows {
			if len(row) > 1 && len(row[0]) > 0 && len(row[1]) > 0 {
				I18nMap.Store(row[0], row[1])
			}
		}
	}
	return nil
}

func getI18nString(val string) string {
	if i18nVal, ok := I18nMap.Load(val); ok {
		i18nStr := i18nVal.(string)
		if len(i18nStr) > 0 {
			return i18nStr
		}
	} else {
		I18nMap.Store(val, "")
	}
	return ""
}

func saveI18nXlsx(path, lang string) {
	os.MkdirAll(path, os.ModePerm)
	fileName := path + "/" + lang + ".xlsx"

	f := excelize.NewFile()
	f.SetDocProps(&excelize.DocProperties{
		Creator: "Excel Parser",
	})
	defer f.Close()

	index, _ := f.NewSheet("Sheet1")
	f.SetActiveSheet(index)

	// Set value of a cell.
	keys := make([]string, 0, 64)
	I18nMap.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	sort.Strings(keys)
	for i, k := range keys {
		axis := fmt.Sprintf("A%d", i+1)
		v, _ := I18nMap.Load(k)
		f.SetSheetRow("Sheet1", axis, &[]interface{}{k, v})
	}

	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}
}
