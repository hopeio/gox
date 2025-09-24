package mysql

const ShowTables = `SHOW TABLES`

func ShowColumns(table string) string {
	return "`SHOW COLUMNS FROM " + table
}
