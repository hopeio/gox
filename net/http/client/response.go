/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"fmt"
	httpx "github.com/hopeio/gox/net/http"
)

type ResponseBody httpx.RespAnyData

func CommonResponse(response interface{}) ResponseBodyCheck {
	return &ResponseBody{Data: response}
}

func (res *ResponseBody) CheckError() error {
	if res.Code != 0 {
		return fmt.Errorf("code: %d, msg: %s", res.Code, res.Msg)
	}
	return nil
}

type ResponseBody2 struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   any    `json:"data"`
}

func CommonResponse2(response interface{}) ResponseBodyCheck {
	return &ResponseBody2{Data: response}
}

func (res *ResponseBody2) CheckError() error {
	if res.Status != 0 {
		return fmt.Errorf("status:%d,msg:%s", res.Status, res.Msg)
	}
	return nil
}

var (
	ErrNotFound            = fmt.Errorf("not found")
	ErrRangeNotSatisfiable = fmt.Errorf("range not satisfiable")
)
