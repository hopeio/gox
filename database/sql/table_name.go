/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package sql

var (
	tableName = [...]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
)

func TableName(name string, id uint64) string {
	if id < 2000_00000 {
		return name
	}
	if id < 2_0000_00000 {
		return name + "_" + tableName[id/2000_00000-1]
	}
	return name + "_" + string(byte(id/2000_00000+49))
}
