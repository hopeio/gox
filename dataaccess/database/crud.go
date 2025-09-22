/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package database

import (
	"database/sql"
	sql2 "github.com/hopeio/gox/dataaccess/database/sql"
)

func ExistsByFilterExprs(db *sql.DB, tableName string, filters sql2.FilterExprs) (bool, error) {
	result := db.QueryRow(`SELECT EXISTS(SELECT * FROM ` + tableName + `WHERE ` + filters.Build() + ` LIMIT 1)`)
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
