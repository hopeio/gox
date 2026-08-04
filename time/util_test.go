/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package time

import (
	"testing"
	stdtime "time"
)

func TestStrToIntMonth(t *testing.T) {
	if got := StrToIntMonth(June); got != 6 {
		t.Fatalf("June = %d", got)
	}
	if got := StrToIntMonth("NotAMonth"); got != 0 {
		t.Fatalf("unknown month = %d", got)
	}
}

func TestGetYMDAndYM(t *testing.T) {
	ts := stdtime.Date(2024, 3, 5, 15, 0, 0, 0, stdtime.UTC)
	if got := GetYMD(ts, "-"); got != "2024-03-05" {
		t.Fatalf("GetYMD = %q", got)
	}
	if got := GetYM(ts, "/"); got != "2024/03" {
		t.Fatalf("GetYM = %q", got)
	}
}

func TestUnixNano(t *testing.T) {
	ts := UnixNano(123456789)
	if ts.UnixNano() != 123456789 {
		t.Fatalf("got %d", ts.UnixNano())
	}
}

func TestDateRoundTrip(t *testing.T) {
	orig := stdtime.Date(2024, 6, 8, 0, 0, 0, 0, stdtime.UTC)
	d := DateFromTime(orig)
	if d.Time().UTC().Format(stdtime.DateOnly) != "2024-06-08" {
		t.Fatalf("Date round trip failed: %v", d.Time())
	}
}

func TestDateTimeJSON(t *testing.T) {
	dt := DateTime(1717843200)
	b, err := dt.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 12 || b[0] != '"' || b[len(b)-1] != '"' {
		t.Fatalf("unexpected json shape: %s", b)
	}
	var decoded DateTime
	if err := decoded.UnmarshalJSON([]byte(`"2024-06-08 00:00:00"`)); err != nil {
		t.Fatal(err)
	}
	if decoded == 0 {
		t.Fatal("expected non-zero DateTime")
	}
	if err := decoded.UnmarshalJSON([]byte(`1717843200`)); err != nil {
		t.Fatal(err)
	}
}

func TestDateMarshalText(t *testing.T) {
	d := DateFromTime(stdtime.Date(2024, 1, 2, 0, 0, 0, 0, stdtime.UTC))
	b, err := d.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "2024-01-02" {
		t.Fatalf("got %q", b)
	}
}
