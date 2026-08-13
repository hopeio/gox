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

// UseMiddleware returns the result.
func UseMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, mw := range middlewares {
		newHandler := mw(handler)
		if newHandler == nil {
			break
		}
		handler = newHandler
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

// NewMiddlewareContext creates and returns a new instance.
func NewMiddlewareContext(handler http.Handler, handlers ...MiddlewareContextHandler) *MiddlewareContext {
	return &MiddlewareContext{
		handler:  handler,
		handlers: handlers,
	}
}

// Use performs the operation.
func (m *MiddlewareContext) Use(mw MiddlewareContextHandler) {
	m.handlers = append(m.handlers, mw)
}

// Next performs the operation.
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

// ServeHTTP executes the operation.
func (m *MiddlewareContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(m.handlers) == 0 {
		return
	}
	m.Request = r
	m.ResponseWriter = w
	m.handlers[0](m)
	m.index = 0
}
