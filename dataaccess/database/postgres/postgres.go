/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package postgres

import "strings"

func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "ERROR: duplicate key")
}
