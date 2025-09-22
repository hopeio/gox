package gorm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TableName(tx *gorm.DB, name string) *gorm.DB {
	tx.Statement.TableExpr = &clause.Expr{SQL: tx.Statement.Quote(name)}
	tx.Statement.Table = name
	return tx
}
