/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package ginctx

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/context/reqctx"
	httpx "github.com/hopeio/gox/net/http"
)

type RequestCtx struct {
	*gin.Context
}

func (ctx RequestCtx) RequestHeader() httpx.Header {
	return httpx.HttpHeader(ctx.Request.Header)
}

func (ctx RequestCtx) RequestContext() context.Context {
	return ctx.Request.Context()
}

func (ctx RequestCtx) Origin() *gin.Context {
	return ctx.Context
}

type Context = reqctx.Context[RequestCtx]

func FromRequest(req *gin.Context) *Context {
	return reqctx.New[RequestCtx](RequestCtx{req})
}
