/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package model

import (
	"github.com/hopeio/gox/database/sql/datatypes"
)

type Enum struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Group uint32 `json:"group" gorm:"uniqueIndex:idx_group_name;not null;default:0;comment:枚举组"`
	Name  string `json:"name" gorm:"uniqueIndex:idx_group_name;not null;comment:名称"`
	Type  uint32 `json:"type" gorm:"comment:类型"`
}

type PostgresEnum struct {
	ID    uint `gorm:"primaryKey"`
	Enums datatypes.StringArray
}

type EnumValue struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	EnumID uint   `json:"enumId" gorm:"index;not null;comment:枚举id"`
	Index  uint32 `json:"index" gorm:"comment:索引"`
	Value  string `json:"value" gorm:"not null;comment:值"`
}
