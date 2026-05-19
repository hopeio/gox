package types

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestBool_IsTrue(t *testing.T) {
	b := Bool(1)
	if !b.IsTrue() {
		t.Error("Bool(1).IsTrue() should be true")
	}
	b = Bool(0)
	if b.IsTrue() {
		t.Error("Bool(0).IsTrue() should be false")
	}
}

func TestBool_IsFalse(t *testing.T) {
	b := Bool(2)
	if !b.IsFalse() {
		t.Error("Bool(2).IsFalse() should be true")
	}
	b = Bool(0)
	if b.IsFalse() {
		t.Error("Bool(0).IsFalse() should be false")
	}
}

func TestBool_IsNone(t *testing.T) {
	b := Bool(0)
	if !b.IsNone() {
		t.Error("Bool(0).IsNone() should be true")
	}
	b = Bool(1)
	if b.IsNone() {
		t.Error("Bool(1).IsNone() should be false")
	}
}

func TestBool_MarshalJSON(t *testing.T) {
	tests := []struct {
		b    Bool
		want string
	}{
		{Bool(0), "null"},
		{Bool(1), "true"},
		{Bool(2), "false"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.b)
		if err != nil {
			t.Errorf("MarshalJSON(%d): %v", tt.b, err)
		}
		if string(data) != tt.want {
			t.Errorf("MarshalJSON(%d) = %q, want %q", tt.b, string(data), tt.want)
		}
	}
}

func TestBool_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  Bool
	}{
		{"true", Bool(1)},
		{"false", Bool(2)},
		{"null", Bool(0)},
		{"1", Bool(1)},
		{"2", Bool(2)},
	}
	for _, tt := range tests {
		var b Bool
		err := json.Unmarshal([]byte(tt.input), &b)
		if err != nil {
			t.Errorf("UnmarshalJSON(%q): %v", tt.input, err)
		}
		if b != tt.want {
			t.Errorf("UnmarshalJSON(%q) = %d, want %d", tt.input, b, tt.want)
		}
	}
}

func TestBool_String(t *testing.T) {
	if Bool(0).String() != "" {
		t.Errorf("Bool(0).String() = %q, want empty", Bool(0).String())
	}
	if Bool(1).String() != "true" {
		t.Errorf("Bool(1).String() = %q, want %q", Bool(1).String(), "true")
	}
	if Bool(2).String() != "false" {
		t.Errorf("Bool(2).String() = %q, want %q", Bool(2).String(), "false")
	}
}

func TestBoolToInt(t *testing.T) {
	if BoolToInt(true) != 1 {
		t.Error("BoolToInt(true) should be 1")
	}
	if BoolToInt(false) != 0 {
		t.Error("BoolToInt(false) should be 0")
	}
}

func TestPairOf(t *testing.T) {
	p := PairOf(1, "a")
	if p.First != 1 || p.Second != "a" {
		t.Errorf("PairOf(1, \"a\") = {%v, %v}, want {1, a}", p.First, p.Second)
	}
}

func TestPair_Val(t *testing.T) {
	p := PairOf(10, 20)
	a, b := p.Val()
	if a != 10 || b != 20 {
		t.Errorf("Pair.Val() = (%v, %v), want (10, 20)", a, b)
	}
}

func TestPairPtrOf(t *testing.T) {
	p := PairPtrOf(1, "a")
	if p.First != 1 || p.Second != "a" {
		t.Errorf("PairPtrOf(1, \"a\") = {%v, %v}, want {1, a}", p.First, p.Second)
	}
}

func TestTupleOf(t *testing.T) {
	tup := TupleOf(1, "a", 3.14)
	if tup.First != 1 || tup.Second != "a" || tup.Third != 3.14 {
		t.Errorf("TupleOf() values incorrect")
	}
}

func TestTuple_Val(t *testing.T) {
	tup := TupleOf(1, 2, 3)
	a, b, c := tup.Val()
	if a != 1 || b != 2 || c != 3 {
		t.Errorf("Tuple.Val() = (%v, %v, %v), want (1, 2, 3)", a, b, c)
	}
}

