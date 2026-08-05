/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package decimal

import (
	"fmt"
)

type DecimalModel struct {
	form        byte
	negative    bool
	coefficient []byte
	exponent    int32
}

// String returns the string representation.
func (d *DecimalModel) String() string {
	if len(d.coefficient) == 0 {
		return "0"
	}

	var buf []byte
	switch {
	case d.exponent <= 0:
		// 0.00ddd
		buf = append(buf, "0."...)
		buf = appendZeros(buf, -d.exponent)
		buf = append(buf, d.coefficient...)

	case /* 0 < */ int(d.exponent) < len(d.coefficient):
		// dd.ddd
		buf = append(buf, d.coefficient[:d.exponent]...)
		buf = append(buf, '.')
		buf = append(buf, d.coefficient[d.exponent:]...)

	default: // len(x.mant) <= x.exp
		// ddd00
		buf = append(buf, d.coefficient...)
		buf = appendZeros(buf, d.exponent-int32(len(d.coefficient)))
	}

	return string(buf)
}

// appendZeros returns the result.
func appendZeros(buf []byte, n int32) []byte {
	for ; n > 0; n-- {
		buf = append(buf, '0')
	}
	return buf
}

// Decompose performs the operation.
func (d *DecimalModel) Decompose(buf []byte) (form byte, negative bool, coefficient []byte, exponent int32) {
	// TODO:
	return d.form, d.negative, d.coefficient, d.exponent
}

// Compose performs the operation.
func (d *DecimalModel) Compose(form byte, negative bool, coefficient []byte, exponent int32) error {
	switch form {
	default:
		return fmt.Errorf("unknown form %d", form)
	case 1, 2:
		d.form = form
		d.negative = negative
		return nil
	case 0:
	}
	d.form = form
	d.negative = negative
	d.exponent = exponent

	d.coefficient = coefficient

	return nil
}
