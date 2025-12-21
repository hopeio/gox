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
	"github.com/hopeio/gox/net/http/gin/binding"
	"github.com/hopeio/gox/net/http/handlerwrap"
	"github.com/hopeio/gox/types"
)

type Service[REQ, RES any] func(*gin.Context, REQ) (RES, *httpx.ErrResp)

func HandlerWrap[REQ, RES any](service Service[*REQ, *RES]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		err := binding.Bind(ctx, req)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			httpx.RespondError(ctx, ctx.Writer, errors.InvalidArgument.Wrap(err))
			ctx.Abort()
			return
		}
		res, reserr := service(ctx, req)
		if reserr != nil {
			httpx.RespondError(ctx, ctx.Writer, reserr)
			ctx.Abort()
			return
		}
		if httpres, ok := any(res).(httpx.CommonResponder); ok {
			httpres.CommonRespond(ctx, httpx.ResponseWriterWrapper{ResponseWriter: ctx.Writer})
			return
		}
		httpx.RespondSuccess(ctx, ctx.Writer, res)
	}
}

func HandlerWrapGRPC[REQ, RES any](service types.GrpcService[*REQ, *RES]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		err := binding.Bind(ctx, req)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			httpx.RespondError(ctx, ctx.Writer, errors.InvalidArgument.Wrap(err))
			ctx.Abort()
			return
		}
		res, err := service(handlerwrap.WrapContext(ctx), req)
		if err != nil {
			httpx.RespondError(ctx, ctx.Writer, err)
			ctx.Abort()
			return
		}
		if httpres, ok := any(res).(httpx.CommonResponder); ok {
			httpres.CommonRespond(ctx, httpx.ResponseWriterWrapper{ResponseWriter: ctx.Writer})
			return
		}
		httpx.RespondSuccess(ctx, ctx.Writer, res)
	}
}

func Respond(ctx *gin.Context, v any) (int, error) {
	if err, ok := v.(error); ok {
		written, err := httpx.RespondError(ctx, ctx.Writer, err)
		ctx.Abort()
		return written, err
	}
	return httpx.RespondSuccess(ctx, ctx.Writer, v)
}
