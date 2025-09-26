/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package errors

import (
	"reflect"
	"strconv"

	stringsx "github.com/hopeio/gox/strings"
)

type IErrRep interface {
	ErrRep() *ErrRep
}

type ErrRep struct {
	Code ErrCode `json:"code"`
	Msg  string  `json:"msg,omitempty"`
}

func NewErrRep(code ErrCode, msg string) *ErrRep {
	return &ErrRep{
		Code: code,
		Msg:  msg,
	}
}

func (x *ErrRep) Error() string {
	return x.Msg
}

func (x *ErrRep) MarshalJSON() ([]byte, error) {
	return stringsx.ToBytes(`{"code":` + strconv.Itoa(int(x.Code)) + `,"msg":` + strconv.Quote(x.Msg) + `}`), nil
}

func ErrRepFrom(err error) *ErrRep {
	if err == nil {
		return nil
	}
	if errrep, ok := err.(*ErrRep); ok {
		return errrep
	}
	type errrep interface{ ErrRep() *ErrRep }
	if se, ok := err.(errrep); ok {
		return se.ErrRep()
	}
	rv := reflect.ValueOf(err)
	kind := rv.Kind()
	if kind >= reflect.Int && kind <= reflect.Int64 {
		return NewErrRep(ErrCode(rv.Int()), err.Error())
	}
	if kind >= reflect.Uint && kind <= reflect.Uint64 {
		return NewErrRep(ErrCode(rv.Uint()), err.Error())
	}
	return NewErrRep(Unknown, err.Error())
}
