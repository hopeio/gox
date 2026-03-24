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

type CommonResp[RESP any] httpx.CommonResp[RESP]

func CommonResponse[RESP any]() client.ResponseBodyCheck {
	return &CommonResp[RESP]{}
}

func (res *CommonResp[RESP]) CheckError() error {
	if res.Code != 0 {
		return fmt.Errorf("code: %d, msg: %s", res.Code, res.Msg)
	}
	return nil
}

func (res *CommonResp[RESP]) GetData() *RESP {
	return &res.Data
}