func TestRef(t *testing.T) {
	val := 42
	r := RefOf(&val)
	v, ok := r.Val()
	if !ok || v != 42 {
		t.Errorf("Ref.Val() = (%v, %v), want (42, true)", v, ok)
	}
}

func TestRef_Get(t *testing.T) {
	val := 10
	r := RefOf(&val)
	if r.Get() != 10 {
		t.Errorf("Ref.Get() = %v, want 10", r.Get())
	}
}

func TestRef_Set(t *testing.T) {
	val := 10
	r := RefOf(&val)
	old := r.Set(20)
	if old != 10 {
		t.Errorf("Ref.Set() old = %v, want 10", old)
	}
	if r.Get() != 20 {
		t.Errorf("Ref.Get() after Set = %v, want 20", r.Get())
	}
}

func TestRef_IsNil(t *testing.T) {
	r := Ref[int]{}
	if !r.IsNil() {
		t.Error("Ref with nil value should be nil")
	}
	val := 1
	r = RefOf(&val)
	if r.IsNil() {
		t.Error("Ref with non-nil value should not be nil")
	}
}

func TestRef_IsNotNil(t *testing.T) {
	r := Ref[int]{}
	if r.IsNotNil() {
		t.Error("Ref with nil value, IsNotNil should be false")
	}
	val := 1
	r = RefOf(&val)
	if !r.IsNotNil() {
		t.Error("Ref with non-nil value, IsNotNil should be true")
	}
}

func TestOption_Some(t *testing.T) {
	opt := Some(42)
	if !opt.IsSome() {
		t.Error("Some(42).IsSome() should be true")
	}
	if opt.IsNone() {
		t.Error("Some(42).IsNone() should be false")
	}
	v, ok := opt.Get()
	if !ok || v != 42 {
		t.Errorf("Some(42).Get() = (%v, %v), want (42, true)", v, ok)
	}
}

func TestOption_None(t *testing.T) {
	opt := None[int]()
	if opt.IsSome() {
		t.Error("None().IsSome() should be false")
	}
	if !opt.IsNone() {
		t.Error("None().IsNone() should be true")
	}
}

func TestOption_Unwrap(t *testing.T) {
	opt := Some(42)
	if opt.Unwrap() != 42 {
		t.Errorf("Some(42).Unwrap() = %v, want 42", opt.Unwrap())
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("None().Unwrap() should panic")
		}
	}()
	opt2 := None[int]()
	opt2.Unwrap()
}

func TestOption_UnwrapOr(t *testing.T) {
	opt := Some(42)
	if opt.UnwrapOr(0) != 42 {
		t.Error("Some(42).UnwrapOr(0) should be 42")
	}
	opt2 := None[int]()
	if opt2.UnwrapOr(99) != 99 {
		t.Error("None().UnwrapOr(99) should be 99")
	}
}

func TestOption_UnwrapOrElse(t *testing.T) {
	opt := Some(42)
	if opt.UnwrapOrElse(func() int { return 0 }) != 42 {
		t.Error("Some(42).UnwrapOrElse() should be 42")
	}
	opt2 := None[int]()
	if opt2.UnwrapOrElse(func() int { return 99 }) != 99 {
		t.Error("None().UnwrapOrElse() should be 99")
	}
}

func TestOption_IfSome(t *testing.T) {
	called := false
	opt := Some(42)
	opt.IfSome(func(v int) {
		called = true
		if v != 42 {
			t.Errorf("IfSome callback got %v, want 42", v)
		}
	})
	if !called {
		t.Error("IfSome callback not called")
	}
	called = false
	opt2 := None[int]()
	opt2.IfSome(func(v int) { called = true })
	if called {
		t.Error("IfSome on None should not call callback")
	}
}

func TestOption_IfNone(t *testing.T) {
	called := false
	opt := None[int]()
	opt.IfNone(func() { called = true })
	if !called {
		t.Error("IfNone callback not called")
	}
	called = false
	opt2 := Some(42)
	opt2.IfNone(func() { called = true })
	if called {
		t.Error("IfNone on Some should not call callback")
	}
}

