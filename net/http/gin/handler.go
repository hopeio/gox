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
			ctx.JSON(http.StatusBadRequest, errors.InvalidArgument.Wrap(err))
			return
		}
		res, reserr := service(ctx, req)
		if reserr != nil {
			reserr.Respond(ctx.Writer)
			return
		}
		if httpres, ok := any(res).(httpx.ICommonRespond); ok {
			httpres.CommonRespond(httpx.CommonResponseWriter{ResponseWriter: ctx.Writer})
			return
		}
		httpx.NewSuccessRespData(res).Respond(ctx.Writer)
	}
}

func HandlerWrapCompatibleGRPC[REQ, RES any](service types.GrpcService[*REQ, *RES]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		err := binding.Bind(ctx, req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errors.InvalidArgument.Wrap(err))
			return
		}
		res, err := service(handlerwrap.WarpContext(ctx), req)
		if err != nil {
			httpx.ErrRespFrom(err).Respond(ctx.Writer)
			return
		}
		if httpres, ok := any(res).(httpx.ICommonRespond); ok {
			httpres.CommonRespond(httpx.CommonResponseWriter{ResponseWriter: ctx.Writer})
			return
		}
		httpx.NewSuccessRespData(res).Respond(ctx.Writer)
	}
}
