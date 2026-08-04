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
	dir := t.TempDir()
	it, err := All(dir)
	if err != nil {
		t.Fatal(err)
	}

	for ent := range it {
		t.Log(ent.Name())
	}
}
