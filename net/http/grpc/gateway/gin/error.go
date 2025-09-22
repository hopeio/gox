/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/encoding/protobuf/jsonpb"
	"github.com/hopeio/gox/errors/errcode"
	httpx "github.com/hopeio/gox/net/http/consts"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"io"
	"net/http"
)

func HttpError(ctx *gin.Context, err error) {
	s, _ := status.FromError(err)
	const fallback = `{"code": 14, "message": "failed to marshal error message"}`

	delete(ctx.Request.Header, httpx.HeaderTrailer)
	ctx.Header(httpx.HeaderContentType, jsonpb.JsonPb.ContentType(nil))

	se := &errcode.ErrRep{Code: errcode.ErrCode(s.Code()), Msg: s.Message()}
	buf, merr := jsonpb.JsonPb.Marshal(se)
	if merr != nil {
		grpclog.Infof("Failed to marshal error message %q: %v", se, merr)
		ctx.Status(http.StatusInternalServerError)
		if _, err := io.WriteString(ctx.Writer, fallback); err != nil {
			grpclog.Infof("Failed to write response: %v", err)
		}
		return
	}

	if _, err := ctx.Writer.Write(buf); err != nil {
		grpclog.Infof("Failed to write response: %v", err)
	}

}
