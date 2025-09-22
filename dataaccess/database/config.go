/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package database

type Config struct {
	Type, Charset, Database, TimeZone string
	Host                              string `flag:"name:db_host;usage:数据库host"`
	User, Password                    string
	TimeFormat                        string
	MaxIdleConns, MaxOpenConns        int
	Port                              int32
}

const (
	Mysql    = "mysql"
	Postgres = "postgres"
	Sqlite   = "sqlite"
)
