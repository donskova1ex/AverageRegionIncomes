package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	filepath := "C:/Users/Admin/Desktop/AverageIncomes.xlsx"
	file, err := excelize.OpenFile(filepath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	sheets := file.GetSheetList()

	fmt.Printf("Работаем с листом: %s\n", sheets[0])
	rows, err := file.GetRows(sheets[0])
	if err != nil {
		fmt.Println(err)
	}
	for srtIndex, row := range rows {
		fmt.Println(srtIndex)
		fmt.Println(row)
	}
}
