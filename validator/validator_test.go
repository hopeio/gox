/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package validator

import (
	"errors"
	"testing"
)

type testValidate struct {
	Name string `validate:"required" comment:"名称"`
	Age  int    `validate:"min=1" comment:"年龄"`
}

func TestValidateStruct(t *testing.T) {
	// Test with invalid struct
	tv := testValidate{Name: "", Age: 0}
	err := ValidateStruct(tv)
	if err == nil {
		t.Error("ValidateStruct() should return error for invalid struct")
	}
}

func TestValidateStruct_Valid(t *testing.T) {
	tv := testValidate{Name: "test", Age: 10}
	err := ValidateStruct(tv)
	if err != nil {
		t.Errorf("ValidateStruct() error for valid struct: %v", err)
	}
}

type customValidator struct {
	err error
}

func (c customValidator) Validate() error {
	return c.err
}

func TestValidateStruct_CustomValidator(t *testing.T) {
	cv := customValidator{err: errors.New("custom error")}
	err := ValidateStruct(cv)
	if err == nil || err.Error() != "custom error" {
		t.Errorf("ValidateStruct() = %v, want custom error", err)
	}

	cv2 := customValidator{err: nil}
	err = ValidateStruct(cv2)
	if err != nil {
		t.Errorf("ValidateStruct() error: %v", err)
	}
}

type customValidatorAll struct {
	err error
}

func (c customValidatorAll) Validate(all bool) error {
	return c.err
}

func TestValidateStruct_CustomValidatorAll(t *testing.T) {
	cv := customValidatorAll{err: errors.New("all error")}
	err := ValidateStruct(cv)
	if err == nil || err.Error() != "all error" {
		t.Errorf("ValidateStruct() = %v, want all error", err)
	}
}

func TestFieldError(t *testing.T) {
	baseErr := errors.New("base error")
	err := FieldError("field1", baseErr)
	if err.Error() != "invalid field field1: base error" {
		t.Errorf("FieldError() = %v, want 'invalid field field1: base error'", err.Error())
	}

	// Nested FieldError
	err = FieldError("field2", err)
	if err.Error() != "invalid field field2.field1: base error" {
		t.Errorf("Nested FieldError() = %v", err.Error())
	}
}

func TestTransError_Nil(t *testing.T) {
	result := TransError(nil, "zh")
	if result != "" {
		t.Errorf("TransError(nil) = %q, want empty string", result)
	}
}

func TestTransError_NonValidatorError(t *testing.T) {
	err := errors.New("some error")
	result := TransError(err, "zh")
	if result != "some error" {
		t.Errorf("TransError() = %q, want 'some error'", result)
	}
}
