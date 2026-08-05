/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package time

import (
	"strings"
	"testing"
	stdtime "time"
)

func TestFormatRelativeTime(t *testing.T) {
	now := stdtime.Now()
	cases := []struct {
		from stdtime.Time
		want string
	}{
		{now.Add(-30 * stdtime.Second), "just now"},
		{now.Add(-5 * stdtime.Minute), "5 minutes ago"},
		{now.Add(-3 * stdtime.Hour), "3 hours ago"},
		{now.Add(-48 * stdtime.Hour), "2 days ago"},
	}
	for _, c := range cases {
		got := FormatRelativeTime(c.from)
		if got != c.want {
			t.Errorf("FormatRelativeTime(%v) = %q, want %q", c.from, got, c.want)
		}
	}
}

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

func TestFormatRelativeTimeLongSpan(t *testing.T) {
	now := stdtime.Now()
	got := FormatRelativeTime(now.Add(-400 * 24 * stdtime.Hour))
	if !strings.Contains(got, "years ago") && !strings.Contains(got, "months ago") {
		t.Fatalf("long span got %q", got)
	}
}
