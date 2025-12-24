/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package reqctx

import (
	"context"
	"net/http"
	"sync"

	contextx "github.com/hopeio/gox/context"
	httpx "github.com/hopeio/gox/net/http"
)

var pool *sync.Pool

func getPool[REQ ReqCtx]() *sync.Pool {
	return &sync.Pool{New: func() any {
		return new(Context[REQ])
	}}
}

type Metadata struct {
	RequestTime
	Token string
	Auth
	device   *DeviceInfo
	Internal string
}

type ReqCtx interface {
	RequestContext() context.Context
	RequestHeader() http.Header
}

type Context[REQ ReqCtx] struct {
	*contextx.Context
	Metadata
	ReqCtx REQ
}

func (c *Context[REQ]) Wrapper() context.Context {
	return context.WithValue(c.Context.Base(), contextx.WrapperKey(), c)
}

func FromContext[REQ ReqCtx](ctx context.Context) (*Context[REQ], bool) {
	if ctx == nil {
		return nil, false
	}
	ctxi := ctx.Value(contextx.WrapperKey())
	c, ok := ctxi.(*Context[REQ])
	if !ok {
		return nil, false
	}
	c.SetBase(ctx)
	return c, ok
}

func New[REQ ReqCtx](req REQ) *Context[REQ] {
	ctx := req.RequestContext()
	c, ok := FromContext[REQ](ctx)
	if ok {
		return c
	}
	if pool == nil {
		pool = getPool[REQ]()
	}

	c, ok = pool.Get().(*Context[REQ])
	if ok {
		c.ReqCtx = req
		c.Metadata.RequestTime = NewRequestAt()
		c.Metadata.Internal = req.RequestHeader().Get(httpx.HeaderInternal)
		c.Metadata.Token = GetToken(req)
		c.Context = contextx.New(ctx)
		return c
	}
	return &Context[REQ]{
		Context: contextx.New(ctx),
		Metadata: Metadata{
			RequestTime: NewRequestAt(),
			Internal:    req.RequestHeader().Get(httpx.HeaderInternal),
			Token:       GetToken(req),
		},
		ReqCtx: req,
	}
}

func (c *Context[REQ]) Device() *DeviceInfo {
	if c.device == nil {
		header := c.ReqCtx.RequestHeader()
		c.device = Device(header.Get(httpx.HeaderDeviceInfo), header.Get(httpx.HeaderAppInfo),
			header.Get(httpx.HeaderArea), header.Get(httpx.HeaderLocation),
			header.Get(httpx.HeaderUserAgent), header.Get(httpx.HeaderXForwardedFor))
	}
	return c.device
}
