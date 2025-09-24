/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package postgres

import (
	"github.com/hopeio/gox/database/sql"
)

const (
	ZeroTimeUCT     = "0001-01-01 00:00:00"
	ZeroTimeUCTZone = ZeroTimeUCT + "+00:00:00"
	ZeroTimeCST     = "0001-01-01 08:05:43"
	ZeroTimeCSTZone = ZeroTimeCST + "+08:05:43"
)

const (
	NotDeletedUCT = sql.ColumnDeletedAt + " = '" + ZeroTimeUCT + "'"
	NotDeletedCST = sql.ColumnDeletedAt + " = '" + ZeroTimeCST + "'"
)

const (
	WithNotDeletedUCT = ` AND ` + NotDeletedUCT
	WithNotDeletedCST = ` AND ` + NotDeletedUCT
)
