/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package httpctx

import (
	"context"
	"net/http"

	"github.com/hopeio/gox/context/reqctx"
	httpx "github.com/hopeio/gox/net/http"
)

type RequestCtx struct {
	*http.Request
	http.ResponseWriter
}

func (ctx RequestCtx) RequestHeader() httpx.Header {
	return httpx.HttpHeader(ctx.Request.Header)
}

func (ctx RequestCtx) RequestContext() context.Context {
	return ctx.Request.Context()
}

type Context = reqctx.Context[RequestCtx]

func FromContext(ctx context.Context) (*Context, bool) {
	return reqctx.FromContext[RequestCtx](ctx)
}

func FromRequest(req RequestCtx) *Context {
	return reqctx.New[RequestCtx](req)
}
