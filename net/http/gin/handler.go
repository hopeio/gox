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
	"github.com/hopeio/gox/net/http/gin/binding"
	"github.com/hopeio/gox/net/http/handlerwrap"
	"github.com/hopeio/gox/types"
	"net/http"
)

// only example

type GinService[REQ, RES any] func(*gin.Context, REQ) (RES, *httpx.ErrRep)

func HandlerWrap[REQ, RES any](service GinService[*REQ, *RES]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		err := binding.Bind(ctx, req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errcode.InvalidArgument.Wrap(err))
			return
		}
		res, reserr := service(ctx, req)
		if reserr != nil {
			reserr.Response(ctx.Writer)
			return
		}
		if httpres, ok := any(res).(httpx.ICommonResponseTo); ok {
			httpres.CommonResponse(httpx.CommonResponseWriter{ctx.Writer})
			return
		}
		httpx.NewSuccessRespData(res).Response(ctx.Writer)
	}
}

func HandlerWrapCompatibleGRPC[REQ, RES any](service types.GrpcService[*REQ, *RES]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		err := binding.Bind(ctx, req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errcode.InvalidArgument.Wrap(err))
			return
		}
		res, err := service(handlerwrap.WarpContext(ctx), req)
		if err != nil {
			httpx.ErrRepFrom(err).Response(ctx.Writer)
			return
		}
		if httpres, ok := any(res).(httpx.ICommonResponseTo); ok {
			httpres.CommonResponse(httpx.CommonResponseWriter{ctx.Writer})
			return
		}
		httpx.NewSuccessRespData(res).Response(ctx.Writer)
	}
}
