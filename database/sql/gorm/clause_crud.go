/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gorm

import (
	sqlx "github.com/hopeio/gox/database/sql"
	"gorm.io/gorm/clause"
)

// ByPrimary returns the result.
func ByPrimary(id any) clause.Expression {
	return clause.Eq{Column: clause.PrimaryColumn, Value: id}
}

// ByName returns the result.
func ByName(name string) clause.Expression {
	return clause.Eq{Column: sqlx.ColumnName, Value: name}
}
