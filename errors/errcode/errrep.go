/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package errcode

import (
	stringsx "github.com/hopeio/gox/strings"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
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

func (x *ErrRep) GRPCStatus() *status.Status {
	return status.New(codes.Code(x.Code), x.Msg)
}

func (x *ErrRep) MarshalJSON() ([]byte, error) {
	return stringsx.ToBytes(`{"code":` + strconv.Itoa(int(x.Code)) + `,"msg":` + strconv.Quote(x.Msg) + `}`), nil
}

func (x *ErrRep) AppendErr(err error) *ErrRep {
	x.Msg += " " + err.Error()
	return x
}

func (x *ErrRep) Wrap(err error) *WrapErrRep {
	return &WrapErrRep{*x, err}
}

func ErrRepFrom(err error) *ErrRep {
	if err == nil {
		return nil
	}
	type errrep interface{ ErrRep() *ErrRep }
	if se, ok := err.(errrep); ok {
		return se.ErrRep()
	}
	return NewErrRep(Unknown, err.Error())
}
