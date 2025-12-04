/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package sql

const (
	Mysql    = "mysql"
	Postgres = "postgres"
	Sqlite   = "sqlite"
)

const (
	TmFmtWithMS = "2006-01-02 15:04:05.999"
	NullStr     = "NULL"
)

const (
	ColumnDeletedAt = "deleted_at"
	ColumnId        = "id"
	ColumnName      = "name"
)

const (
	ExprEqual    = " = ?"
	ExprNotEqual = " != ?"
	ExprGreater  = " > ?"
)

const (
	Insert    = "INSERT"
	Select    = "SELECT"
	Update    = "UPDATE"
	Delete    = "DELETE"
	LeftJoin  = "LEFT JOIN"
	RightJoin = "RIGHT JOIN"
	InnerJoin = "INNER JOIN"
	Limit     = `LIMIT %d`
	Offset    = `OFFSET %d`
	Limit1    = `LIMIT 1`
)

const (
	NotDeleted     = ColumnDeletedAt + " IS " + NullStr
	WithNotDeleted = ` AND ` + NotDeleted
)

const (
	IdEqual   = ColumnId + ExprEqual
	NameEqual = ColumnName + ExprEqual
)
