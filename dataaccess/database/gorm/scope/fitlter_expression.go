/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scope

import (
	dbi "github.com/hopeio/gox/dataaccess/database/sql"
	"gorm.io/gorm"
	"strings"
)

type FilterExprs dbi.FilterExprs

func (f FilterExprs) Build(db *gorm.DB) *gorm.DB {
	for _, filter := range f {
		filter.Field = strings.TrimSpace(filter.Field)

		if filter.Field == "" {
			continue
		}

		db = db.Where(filter.Field+" "+filter.Operation.SQL(), filter.Value...)
	}
	return db
}
