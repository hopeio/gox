/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package time

import (
	"testing"
	stdtime "time"
)

func TestParseTimeDateTimeDate(t *testing.T) {
	if tm, err := ParseTime("15:04:05"); err != nil || tm.Hour() != 15 {
		t.Fatalf("ParseTime err=%v tm=%v", err, tm)
	}
	if tm, err := ParseDateTime("2024-06-08 12:00:00"); err != nil || tm.Year() != 2024 {
		t.Fatalf("ParseDateTime err=%v tm=%v", err, tm)
	}
	if tm, err := ParseDate("2024-06-08"); err != nil || tm.Month() != stdtime.June {
		t.Fatalf("ParseDate err=%v tm=%v", err, tm)
	}
}
