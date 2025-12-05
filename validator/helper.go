/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package validator

import "strings"

// Validator is a general interface that allows a message to be validated.
type Validator interface {
	Validate() error
}

type ValidatorAll interface {
	Validate(all bool) error
}

func ValidateStruct(o any) error {
	switch validator := o.(type) {
	case ValidatorAll:
		return validator.Validate(true)
	case Validator:
		return validator.Validate()
	}
	return DefaultValidate.Struct(o)
}

type fieldError struct {
	fieldStack []string
	nestedErr  error
}

func (f *fieldError) Error() string {
	return "invalid field " + strings.Join(f.fieldStack, ".") + ": " + f.nestedErr.Error()
}

// FieldError wraps a given Validator error providing a message call stack.
func FieldError(fieldName string, err error) error {
	if fErr, ok := err.(*fieldError); ok {
		fErr.fieldStack = append([]string{fieldName}, fErr.fieldStack...)
		return err
	}
	return &fieldError{
		fieldStack: []string{fieldName},
		nestedErr:  err,
	}
}
