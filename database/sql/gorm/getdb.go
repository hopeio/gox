/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gorm

import (
	"context"

	loggerx "github.com/hopeio/gox/database/sql/gorm/logger"
	"github.com/hopeio/gox/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDBWithLogger(db *gorm.DB, log *log.Logger, conf *logger.Config) *gorm.DB {
	return db.Session(&gorm.Session{
		Logger: &loggerx.Logger{Logger: log.Logger,
			Config: conf,
		}})
}

func NewDBWithContext(db *gorm.DB, ctx context.Context) *gorm.DB {
	return db.Session(&gorm.Session{Context: ctx})
}

func NewTraceDB(db *gorm.DB, ctx context.Context, traceId string) *gorm.DB {
	return db.Session(&gorm.Session{Context: loggerx.SetTranceId(ctx, traceId), NewDB: true})
}
