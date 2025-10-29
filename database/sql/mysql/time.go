/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package mysql

import (
	"time"

	timex "github.com/hopeio/gox/time"
)

func Now() string {
	return time.Now().Format(timex.LayoutTimeMacro)
}