func TestMapOption(t *testing.T) {
	opt := Some(42)
	mapped := MapOption(opt, func(v int) string {
		return "result"
	})
	if !mapped.IsSome() || mapped.Unwrap() != "result" {
		t.Error("MapOption on Some should return Some")
	}
	mapped2 := MapOption(None[int](), func(v int) string { return "result" })
	if mapped2.IsSome() {
		t.Error("MapOption on None should return None")
	}
}

func TestOk(t *testing.T) {
	r := Ok(42)
	if !r.IsOk() {
		t.Error("Ok(42).IsOk() should be true")
	}
	if r.IsErr() {
		t.Error("Ok(42).IsErr() should be false")
	}
}

func TestErr(t *testing.T) {
	r := Err[int](errors.New("fail"))
	if r.IsOk() {
		t.Error("Err().IsOk() should be false")
	}
	if !r.IsErr() {
		t.Error("Err().IsErr() should be true")
	}
}

func TestResult_Val(t *testing.T) {
	r := Ok(42)
	v, err := r.Val()
	if err != nil || v != 42 {
		t.Errorf("Ok(42).Val() = (%v, %v), want (42, nil)", v, err)
	}
	r2 := Err[int](errors.New("fail"))
	v2, err2 := r2.Val()
	if err2 == nil {
		t.Error("Err().Val() should return error")
	}
	_ = v2
}

func TestResult_OrPanic(t *testing.T) {
	r := Ok(42)
	if r.OrPanic() != 42 {
		t.Errorf("Ok(42).OrPanic() = %v, want 42", r.OrPanic())
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("Err().OrPanic() should panic")
		}
	}()
	Err[int](errors.New("fail")).OrPanic()
}

func TestResult_Or(t *testing.T) {
	if Ok(42).Or(0) != 42 {
		t.Error("Ok(42).Or(0) should be 42")
	}
	if Err[int](errors.New("fail")).Or(99) != 99 {
		t.Error("Err().Or(99) should be 99")
	}
}

func TestResult_OrDefault(t *testing.T) {
	if Ok(42).OrDefault() != 42 {
		t.Error("Ok(42).OrDefault() should be 42")
	}
	if Err[int](errors.New("fail")).OrDefault() != 0 {
		t.Error("Err().OrDefault() should be 0")
	}
}

func TestResult_IsOkAnd(t *testing.T) {
	r := Ok(42)
	if !r.IsOkAnd(func(v int) bool { return v > 0 }) {
		t.Error("Ok(42).IsOkAnd(v > 0) should be true")
	}
	if r.IsOkAnd(func(v int) bool { return v < 0 }) {
		t.Error("Ok(42).IsOkAnd(v < 0) should be false")
	}
	if Err[int](errors.New("fail")).IsOkAnd(func(v int) bool { return true }) {
		t.Error("Err().IsOkAnd() should be false")
	}
}

func TestResult_IfOk(t *testing.T) {
	called := false
	Ok(42).IfOk(func(v int) { called = true })
	if !called {
		t.Error("IfOk callback not called")
	}
	called = false
	Err[int](errors.New("fail")).IfOk(func(v int) { called = true })
	if called {
		t.Error("IfOk on Err should not call callback")
	}
}

func TestResult_IfErr(t *testing.T) {
	called := false
	Err[int](errors.New("fail")).IfErr(func(err error) { called = true })
	if !called {
		t.Error("IfErr callback not called")
	}
	called = false
	Ok(42).IfErr(func(err error) { called = true })
	if called {
		t.Error("IfErr on Ok should not call callback")
	}
}

func TestCastSigned(t *testing.T) {
	if CastSigned[int, int8](42) != 42 {
		t.Error("CastSigned(42) should be 42")
	}
}

func TestCastFloat(t *testing.T) {
	var v float32 = 3.14
	result := CastFloat[float64, float32](v)
	if result != float64(v) {
		t.Error("CastFloat should work")
	}
}

func TestCastUnsigned(t *testing.T) {
	if CastUnsigned[uint, uint8](42) != 42 {
		t.Error("CastUnsigned(42) should be 42")
	}
}

func TestCastInteger(t *testing.T) {
	if CastInteger[int, int32](42) != 42 {
		t.Error("CastInteger(42) should be 42")
	}
}
