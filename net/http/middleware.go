/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func UseMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, mw := range middlewares {
		handler = mw(handler)
	}
	return handler
}

type MiddlewareContextHandler func(*MiddlewareContext)

type MiddlewareContext struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	handler        http.Handler
	handlers       []MiddlewareContextHandler
	index          int
}

func NewMiddlewareContext(handler http.Handler, handlers ...MiddlewareContextHandler) *MiddlewareContext {
	return &MiddlewareContext{
		handler:  handler,
		handlers: handlers,
	}
}

func (m *MiddlewareContext) Use(mw MiddlewareContextHandler) {
	m.handlers = append(m.handlers, mw)
}

func (m *MiddlewareContext) Next() {
	if m.index >= len(m.handlers) {
		return
	}
	if m.index == len(m.handlers)-1 {
		m.index++
		m.handler.ServeHTTP(m.ResponseWriter, m.Request)
		return
	}
	m.index++
	m.handlers[m.index](m)
}

func (m *MiddlewareContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Request = r
	m.ResponseWriter = w
	m.handlers[0](m)
	m.index = 0
}
