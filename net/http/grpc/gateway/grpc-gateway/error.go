/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package grpc_gateway

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hopeio/gox/errors"
	http2 "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/grpc/gateway"
	stringsx "github.com/hopeio/gox/strings"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

func RoutingErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, httpStatus int) {
	w.WriteHeader(httpStatus)
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(stringsx.ToBytes(http.StatusText(httpStatus)))
}

func CustomHttpError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {

	s, ok := status.FromError(err)
	const fallback = `{"code": 14, "msg": "failed to marshal error message"}`

	w.Header().Del(http2.HeaderTrailer)
	w.Header().Set(http2.HeaderContentType, marshaler.ContentType(nil))
	se := &errors.ErrRep{Code: errors.ErrCode(s.Code()), Msg: s.Message()}
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("Failed to extract ServerMetadata from context")
	}

	gateway.HandleForwardResponseServerMetadata(w, md.HeaderMD)

	buf, merr := marshaler.Marshal(se)
	if merr != nil {
		grpclog.Infof("Failed to marshal error message %q: %v", se, merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			grpclog.Infof("Failed to write response: %v", err)
		}
		return
	}

	var wantsTrailers bool

	if te := r.Header.Get(http2.HeaderTE); strings.Contains(strings.ToLower(te), "trailers") {
		wantsTrailers = true
		gateway.HandleForwardResponseTrailerHeader(w, md.TrailerMD)
		w.Header().Set(http2.HeaderTransferEncoding, "chunked")
	}

	/*	st := HTTPStatusFromCode(se.Code)
		w.WriteHeader(st)*/
	if _, err := w.Write(buf); err != nil {
		grpclog.Infof("Failed to write response: %v", err)
	}
	if wantsTrailers {
		gateway.HandleForwardResponseTrailer(w, md.TrailerMD)
	}
}
