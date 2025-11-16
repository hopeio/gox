/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"fmt"

	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/client"
)

type RespData[RES any] httpx.RespData[RES]

func CommonResponse[RES any]() client.ResponseBodyCheck {
	return &RespData[RES]{}
}

func (res *RespData[RES]) CheckError() error {
	if res.Code != 0 {
		return fmt.Errorf("code: %d, msg: %s", res.Code, res.Msg)
	}
	return nil
}

func (res *RespData[RES]) GetData() *RES {
	return &res.Data
}
