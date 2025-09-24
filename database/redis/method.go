/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package redis

import "github.com/google/uuid"

func LockCmd() (uuid.UUID, string) {
	id := uuid.New()
	cmd := "SETNX " + id.String() + " EXPIRE 100000"
	return id, cmd
}
