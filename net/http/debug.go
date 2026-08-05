/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
)

// HandleDebug ...
func HandleDebug(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/debug/stack", HandleStack)
	if prefix != "" && prefix != "GET " {
		mux.HandleFunc(prefix+"/debug/pprof/", pprof.Index)
		mux.HandleFunc(prefix+"/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc(prefix+"/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc(prefix+"/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc(prefix+"/debug/pprof/trace", pprof.Trace)
		mux.Handle(prefix+"/debug/vars", expvar.Handler())
	}
}

// HandleStack ...
func HandleStack(w http.ResponseWriter, r *http.Request) {
	w.Write(debug.Stack())
}
