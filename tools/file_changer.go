package tools

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"regexp"
	"strconv"
	"strings"
)

func emptySheetsMap(file *excelize.File) (map[string]bool, error) {
	sheets := file.GetSheetList()
	emptyMap := make(map[string]bool)
	for _, sheet := range sheets {
		rows, err := file.GetRows(sheet)
		if err != nil {
			return nil, fmt.Errorf("error getting rows in empty sheets checking: %v", err)
		}
		emptyMap[sheet] = len(rows) == 0
	}
	return emptyMap, nil
}

func FormattingFileRows(file *excelize.File) ([][]string, error) {
	emptySheetsMap, err := emptySheetsMap(file)
	if err != nil {
		return nil, fmt.Errorf("error getting emptySheetsMap: %v", err)
	}

	newRows := make([][]string, 0)

	for sheet, isEmpty := range emptySheetsMap {
		if !isEmpty {
			for i := 0; i < 2; i++ {
				err := file.RemoveRow(sheet, 1)
				if err != nil {
					fmt.Printf("error removing row: %v\n", err)
				}
			}

			err = file.RemoveCol(sheet, "B")
			if err != nil {
				fmt.Printf("error removing col: %v\n", err)
			}

			var colsToDelete []int

			rows, err := file.GetRows(sheet)
			if err != nil {
				fmt.Printf("Error getting rows: %v\n", err)
			}

			for idx, row := range rows[1] {
				if strings.Contains(row, "год") {
					colsToDelete = append(colsToDelete, idx+1)
				}
			}

			for i := len(colsToDelete) - 1; i >= 0; i-- {
				colName, _ := excelize.ColumnNumberToName(colsToDelete[i])
				if err := file.RemoveCol(sheet, colName); err != nil {
					fmt.Printf("error removing col: %v\n", err)
				}
			}

			rows, err = file.GetRows(sheet)
			if err != nil {
				fmt.Printf("error getting rows: %v\n", err)
			}

			re := regexp.MustCompile(`(\d{4})\s+год`)

			newHeaderValues := make([]string, 0)
			newHeaderValues = append(newHeaderValues, "")

			for i := 1; i < len(rows[0]); {
				cellValue := rows[0][i]
				match := re.FindStringSubmatch(cellValue)

				if len(match) > 0 {
					for q := 1; q <= 4; q++ {
						newHeaderValues = append(newHeaderValues, match[1]+"."+strconv.Itoa(q))
						i++
					}
				}
			}

			newRows = append(newRows, newHeaderValues[0:len(rows[2 : len(rows)-4][0])])
			newRows = append(newRows, rows[2:len(rows)-4]...)

		}
	}
	return newRows, nil
}
