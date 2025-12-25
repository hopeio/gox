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
)

type RequestCtx struct {
	*http.Request
	http.ResponseWriter
}

func (ctx RequestCtx) RequestHeader() http.Header {
	return ctx.Request.Header
}

func (ctx RequestCtx) RequestContext() context.Context {
	return ctx.Request.Context()
}

type Context = reqctx.Context[RequestCtx]

func FromContext(ctx context.Context) (*Context, bool) {
	return reqctx.FromContext[RequestCtx](ctx)
}

func FromRequest(w http.ResponseWriter, r *http.Request) *Context {
	return reqctx.New[RequestCtx](RequestCtx{Request: r, ResponseWriter: w})
}
