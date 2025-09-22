/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package serializer

import "gorm.io/gorm/schema"

func init() {
	schema.RegisterSerializer("json", JSONSerializer{})

	schema.RegisterSerializer("string_array", StringArraySerializer{})

}
