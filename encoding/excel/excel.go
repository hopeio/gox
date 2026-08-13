/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package excel

import (
	"fmt"
	"strconv"

	"errors"

	"github.com/xuri/excelize/v2"
)

// NewFile creates and returns a new instance.
func NewFile(sheet string, header []string) (*excelize.File, error) {
	if len(header) == 0 {
		return nil, errors.New("excel: header must not be empty")
	}
	endColumn := ColumnLetter[len(header)-1]
	f := excelize.NewFile()
	//cell style
	style, err := f.NewStyle(Style)
	if err != nil {
		return nil, err
	}
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetColStyle(sheet, "A:"+endColumn, style)
	headerStyle, _ := f.NewStyle(HeaderStyle)
	f.SetCellStyle(sheet, "A1", endColumn+"1", headerStyle)
	for i := range header {
		/*		axis, _ := excelize.CoordinatesToCellName(i+1, 1)
				f.SetCellValue(sheet, axis, header[i])*/
		f.SetCellValue(sheet, ColumnLetter[i]+"1", header[i])
	}
	f.SetRowHeight(sheet, 1, 30)
	return f, nil
}

// NewSheet creates and returns a new instance.
func NewSheet(f *excelize.File, sheet string, header []string) error {
	endColumn := ColumnLetter[len(header)-1]
	//cell style
	style, _ := f.NewStyle(Style)
	headerStyle, _ := f.NewStyle(HeaderStyle)
	f.NewSheet(sheet)
	f.SetColStyle(sheet, "A:"+endColumn, style)
	f.SetCellStyle(sheet, "A1", endColumn+"1", headerStyle)

	for i := range header {
		f.SetCellValue(sheet, ColumnLetter[i]+"1", header[i])
	}
	f.SetRowHeight(sheet, 1, 30)
	return nil
}

// SetNoteRow updates or inserts a value.
func SetNoteRow(sw *excelize.StreamWriter, f *excelize.File, rowNum, maxCol int, rowHeight float64, note string) error {
	styleId, _ := f.NewStyle(noteCellStyle)
	rowNumStr := strconv.Itoa(rowNum)
	_ = sw.MergeCell(ColumnLetter[0]+rowNumStr, ColumnLetter[maxCol-1]+rowNumStr)
	_ = sw.SetRow(ColumnLetter[0]+rowNumStr, []interface{}{
		excelize.Cell{StyleID: styleId, Value: note}, excelize.RowOpts{Height: rowHeight},
	})
	return nil
}

// SetHeaderRow updates or inserts a value.
func SetHeaderRow(sw *excelize.StreamWriter, f *excelize.File, rowNum int, colWith float64, headers []string) {
	styleId, _ := f.NewStyle(headerCellStyle)
	cell := make([]interface{}, 0, len(headers))
	for i := range headers {
		cell = append(cell, excelize.Cell{StyleID: styleId, Value: headers[i]})
	}
	rowNumStr := strconv.Itoa(rowNum)
	if colWith > 0 {
		_ = sw.SetColWidth(1, len(headers), colWith)
	}
	_ = sw.SetRow(ColumnLetter[0]+rowNumStr, cell)
}

// SetBodyRow updates or inserts a value.
func SetBodyRow(sw *excelize.StreamWriter, f *excelize.File, rowNum int, cellVal []any) {
	styleId, _ := f.NewStyle(bodyCellStyle)
	cell := make([]interface{}, 0, len(cellVal))
	for i := range cellVal {
		cell = append(cell, excelize.Cell{StyleID: styleId, Value: cellVal[i]})
	}
	rowNumStr := strconv.Itoa(rowNum)
	_ = sw.SetRow(ColumnLetter[0]+rowNumStr, cell)
}

// SetCellDropListStyle updates or inserts a value.
func SetCellDropListStyle(sw *excelize.StreamWriter, f *excelize.File, hiddenSheetName string, dropListSize int, effectCellStart, effectCellEnd string) {
	definedName := &excelize.DefinedName{
		Name:     hiddenSheetName,
		Comment:  "",
		RefersTo: fmt.Sprintf("%s!$A$1:$A$%d", hiddenSheetName, dropListSize),
		Scope:    "",
	}
	_ = f.SetDefinedName(definedName)
	validation := excelize.NewDataValidation(true)
	validation.Formula1 = fmt.Sprintf("<formula1>%s</formula1>", definedName.Name)
	validation.Sqref = fmt.Sprintf("%s:%s", effectCellStart, effectCellEnd)
	validation.Type = "list"
	validation.SetError(excelize.DataValidationErrorStyleStop, "Invalid input", "Please choose a value from this column's dropdown list")
	_ = f.AddDataValidation(sw.Sheet, validation)
}

// SetBodyRow2 updates or inserts a value.
func SetBodyRow2(sw *excelize.StreamWriter, f *excelize.File, rowNum int, cellVal []string, errStyleLoc map[int]byte) {
	styleId, _ := f.NewStyle(bodyCellStyle)
	errStyleId, _ := f.NewStyle(errorBodyCellStyle)
	cell := make([]interface{}, 0, len(cellVal))
	for i := range cellVal {
		if _, ok := errStyleLoc[i]; ok {
			cell = append(cell, excelize.Cell{StyleID: errStyleId, Value: cellVal[i]})
		} else {
			cell = append(cell, excelize.Cell{StyleID: styleId, Value: cellVal[i]})
		}
	}
	rowNumStr := strconv.Itoa(rowNum)
	_ = sw.SetRow(ColumnLetter[0]+rowNumStr, cell)
}

// MergeCell performs the operation.
func MergeCell(sw *excelize.StreamWriter, startRowNum, endRowNum int, cells []string) {
	for _, cell := range cells {
		startRowNumStr := fmt.Sprintf("%s%d", cell, startRowNum)
		endRowNumStr := fmt.Sprintf("%s%d", cell, endRowNum)
		sw.MergeCell(startRowNumStr, endRowNumStr)
	}
}
