/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package reqctx

import (
	"context"
	"strings"
	"sync"

	contextx "github.com/hopeio/gox/context"
	httpx "github.com/hopeio/gox/net/http"
)

func GetPool[REQ ReqCtx]() sync.Pool {
	return sync.Pool{New: func() any {
		return new(Context[REQ])
	}}
}

type ReqMeta struct {
	Token string
	Auth
	device   *DeviceInfo
	Internal string
	RequestAt
}

type ReqCtx interface {
	RequestContext() context.Context
	RequestHeader() httpx.Header
}

type Context[REQ ReqCtx] struct {
	contextx.Context
	ReqMeta
	ReqCtx REQ
}

func methodFamily(m string) string {
	m = strings.TrimPrefix(m, "/") // remove leading slash
	if i := strings.Index(m, "/"); i >= 0 {
		m = m[:i] // remove everything from second slash
	}
	return m
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
	return &Context[REQ]{
		Context: *contextx.New(ctx),
		ReqMeta: ReqMeta{
			RequestAt: NewRequestAt(),
			Internal:  req.RequestHeader().Get(httpx.HeaderGrpcInternal),
			Token:     GetToken(req),
		},
		ReqCtx: req,
	}
}

func (c *Context[REQ]) Device() *DeviceInfo {
	if c.device == nil {
		header := c.ReqCtx.RequestHeader()
		c.device = Device(header.Get(httpx.HeaderDeviceInfo),
			header.Get(httpx.HeaderArea), header.Get(httpx.HeaderLocation),
			header.Get(httpx.HeaderUserAgent), header.Get(httpx.HeaderXForwardedFor))
	}
	return c.device
}
