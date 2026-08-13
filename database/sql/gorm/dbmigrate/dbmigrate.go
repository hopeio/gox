// Package dbmigrate 极简 schema 迁移：版本表 + 只跑一次的 SQL 列表。
// AutoMigrate 只管建新表；对已有库的列/索引变更登记在这里，可追溯、不重复执行。
package dbmigrate

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Migration 一次结构变更；Version 全局唯一且不可复用（约定 YYYYMMDDNN_描述）。
type Migration struct {
	Version string
	SQL     []string
}

const table = "schema_migrations"

// Run 依序执行未应用的迁移，每个迁移一个事务，失败即停。
func Run(ctx context.Context, db *gorm.DB, migrations []Migration) error {
	if err := db.WithContext(ctx).Exec(
		`CREATE TABLE IF NOT EXISTS ` + table + ` (
			version varchar(64) PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`,
	).Error; err != nil {
		return fmt.Errorf("create %s: %w", table, err)
	}
	for _, m := range migrations {
		var n int64
		if err := db.WithContext(ctx).Table(table).
			Where("version = ?", m.Version).Count(&n).Error; err != nil {
			return fmt.Errorf("check %s: %w", m.Version, err)
		}
		if n > 0 {
			continue
		}
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			for _, s := range m.SQL {
				if err := tx.Exec(s).Error; err != nil {
					return err
				}
			}
			return tx.Exec(`INSERT INTO `+table+` (version) VALUES (?)`, m.Version).Error
		}); err != nil {
			return fmt.Errorf("migration %s: %w", m.Version, err)
		}
	}
	return nil
}
