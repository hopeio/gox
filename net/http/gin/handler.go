/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/errors"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/types"
)

type Service[REQ, RES any] func(*gin.Context, REQ) (RES, *httpx.ErrResp)

func HandlerWrap[REQ, RES any](service Service[*REQ, *RES]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		err := Bind(ctx, req)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			httpx.RespondError(ctx.Writer, ctx.Request, errors.InvalidArgument.Wrap(err))
			ctx.Abort()
			return
		}
		res, reserr := service(ctx, req)
		if reserr != nil {
			httpx.RespondError(ctx.Writer, ctx.Request, reserr)
			ctx.Abort()
			return
		}
		if httpres, ok := any(res).(http.Handler); ok {
			httpres.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}
		httpx.RespondSuccess(ctx.Writer, ctx.Request, res)
	}
}

func HandlerWrapGRPC[REQ, RES any](service types.GrpcService[*REQ, *RES]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		err := Bind(ctx, req)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			httpx.RespondError(ctx.Writer, ctx.Request, errors.InvalidArgument.Wrap(err))
			ctx.Abort()
			return
		}
		res, err := service(httpx.WrapContext(ctx), req)
		if err != nil {
			httpx.RespondError(ctx.Writer, ctx.Request, err)
			ctx.Abort()
			return
		}
		if httpres, ok := any(res).(http.Handler); ok {
			httpres.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}
		httpx.RespondSuccess(ctx.Writer, ctx.Request, res)
	}
}

func Respond(ctx *gin.Context, v any) {
	if err, ok := v.(error); ok {
		httpx.RespondError(ctx.Writer, ctx.Request, err)
		ctx.Abort()

	}
	httpx.RespondSuccess(ctx.Writer, ctx.Request, v)
}
