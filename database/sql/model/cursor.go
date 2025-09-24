/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package model

func EndCallbackSQL(typ string) string {
	return `UPDATE cursor SET prev = next, cursor = '' WHERE type = '` + typ + `'`
}
