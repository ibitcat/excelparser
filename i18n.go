package main

import (
	"fmt"
	"os"
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

func saveI18nXlsx(path, lang string) {
	os.MkdirAll(path, os.ModePerm)
	fileName := path + "/" + lang + ".xlsx"

	f := excelize.NewFile()
	defer f.Close()

	index, _ := f.NewSheet("Sheet1")
	f.SetActiveSheet(index)

	// Set value of a cell.
	idx := 1
	I18nMap.Range(func(key, value interface{}) bool {
		axis := fmt.Sprintf("A%d", idx)
		f.SetSheetRow("Sheet1", axis, &[]interface{}{key, value})
		idx++
		return true
	})
	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}
}
