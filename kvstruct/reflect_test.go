/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package kvstruct

import (
	"encoding"
	"reflect"
	"testing"
	"time"
)

type textField string

func (t *textField) UnmarshalText(b []byte) error {
	*t = textField("txt:" + string(b))
	return nil
}

var _ encoding.TextUnmarshaler = (*textField)(nil)

func setAndCheck[T any](t *testing.T, val string, want T) {
	t.Helper()
	var got T
	if err := ParseStringSetReflectValue(reflect.ValueOf(&got).Elem(), val, nil); err != nil {
		t.Fatalf("ParseStringSetReflectValue(%q) err = %v", val, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseStringSetReflectValue(%q) = %#v, want %#v", val, got, want)
	}
}

func TestParseStringSetReflectValue_Empty(t *testing.T) {
	var i int
	if err := ParseStringSetReflectValue(reflect.ValueOf(&i).Elem(), "", nil); err != nil {
		t.Fatal(err)
	}
	if i != 0 {
		t.Fatalf("empty string should not change value, got %d", i)
	}
}

func TestParseStringSetReflectValue_Integers(t *testing.T) {
	setAndCheck(t, "42", int(42))
	setAndCheck(t, "127", int8(127))
	setAndCheck(t, "32767", int16(32767))
	setAndCheck(t, "2147483647", int32(2147483647))
	setAndCheck(t, "9223372036854775807", int64(9223372036854775807))
}

func TestParseStringSetReflectValue_IntOverflow(t *testing.T) {
	var v int8
	if err := ParseStringSetReflectValue(reflect.ValueOf(&v).Elem(), "128", nil); err == nil {
		t.Fatal("want overflow error for int8")
	}
}

func TestParseStringSetReflectValue_Unsigned(t *testing.T) {
	setAndCheck(t, "255", uint8(255))
	setAndCheck(t, "65535", uint16(65535))
	setAndCheck(t, "4294967295", uint32(4294967295))
	setAndCheck(t, "18446744073709551615", uint64(18446744073709551615))
	setAndCheck(t, "99", uint(99))
}

func TestParseStringSetReflectValue_UnsignedOverflow(t *testing.T) {
	var v uint8
	if err := ParseStringSetReflectValue(reflect.ValueOf(&v).Elem(), "256", nil); err == nil {
		t.Fatal("want overflow error for uint8")
	}
}

func TestParseStringSetReflectValue_Bool(t *testing.T) {
	setAndCheck(t, "true", true)
	setAndCheck(t, "false", false)
	setAndCheck(t, "1", true)
	var v bool
	if err := ParseStringSetReflectValue(reflect.ValueOf(&v).Elem(), "maybe", nil); err == nil {
		t.Fatal("want parse error for invalid bool")
	}
}

func TestParseStringSetReflectValue_Float(t *testing.T) {
	setAndCheck(t, "3.14", float32(3.14))
	setAndCheck(t, "2.71828", float64(2.71828))
}

func TestParseStringSetReflectValue_String(t *testing.T) {
	setAndCheck(t, "hello", "hello")
}

func TestParseStringSetReflectValue_Duration(t *testing.T) {
	setAndCheck(t, "500ms", time.Duration(500*time.Millisecond))
	setAndCheck(t, "1h30m", time.Duration(90*time.Minute))
	var d time.Duration
	if err := ParseStringSetReflectValue(reflect.ValueOf(&d).Elem(), "not-a-duration", nil); err == nil {
		t.Fatal("want parse error for invalid duration")
	}
}

type timeUnixCfg struct {
	T time.Time `format:"unix"`
}

type timeUnixNanoCfg struct {
	T time.Time `format:"unixnano"`
}

type timeUTCCfg struct {
	T time.Time `format:"2006-01-02T15:04:05Z07:00" time_utc:"true"`
}
func TestParseStringSetReflectValue_TimeRFC3339(t *testing.T) {
	field, _ := reflect.TypeOf(timeUTCCfg{}).FieldByName("T")
	var got time.Time
	if err := ParseStringSetReflectValue(reflect.ValueOf(&got).Elem(), "2024-06-08T04:00:00Z", &field); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, 6, 8, 4, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseStringSetReflectValue_TimeViaTextUnmarshaler(t *testing.T) {
	// time.Time implements encoding.TextUnmarshaler; ParseStringSetReflectValue
	// uses that path and ignores struct field format tags.
	var got time.Time
	if err := ParseStringSetReflectValue(reflect.ValueOf(&got).Elem(), "2024-06-08T04:00:00Z", nil); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, 6, 8, 4, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestSetTimeFieldUnixDirect(t *testing.T) {
	field, _ := reflect.TypeOf(timeUnixCfg{}).FieldByName("T")
	if field.Tag.Get("format") != "unix" {
		t.Fatalf("format tag = %q", field.Tag.Get("format"))
	}
	var got time.Time
	if err := setTimeField("0", &field, reflect.ValueOf(&got).Elem()); err != nil {
		t.Fatal(err)
	}
	if got.Unix() != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestSetTimeFieldUnixNanoDirect(t *testing.T) {
	field, _ := reflect.TypeOf(timeUnixNanoCfg{}).FieldByName("T")
	var got time.Time
	if err := setTimeField("1000000000", &field, reflect.ValueOf(&got).Elem()); err != nil {
		t.Fatal(err)
	}
	if got.Unix() != 1 || got.Nanosecond() != 0 {
		t.Fatalf("unixnano got %v", got)
	}
}

func TestParseStringSetReflectValue_Slice(t *testing.T) {
	var got []int
	if err := ParseStringSetReflectValue(reflect.ValueOf(&got).Elem(), "1,2,3", nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("got %#v", got)
	}
}

func TestParseStringSetReflectValue_Array(t *testing.T) {
	var got [2]string
	if err := ParseStringSetReflectValue(reflect.ValueOf(&got).Elem(), "a,b", nil); err != nil {
		t.Fatal(err)
	}
	if got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %#v", got)
	}
}

func TestParseStringSetReflectValue_ArrayLengthMismatch(t *testing.T) {
	var got [3]int
	if err := ParseStringSetReflectValue(reflect.ValueOf(&got).Elem(), "1,2", nil); err == nil {
		t.Fatal("want length mismatch error")
	}
}

func TestParseStringSetReflectValue_NestedSliceUnsupported(t *testing.T) {
	var got [][]int
	if err := ParseStringSetReflectValue(reflect.ValueOf(&got).Elem(), "1,2", nil); err == nil {
		t.Fatal("want unsupported sub type error")
	}
}

func TestParseStringSetReflectValue_TextUnmarshaler(t *testing.T) {
	var got textField
	if err := ParseStringSetReflectValue(reflect.ValueOf(&got).Elem(), "x", nil); err != nil {
		t.Fatal(err)
	}
	if got != "txt:x" {
		t.Fatalf("got %q", got)
	}
}

func TestParseStringSetReflectValue_UnknownType(t *testing.T) {
	var ch chan int
	if err := ParseStringSetReflectValue(reflect.ValueOf(&ch).Elem(), "1", nil); err != errUnknownType {
		t.Fatalf("err = %v, want errUnknownType", err)
	}
}

func TestParseStringsSetReflectValue(t *testing.T) {
	var got []string
	if err := ParseStringsSetReflectValue(reflect.ValueOf(&got).Elem(), []string{"a", "b"}, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("got %#v", got)
	}
}

func TestParseStringsSetReflectValue_Scalar(t *testing.T) {
	var got int
	if err := ParseStringsSetReflectValue(reflect.ValueOf(&got).Elem(), []string{"7"}, nil); err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Fatalf("got %d", got)
	}
}

func TestParseStringsSetReflectValue_Empty(t *testing.T) {
	var got int
	if err := ParseStringsSetReflectValue(reflect.ValueOf(&got).Elem(), nil, nil); err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Fatalf("empty vals should no-op, got %d", got)
	}
}
