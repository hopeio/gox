package excel

import (
	"database/sql"

	"github.com/xuri/excelize/v2"
)

// TODO
type NullString struct {
	sql.Null[string]
}

func (n NullString) String() string {
	return n.V
}

func Export(rows *sql.Rows, filename string) error {
	f := excelize.NewFile()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	/*	types,err:=rows.ColumnTypes()
		if err != nil {
			return err
		}*/
	err = f.SetSheetRow("Sheet1", "A1", &columns)
	if err != nil {
		return err
	}
	columnValues := make([]NullString, len(columns))
	columnValuePtrs := make([]any, len(columns))
	for i := range columnValues {
		columnValuePtrs[i] = &columnValues[i]
	}
	row := 2
	for rows.Next() {
		err = rows.Scan(columnValuePtrs...)
		if err != nil {
			return err
		}
		cell, err := excelize.CoordinatesToCellName(1, row)
		if err != nil {
			return err
		}
		err = f.SetSheetRow("Sheet1", cell, &columnValues)
		if err != nil {
			return err
		}
		row++
	}
	err = rows.Close()
	if err != nil {
		return err
	}
	return f.SaveAs(filename)
}
