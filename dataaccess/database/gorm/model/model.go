/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package model

import (
	"github.com/hopeio/gox/types/model"
	"gorm.io/gorm"
	"time"
)

type Model struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"createdAt" gorm:"default:now()"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitzero" gorm:"index"`
}

type ModelTime struct {
	CreatedAt time.Time      `json:"createdAt" gorm:"default:now()"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitzero" gorm:"index"`
}

type Enum struct {
	model.Enum
	Model
}

type EnumValue struct {
	model.EnumValue
	Model
}
