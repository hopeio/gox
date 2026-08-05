/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package bits

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/hopeio/gox/types/constraints"
)

// ViewBin ...
func ViewBin(v any) {
	vv := reflect.ValueOf(v)
	var binary string
	switch v.(type) {
	case int, int8, int16, int32, int64:
		if vv.Int() < 0 {
			f := fmt.Sprintf("%064b", uint64(vv.Int()))
			binary = f[len(f)-int(vv.Type().Size())*8:]
		} else {
			binary = fmt.Sprintf("%0"+strconv.Itoa(int(vv.Type().Size())*8)+"b", v)
		}
	case uint, uint8, uint16, uint32, uint64:
		binary = fmt.Sprintf("%0"+strconv.Itoa(int(vv.Type().Size())*8)+"b", v)
	case float32, float64:
		f := vv.Float()
		ViewBin(*(*int64)(unsafe.Pointer(&f)))
		return
	}
	var out []string
	for i := 0; i < 8; i++ {
		out = append(out, binary[i*8:(i+1)*8])
	}
	fmt.Println(strings.Join(out, " "), " ", v)
}

// ToBytes ...
func ToBytes[T constraints.Number](t T) []byte {
	size := unsafe.Sizeof(t)
	return unsafe.Slice((*byte)(unsafe.Pointer(&t)), size)
}

// FromBytes ...
func FromBytes[T constraints.Number](bytes []byte) T {
	return *(*T)(unsafe.Pointer(unsafe.SliceData(bytes)))
}
