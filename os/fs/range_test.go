/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fs

import (
	"testing"
)

func TestRange(t *testing.T) {
	it, err := All("D:\\data")
	if err != nil {
		t.Error(err)
	}

	for ent := range it {
		t.Log(ent.Name())
	}
	if err != nil {
		t.Error(err)
	}
}
