package mysql

const ShowTables = `SHOW TABLES`

// ShowColumns returns the result.
func ShowColumns(table string) string {
	return "`SHOW COLUMNS FROM " + table
}
