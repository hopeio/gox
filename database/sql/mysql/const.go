/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package mysql

import (
	sqlx "github.com/hopeio/gox/database/sql"
)

const (
	DateTimeZero  = "0001-01-01 00:00:00"
	TimeStampZero = "0000-00-00 00:00:00"
)

const (
	NotDeleted = sqlx.ColumnDeletedAt + " = '" + DateTimeZero + "'"
)
