/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package model

import (
	"github.com/hopeio/gox/types/model"
	"gorm.io/gorm"
)

type Dict struct {
	model.Dict
	ModelTime
}

func DictGetValue(db *gorm.DB, typ int, key string) (string, error) {
	var value string
	err := db.Table(`dict`).Select(`value`).Where(`type = ? AND key=?`, typ, key).Scan(&value).Error
	if err != nil {
		return "", err
	}
	return value, nil
}

func DictSetValue(db *gorm.DB, typ int, key, value string) error {
	return db.Table(`dict`).Where(`type = ? AND key=?`, typ, key).UpdateColumn("value", value).Error
}
