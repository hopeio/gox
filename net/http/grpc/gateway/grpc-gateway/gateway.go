/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package grpc_gateway

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/consts"
	"github.com/hopeio/gox/net/http/grpc/gateway"
	"google.golang.org/grpc/metadata"
	"net/http"
	"net/url"
)

type GatewayHandler func(context.Context, *runtime.ServeMux)

func New(opts ...runtime.ServeMuxOption) *runtime.ServeMux {
	opts = append([]runtime.ServeMuxOption{
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &JSONPb{}),
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			area, err := url.PathUnescape(req.Header.Get(consts.HeaderArea))
			if err != nil {
				area = ""
			}
			var token = httpx.GetToken(req)
			return metadata.MD{
				consts.HeaderArea:          {area},
				consts.HeaderDeviceInfo:    {req.Header.Get(consts.HeaderDeviceInfo)},
				consts.HeaderLocation:      {req.Header.Get(consts.HeaderLocation)},
				consts.HeaderAuthorization: {token},
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
