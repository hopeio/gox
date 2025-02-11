/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scope

import (
	"github.com/hopeio/utils/dao/database/sql"
	"gorm.io/gorm"
)

type Scope func(*gorm.DB) *gorm.DB

func NewScope(field string, op sql.ConditionOperation, args ...interface{}) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Clauses()
		return db.Where(field+op.SQL(), args...)
	}
}

// var dao ChainScope
// dao.ById(1),ByName("a").Exec(db).First(v)
type ChainScope []func(db *gorm.DB) *gorm.DB

func (c ChainScope) ById(id any) ChainScope {
	return append(c, NewScope(sql.ColumnId, sql.Equal, id))
}

func (c ChainScope) ByName(name any) ChainScope {
	return append(c, func(db *gorm.DB) *gorm.DB {
		return db.Where(sql.NameEqual, name)
	})
}

func (c ChainScope) Exec(db *gorm.DB) *gorm.DB {
	db = db.Scopes(c...)
	return db
}
