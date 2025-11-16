/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package errors

import (
	"strconv"
)

type ErrCode uint32

func (x ErrCode) String() string {
	value, ok := codeMsgMap[x]
	if ok {
		return value
	}
	return "Unknown Error, Code:" + strconv.Itoa(int(x))
}

func (x ErrCode) ErrResp() *ErrResp {
	return &ErrResp{Code: x, Msg: x.String()}
}

func (x ErrCode) Msg(msg string) *ErrResp {
	return &ErrResp{Code: x, Msg: msg}
}

func (x ErrCode) Wrap(err error) *ErrResp {
	return &ErrResp{Code: x, Msg: err.Error()}
}

func (x ErrCode) Error() string {
	return x.String()
}
