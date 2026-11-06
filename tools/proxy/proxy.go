package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/rs/cors"
)

// Rewrite returns the result.
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
			Proxy:             http.ProxyFromEnvironment, // use proxy
			ForceAttemptHTTP2: true,
		},
		ModifyResponse: func(resp *http.Response) error {
			delete(resp.Header, "Access-Control-Allow-Origin")
			return nil
		},
	}
	return proxy
}

// Director returns the result.
func Director() *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			targets := r.Header["Target-Url"]
			if len(targets) == 0 {
				return
			}
			target := targets[0]
			targetUrl, err := url.Parse(target)
			if err != nil {
				// 解析失败时保持原请求，避免对 nil URL 解引用
				return
			}
			r.Host = targetUrl.Host
			r.URL.Host = targetUrl.Host
			r.URL.Scheme = targetUrl.Scheme

			r.Header["Refer"] = r.Header["Target-Refer"]
			r.Header["Origin"] = r.Header["Target-Origin"]
		},
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment, // use proxy
			ForceAttemptHTTP2: true,
		},
		ModifyResponse: func(resp *http.Response) error {
			delete(resp.Header, "Access-Control-Allow-Origin")
			return nil
		},
	}
	return proxy
}

// DirectorServer performs the operation.
func DirectorServer(addr string) error {
	server := cors.AllowAll()
	return http.ListenAndServe(addr, server.Handler(Director()))
}

func main() {
	DirectorServer(":8080")
}
