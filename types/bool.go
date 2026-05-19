/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import "errors"

type Bool int8

func (b Bool) IsTrue() bool {
	return b == 1
}

func (b Bool) IsFalse() bool {
	return b == 2
}

func (b Bool) IsNone() bool {
	return b == 0
}

func (b Bool) MarshalJSON() ([]byte, error) {
	switch b {
	case 0:
		return []byte("null"), nil
	case 1:
		return []byte("true"), nil
	case 2:
		return []byte("false"), nil
	}
	return nil, errors.New("invalid bool value")
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bool) UnmarshalJSON(data []byte) error {
	if len(data) <= 5 {
		switch string(data) {
		case "true":
			*b = 1
			return nil
		case "false":
			*b = 2
			return nil
		case "null":
			*b = 0
			return nil
		case "1":
			*b = 1
			return nil
		case "2":
			*b = 2
			return nil
		}
	}

	return errors.New("invalid bool value")
}

// MarshalText
func (b Bool) MarshalText() ([]byte, error) {
	switch b {
	case 0:
		return []byte(""), nil
	case 1:
		return []byte("true"), nil
	case 2:
		return []byte("false"), nil
	}
	return nil, errors.New("invalid bool value")
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *Bool) UnmarshalText(data []byte) error {
	if len(data) <= 5 {
		switch string(data) {
		case "true":
			*b = 1
			return nil
		case "false":
			*b = 2
			return nil
		case "null":
			*b = 0
			return nil
		case "1":
			*b = 1
			return nil
		case "2":
			*b = 2
			return nil
		}
	}
	return errors.New("invalid bool value")
}

// MarshalBinary
func (b Bool) MarshalBinary() ([]byte, error) {
	switch b {
	case 0:
		return []byte{0}, nil
	case 1:
		return []byte{1}, nil
	case 2:
		return []byte{2}, nil
	}
	return nil, errors.New("invalid bool value")
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (b *Bool) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return errors.New("invalid bool value")
	}
	switch data[0] {
	case 1:
		*b = 1
		return nil
	case 2:
		*b = 2
		return nil
	case 0:
		*b = 0
		return nil
	}
	return errors.New("invalid bool value")
}

// String implements Stringer.
func (b Bool) String() string {
	switch b {
	case 0:
		return ""
	case 1:
		return "true"
	case 2:
		return "false"
	}
	return "invalid bool value"
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
