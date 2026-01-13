/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package excel

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func NewFile(sheet string, header []string) (*excelize.File, error) {
	endColumn := ColumnLetter[len(header)-1]
	f := excelize.NewFile()
	//单元格样式
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

func NewSheet(f *excelize.File, sheet string, header []string) error {
	endColumn := ColumnLetter[len(header)-1]
	//单元格样式
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

// SetNoteRow
//
//	为excel stream设置提醒行
//
// 参数
//
//	sw, excel StreamWriter
//	rowNum, 要作为提醒文本行的excel行号，行号是从1开始，第一行行号为1
//	maxCol, 提醒文本最多占用多少格
//	rowHeight, 行高
//	note, 提醒文案
func SetNoteRow(sw *excelize.StreamWriter, f *excelize.File, rowNum, maxCol int, rowHeight float64, note string) error {
	styleId, _ := f.NewStyle(noteCellStyle)
	rowNumStr := strconv.Itoa(rowNum)
	_ = sw.MergeCell(ColumnLetter[0]+rowNumStr, ColumnLetter[maxCol-1]+rowNumStr)
	_ = sw.SetRow(ColumnLetter[0]+rowNumStr, []interface{}{
		excelize.Cell{StyleID: styleId, Value: note}, excelize.RowOpts{Height: rowHeight},
	})
	return nil
}

// SetHeaderRow
//
//	为excel stream设置表头
//
// 参数
//
//	sw, excel StreamWriter
//	rowNum, 要作为表头的excel行号，行号是从1开始的，第一行行号为1
//	colWidth, 列宽, 为0时表示采用默认不设置
//	headers, 表头文案列表
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

// SetBodyRow
//
//	为excel stream设置行
//
// 参数
//
//	sw, excel StreamWriter
//	rowNum, 要设置的excel行号，行号是从1开始的，第一行行号为1
//	cellVal, 单元格内容
func SetBodyRow(sw *excelize.StreamWriter, f *excelize.File, rowNum int, cellVal []any) {
	styleId, _ := f.NewStyle(bodyCellStyle)
	cell := make([]interface{}, 0, len(cellVal))
	for i := range cellVal {
		cell = append(cell, excelize.Cell{StyleID: styleId, Value: cellVal[i]})
	}
	rowNumStr := strconv.Itoa(rowNum)
	_ = sw.SetRow(ColumnLetter[0]+rowNumStr, cell)
}

// SetCellDropListStyle 设置单元格下拉框选项样式，根据提供的下拉列表枚举值，作为单元格内容待候选值
//
//		当下拉框列表太大时，无法直接采用AddDataValidation的形式添加
//		这里通过将下拉枚举列表提前写入到隐藏的Sheet中，然后引用此隐藏Sheet再加载的方式实现
//	 @param sw excel streamWriter
//	 @param f excel file
//	 @param hiddenSheetName 存在此f中的隐藏Sheet
//	 @param dropListSize 下拉列表容量大小
//	 @param effectCellStart effectCellEnd 下拉框选项样式作用的单元格范围 起始位置 结束位置 例如 K2 K31
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
	validation.SetError(excelize.DataValidationErrorStyleStop, "错误的输入内容", "请输入此列下拉框列表中已有的值")
	_ = f.AddDataValidation(sw.Sheet, validation)
}

// SetBodyRow2
//
//	为excel stream设置行
//
// 参数
//
//	sw, excel StreamWriter
//	rowNum, 要设置的excel行号，行号是从1开始的，第一行行号为1
//	cellVal, 单元格内容
//	errStyleLoc, 红色错误样式的单元格位置
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

// 合并单元格
// 参数
//
//	sw, excel StreamWriter
//	startRowNum, 要合并开始的excel行号
//	endRowNum, 要合并结束的excel行号
//	cells, 要合并单元格的列("A","B","C")
func MergeCell(sw *excelize.StreamWriter, startRowNum, endRowNum int, cells []string) {
	for _, cell := range cells {
		startRowNumStr := fmt.Sprintf("%s%d", cell, startRowNum)
		endRowNumStr := fmt.Sprintf("%s%d", cell, endRowNum)
		sw.MergeCell(startRowNumStr, endRowNumStr)
	}
}
