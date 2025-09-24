/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scope

import (
	sql2 "github.com/hopeio/gox/database/sql"
	"gorm.io/gorm"
)

type Scope func(*gorm.DB) *gorm.DB

func NewScope(field string, op sql2.ConditionOperation, args ...interface{}) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Clauses()
		return db.Where(field+op.SQL(), args...)
	}
}

// var dao ChainScope
// dao.ById(1),ByName("a").Exec(db).First(v)
type ChainScope []func(db *gorm.DB) *gorm.DB

func (c ChainScope) ById(id any) ChainScope {
	return append(c, NewScope(sql2.ColumnId, sql2.Equal, id))
}

func (c ChainScope) ByName(name any) ChainScope {
	return append(c, func(db *gorm.DB) *gorm.DB {
		return db.Where(sql2.NameEqual, name)
	})
}

func (c ChainScope) Exec(db *gorm.DB) *gorm.DB {
	db = db.Scopes(c...)
	return db
}
