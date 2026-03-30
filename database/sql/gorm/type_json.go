/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gorm

import (
	"context"
	"database/sql/driver"

	sqlx "github.com/hopeio/gox/database/sql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Json[T any] sqlx.Json[T]

func (*Json[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case sqlx.Sqlite, sqlx.Mysql:
		return "json"
	case sqlx.Postgres:
		return "jsonb"
	}
	return ""
}

func (j *Json[T]) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	v, _ := (*sqlx.Json[T])(j).Value()
	return clause.Expr{
		SQL:  "?",
		Vars: []any{v},
	}
}

func (j *Json[T]) Value() (driver.Value, error) {
	// Scan a value into struct from database driver
	return (*sqlx.Json[T])(j).Value()
}

func (j *Json[T]) Scan(v any) error {
	// Scan a value into struct from database driver
	return (*sqlx.Json[T])(j).Scan(v)
}
