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

func HandleDebug(prefix string) {
	http.HandleFunc(prefix+"/debug/stack", HandleStack)
	if prefix != "" && prefix != "GET " {
		http.HandleFunc(prefix+"/debug/pprof/", pprof.Index)
		http.HandleFunc(prefix+"/debug/pprof/cmdline", pprof.Cmdline)
		http.HandleFunc(prefix+"/debug/pprof/profile", pprof.Profile)
		http.HandleFunc(prefix+"/debug/pprof/symbol", pprof.Symbol)
		http.HandleFunc(prefix+"/debug/pprof/trace", pprof.Trace)
		http.Handle(prefix+"/debug/vars", expvar.Handler())
	}
}

func HandleStack(w http.ResponseWriter, r *http.Request) {
	w.Write(debug.Stack())
}
