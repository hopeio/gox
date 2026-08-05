package mysql

const ShowTables = `SHOW TABLES`

// ShowColumns ...
func ShowColumns(table string) string {
	return "`SHOW COLUMNS FROM " + table
}
