/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"fmt"
)


var (
	ErrNotFound            = fmt.Errorf("not found")
	ErrRangeNotSatisfiable = fmt.Errorf("range not satisfiable")
)
