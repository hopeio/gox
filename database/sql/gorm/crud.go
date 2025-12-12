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

func DeleteById(db *gorm.DB, tableName string, id uint64) error {
	sql := sqlx.DeleteByIdSQL(tableName)
	return db.Exec(sql, id).Error
}

func Delete(db *gorm.DB, tableName string, column string, value any) error {
	sql := sqlx.DeleteSQL(tableName, column)
	return db.Exec(sql, value).Error
}

func ExistsByColumn(db *gorm.DB, tableName, column string, value interface{}) (bool, error) {
	return ExistsBySQL(db, sqlx.ExistsSQL(tableName, column, false), value)
}

func ExistsByColumnWithDeletedAt(db *gorm.DB, tableName, column string, value interface{}) (bool, error) {
	return ExistsBySQL(db, sqlx.ExistsSQL(tableName, column, true), value)
}

func ExistsBySQL(db *gorm.DB, sql string, value ...any) (bool, error) {
	var exists bool
	err := db.Raw(sql, value...).Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

// 根据查询语句查询数据是否存在
func ExistsByQuery(db *gorm.DB, qsql string, value ...any) (bool, error) {
	var exists bool
	err := db.Raw(sqlx.ExistsByQuerySQL(qsql), value...).Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

func Exists(db *gorm.DB, tableName, column string, value interface{}, withDeletedAt bool) (bool, error) {
	return ExistsBySQL(db, sqlx.ExistsSQL(tableName, column, withDeletedAt), value)
}

func ExistsByFilterExprs(db *gorm.DB, tableName string, filters sqlx.FilterExprs) (bool, error) {
	var exists bool
	err := db.Raw(sqlx.ExistsByFilterExprsSQL(tableName, filters)).Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetById[T any](db *gorm.DB, id any) (*T, error) {
	t := new(T)
	err := db.First(t, id).Error
	return t, err
}
