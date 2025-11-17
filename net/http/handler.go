/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"net/http"
)

type Handlers []http.Handler

func (hs Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, handler := range hs {
		handler.ServeHTTP(w, r)
	}
}

type HandlerFuncs []http.HandlerFunc

func (hs HandlerFuncs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, handler := range hs {
		handler(w, r)
	}
}

type Middleware func(http.Handler) http.Handler

// UseMiddleware applies middlewares to a http.HandlerFunc
func UseMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
