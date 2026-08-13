/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package sql

import "database/sql"

const (
	existsByColumnSQL = `SELECT EXISTS(SELECT * FROM %s WHERE %s = ?` + WithNotDeleted + ` LIMIT 1)`
	existsSQL         = `SELECT EXISTS(SELECT * FROM %s WHERE %s  LIMIT 1)`
	deleteSQL         = `Update %s SET deleted_at = now() WHERE %s = ?` + WithNotDeleted
)

// ExistsSQL returns the result.
func ExistsSQL(tableName, column string, withDeletedAt bool) string {
	sql := `SELECT EXISTS(SELECT * FROM ` + tableName + ` WHERE ` + column + ` = ?`
	if withDeletedAt {
		sql += WithNotDeleted
	}
	sql += ` LIMIT 1)`
	return sql
}

// DeleteSQL removes or resets state.
func DeleteSQL(tableName, column string) string {
	return `Update ` + tableName + ` SET deleted_at = now() WHERE ` + column + ` = ?` + WithNotDeleted
}

// DeleteByIdSQL removes or resets state.
func DeleteByIdSQL(tableName string) string {
	return `Update ` + tableName + ` SET deleted_at = now() WHERE id = ?`
}

// ExistsByQuerySQL returns the result.
func ExistsByQuerySQL(qsql string) string {
	return `SELECT EXISTS(` + qsql + ` LIMIT 1)`
}

// ExistsByFilterExprsSQL returns the result.
func ExistsByFilterExprsSQL(tableName string, filters FilterExprs) string {
	return `SELECT EXISTS(SELECT * FROM ` + tableName + ` WHERE ` + filters.Build() + ` LIMIT 1)`
}

// ExistsByFilterExprs performs the operation.
func ExistsByFilterExprs(db *sql.DB, tableName string, filters FilterExprs) (bool, error) {
	result := db.QueryRow(`SELECT EXISTS(SELECT * FROM ` + tableName + ` WHERE ` + filters.Build() + ` LIMIT 1)`)
	if err := result.Err(); err != nil {
		return false, err
	}
	var exists bool
	err := result.Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
