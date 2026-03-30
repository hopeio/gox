/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package redis

import (
	"testing"
	"time"
)

type Foo struct {
	Time time.Time
	Bar
}

type Bar struct {
	Int int
}

func TestHashEncode(t *testing.T) {
	u := &Foo{Time: time.Now(), Bar: Bar{Int: 1}}
	redisArgs := HashEncode(u)
	for i := 0; i < len(redisArgs); i += 2 {
		t.Log(redisArgs[i], redisArgs[i+1])
	}
}
