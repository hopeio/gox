/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	sqlx "github.com/hopeio/gox/database/sql"
	"gorm.io/gorm/clause"
)

type ChainClause []clause.Expression

func (c *ChainClause) ById(id any) *ChainClause {
	*c = append(*c, clause.Eq{Column: clause.PrimaryColumn, Value: id})
	return c
}

func (c *ChainClause) ByName(name string) *ChainClause {
	*c = append(*c, clause.Eq{Column: sqlx.ColumnName, Value: name})
	return c
}
