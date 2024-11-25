/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package model

import (
	model1 "github.com/hopeio/utils/dao/database/model"
	"github.com/hopeio/utils/types/model"
	"gorm.io/gorm"
)

type Cursor struct {
	model.Cursor
	ModelTime
}

func GetCursor(db *gorm.DB, typ string) (*Cursor, error) {
	var cursor Cursor
	err := db.Where(`type = ?`, typ).First(&cursor).Error
	if err != nil {
		return nil, err
	}
	return &cursor, nil
}

func EndCallback(db *gorm.DB, typ string) {
	db.Exec(model1.EndCallbackSQL(typ))
}
