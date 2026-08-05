/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package rand

import "math/rand"

// Intn ...
func Intn(min, max int) int {
	return rand.Intn(max-min) + min
}
