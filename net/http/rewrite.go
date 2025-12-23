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

func Rewrite() *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			targets := r.In.Header["Target-Url"]
			if len(targets) == 0 {
				return
			}
			target := targets[0]
			targetUrl, _ := url.Parse(target)
			r.Out.URL = r.In.URL
			r.Out.Host = targetUrl.Host
			r.Out.URL.Host = targetUrl.Host
			r.Out.URL.Scheme = targetUrl.Scheme

			r.Out.Header["Refer"] = r.In.Header["Target-Refer"]
			r.Out.Header["Origin"] = r.In.Header["Target-Origin"]
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

func RewriteServer(addr string) error {
	server := cors.AllowAll()
	return http.ListenAndServe(addr, server.Handler(Rewrite()))
}
