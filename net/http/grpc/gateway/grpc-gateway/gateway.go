/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package grpc_gateway

import (
	"context"
	"net/http"
	"net/url"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/grpc/gateway"
	"google.golang.org/grpc/metadata"
)

type GatewayHandler func(context.Context, *runtime.ServeMux)

func New(opts ...runtime.ServeMuxOption) *runtime.ServeMux {
	opts = append([]runtime.ServeMuxOption{
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &JSONPb{}),
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			area, err := url.PathUnescape(req.Header.Get(httpx.HeaderArea))
			if err != nil {
				area = ""
			}
			var token = httpx.GetToken(req)
			return metadata.MD{
				httpx.HeaderArea:          {area},
				httpx.HeaderDeviceInfo:    {req.Header.Get(httpx.HeaderDeviceInfo)},
				httpx.HeaderLocation:      {req.Header.Get(httpx.HeaderLocation)},
				httpx.HeaderAuthorization: {token},
			}
		}),
		runtime.WithIncomingHeaderMatcher(gateway.InComingHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(gateway.OutgoingHeaderMatcher),
		runtime.WithForwardResponseOption(gateway.Response),
		runtime.WithRoutingErrorHandler(RoutingErrorHandler),
		runtime.WithErrorHandler(CustomHttpError),
	}, opts...)
	return runtime.NewServeMux(opts...)
}
