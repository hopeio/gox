/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gorm

import (
	sqlx "github.com/hopeio/gox/database/sql"
	"gorm.io/gorm"
)

// DeleteByPrimary ...
func DeleteByPrimary(db *gorm.DB, tableName string, primary any) error {
	sql := sqlx.DeleteByIdSQL(tableName)
	return db.Exec(sql, primary).Error
}

// Delete ...
func Delete(db *gorm.DB, tableName string, column string, value any) error {
	sql := sqlx.DeleteSQL(tableName, column)
	return db.Exec(sql, value).Error
}

// ExistsByColumn ...
func ExistsByColumn(db *gorm.DB, tableName, column string, value any) (bool, error) {
	return ExistsBySQL(db, sqlx.ExistsSQL(tableName, column, false), value)
}

// ExistsByColumnWithDeletedAt ...
func ExistsByColumnWithDeletedAt(db *gorm.DB, tableName, column string, value any) (bool, error) {
	return ExistsBySQL(db, sqlx.ExistsSQL(tableName, column, true), value)
}

// ExistsBySQL ...
func ExistsBySQL(db *gorm.DB, sql string, value ...any) (bool, error) {
	var exists bool
	err := db.Raw(sql, value...).Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ExistsByQuery ...
func ExistsByQuery(db *gorm.DB, qsql string, value ...any) (bool, error) {
	var exists bool
	err := db.Raw(sqlx.ExistsByQuerySQL(qsql), value...).Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Exists ...
func Exists(db *gorm.DB, tableName, column string, value any, withDeletedAt bool) (bool, error) {
	return ExistsBySQL(db, sqlx.ExistsSQL(tableName, column, withDeletedAt), value)
}

// ExistsByFilterExprs ...
func ExistsByFilterExprs(db *gorm.DB, tableName string, filters sqlx.FilterExprs) (bool, error) {
	var exists bool
	err := db.Raw(sqlx.ExistsByFilterExprsSQL(tableName, filters)).Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetByPrimary ...
func GetByPrimary[T any](db *gorm.DB, primary any) (*T, error) {
	t := new(T)
	err := db.First(t, primary).Error
	return t, err
}
