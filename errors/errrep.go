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

type IErrResp interface {
	ErrResp() *ErrResp
}

type ErrResp struct {
	Code ErrCode `json:"code"`
	Msg  string  `json:"msg,omitempty"`
}

func NewErrResp(code ErrCode, msg string) *ErrResp {
	return &ErrResp{
		Code: code,
		Msg:  msg,
	}
}

func (x *ErrResp) Error() string {
	return x.Msg
}

func (x *ErrResp) MarshalJSON() ([]byte, error) {
	return stringsx.ToBytes(`{"code":` + strconv.Itoa(int(x.Code)) + `,"msg":` + strconv.Quote(x.Msg) + `}`), nil
}

func ErrRespFrom(err error) *ErrResp {
	if err == nil {
		return nil
	}
	if errrep, ok := err.(*ErrResp); ok {
		return errrep
	}
	type errrep interface{ ErrRep() *ErrResp }
	if se, ok := err.(errrep); ok {
		return se.ErrRep()
	}
	rv := reflect.ValueOf(err)
	kind := rv.Kind()
	if kind >= reflect.Int && kind <= reflect.Int64 {
		return NewErrResp(ErrCode(rv.Int()), err.Error())
	}
	if kind >= reflect.Uint && kind <= reflect.Uint64 {
		return NewErrResp(ErrCode(rv.Uint()), err.Error())
	}
	return NewErrResp(Unknown, err.Error())
}
