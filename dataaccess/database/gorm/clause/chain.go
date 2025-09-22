/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	dbi "github.com/hopeio/gox/dataaccess/database/sql"
	"gorm.io/gorm/clause"
)

type ChainClause []clause.Interface

func (c ChainClause) ById(id any) ChainClause {
	return append(c, clause.Where{Exprs: []clause.Expression{clause.Eq{Column: clause.PrimaryColumn, Value: id}}})
}

func (c ChainClause) ByName(name string) ChainClause {
	return append(c, clause.Where{Exprs: []clause.Expression{clause.Eq{Column: dbi.ColumnName, Value: name}}})
}
