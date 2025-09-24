/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package postgres

import "time"

func Now() string {
	return time.Now().Format(time.RFC3339Nano)
}
