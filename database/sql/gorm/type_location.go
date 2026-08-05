/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Location struct {
	X, Y float64
}

// GormDataType ...
func (loc Location) GormDataType() string {
	return "geometry"
}

// GormValue ...
func (loc Location) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL:  "ST_PointFromText(?)",
		Vars: []interface{}{fmt.Sprintf("POINT(%f %f)", loc.X, loc.Y)},
	}
}

// Scan ...
func (loc *Location) Scan(v interface{}) error {
	// Scan a value into struct from database driver
	return nil
}
