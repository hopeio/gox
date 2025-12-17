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
	"github.com/hopeio/pick"
)

type Service[REQ, RES any] func(*gin.Context, REQ) (RES, *httpx.ErrResp)

func HandlerWrap[REQ, RES any](service Service[*REQ, *RES]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		err := binding.Bind(ctx, req)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			Respond(ctx, errors.InvalidArgument.Wrap(err))
			return
		}
		res, reserr := service(ctx, req)
		if reserr != nil {
			reserr.Respond(ctx, ctx.Writer)
			return
		}
		if httpres, ok := any(res).(httpx.CommonResponder); ok {
			httpres.CommonRespond(ctx, httpx.ResponseWriterWrapper{ResponseWriter: ctx.Writer})
			return
		}
		httpx.NewSuccessRespData(res).Respond(ctx, ctx.Writer)
	}
}

func HandlerWrapGRPC[REQ, RES any](service types.GrpcService[*REQ, *RES]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		err := binding.Bind(ctx, req)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			Respond(ctx, errors.InvalidArgument.Wrap(err))
			return
		}
		res, err := service(handlerwrap.WrapContext(ctx), req)
		if err != nil {
			httpx.ErrRespFrom(err).Respond(ctx, ctx.Writer)
			return
		}
		if httpres, ok := any(res).(httpx.CommonResponder); ok {
			httpres.CommonRespond(ctx, httpx.ResponseWriterWrapper{ResponseWriter: ctx.Writer})
			return
		}
		httpx.NewSuccessRespData(res).Respond(ctx, ctx.Writer)
	}
}

func Respond(ctx *gin.Context, v any) (int, error) {
	ctx.Header(httpx.HeaderContentType, pick.DefaultMarshaler.ContentType(v))
	data, err := httpx.DefaultMarshaler.Marshal(v)
	if err != nil {
		data = []byte(err.Error())
	}
	n, err := ctx.Writer.Write(data)
	ctx.Abort()
	return n, err
}
