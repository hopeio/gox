/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/net/http/consts"

	"github.com/hopeio/gox/encoding/protobuf/jsonpb"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/grpc"
	"github.com/hopeio/gox/net/http/grpc/gateway"
	"google.golang.org/protobuf/proto"
)

func ForwardResponseMessage(ctx *gin.Context, md grpc.ServerMetadata, message proto.Message) {
	if res, ok := message.(httpx.ICommonResponseTo); ok {
		res.CommonResponse(httpx.CommonResponseWriter{ctx.Writer})
		return
	}
	gateway.HandleForwardResponseServerMetadata(ctx.Writer, md.HeaderMD)
	gateway.HandleForwardResponseTrailerHeader(ctx.Writer, md.TrailerMD)

	contentType := jsonpb.JsonPb.ContentType(message)
	ctx.Header(consts.HeaderContentType, contentType)

	if !message.ProtoReflect().IsValid() {
		ctx.Writer.Write(httpx.ResponseOk)
		return
	}
	gateway.HandleForwardResponseTrailer(ctx.Writer, md.TrailerMD)
	err := gateway.Response(ctx, ctx.Writer, message)
	if err != nil {
		HttpError(ctx, err)
		return
	}
}
