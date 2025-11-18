/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gin

import (
	"github.com/gin-gonic/gin"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/grpc"
	"github.com/hopeio/gox/net/http/grpc/gateway"
	"google.golang.org/protobuf/proto"
)

func ForwardResponseMessage(ctx *gin.Context, md grpc.ServerMetadata, message proto.Message) {
	if res, ok := message.(httpx.CommonResponder); ok {
		res.CommonRespond(ctx, httpx.ResponseWriterWrapper{ResponseWriter: ctx.Writer})
		return
	}
	gateway.HandleForwardResponseServerMetadata(ctx.Writer, md.HeaderMD)
	gateway.HandleForwardResponseTrailerHeader(ctx.Writer, md.TrailerMD)

	contentType := gateway.JsonPb.ContentType(message)
	ctx.Header(httpx.HeaderContentType, contentType)

	if !message.ProtoReflect().IsValid() {
		ctx.Writer.Write(httpx.RespOk)
		return
	}
	gateway.HandleForwardResponseTrailer(ctx.Writer, md.TrailerMD)
	err := gateway.ForwardResponseMessage(ctx.Writer, ctx.Request, message)
	if err != nil {
		HttpError(ctx, err)
		return
	}
}
