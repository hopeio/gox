/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package time

import (
	"testing"
)

type stubTime struct {
	seconds int64
	nanos   int32
}

func (s stubTime) GetSeconds() int64 { return s.seconds }
func (s stubTime) GetNanos() int32   { return s.nanos }

func TestCheckValid(t *testing.T) {
	valid := stubTime{seconds: 0, nanos: 0}
	if !IsValid(valid) {
		t.Fatal("zero timestamp should be valid")
	}
	if err := CheckValid(valid); err != nil {
		t.Fatal(err)
	}

	if Check(nil) != InvalidNil {
		t.Fatal("nil should be InvalidNil")
	}
	if IsValid(nil) {
		t.Fatal("nil should not be valid")
	}
	if err := CheckValid(nil); err == nil || err.Error() != "invalid nil Timestamp" {
		t.Fatalf("err = %v", err)
	}

	under := stubTime{seconds: -62135596801, nanos: 0}
	if Check(under) != InvalidUnderflow {
		t.Fatal("expected underflow")
	}
	over := stubTime{seconds: 253402300800, nanos: 0}
	if Check(over) != InvalidOverflow {
		t.Fatal("expected overflow")
	}
	badNanos := stubTime{seconds: 0, nanos: 1e9}
	if Check(badNanos) != InvalidNanos {
		t.Fatal("expected invalid nanos")
	}
}
