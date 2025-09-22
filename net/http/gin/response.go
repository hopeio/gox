/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/errors/errcode"
	httpx "github.com/hopeio/gox/net/http"
)

func RespSuccessMsg(ctx *gin.Context, msg string) {
	httpx.RespSuccessMsg(ctx.Writer, msg)
}

func RespErrRep(ctx *gin.Context, rep *errcode.ErrRep) {
	httpx.RespErrRep(ctx.Writer, rep)
}

func Response(ctx *gin.Context, code errcode.ErrCode, msg string, data interface{}) {
	httpx.Response(ctx.Writer, code, msg, data)
}

func RespSuccess[T any](ctx *gin.Context, msg string, data T) {
	httpx.RespSuccess(ctx.Writer, msg, data)
}
