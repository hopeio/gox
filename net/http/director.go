/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/rs/cors"
)

func Director() *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			targets := r.Header["Target-Url"]
			if len(targets) == 0 {
				return
			}
			target := targets[0]
			targetUrl, _ := url.Parse(target)
			r.Host = targetUrl.Host
			r.URL.Host = targetUrl.Host
			r.URL.Scheme = targetUrl.Scheme

			r.Header["Refer"] = r.Header["Target-Refer"]
			r.Header["Origin"] = r.Header["Target-Origin"]
		},
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment, // 代理使用
			ForceAttemptHTTP2: true,
		},
		ModifyResponse: func(resp *http.Response) error {
			delete(resp.Header, "Access-Control-Allow-Origin")
			return nil
		},
	}
	return proxy
}

func DirectorServer(addr string) error {
	server := cors.AllowAll()
	return http.ListenAndServe(addr, server.Handler(Director()))
}
