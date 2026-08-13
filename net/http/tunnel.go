/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/hopeio/gox/log"
)

// Tunneling performs the operation.
func Tunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		dest_conn.Close()
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "HTTP/1.1 200 Connection Established\r\n\r\n")
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		// Hijack 失败必须 return：继续走 transfer 会对 nil conn 解引用
		dest_conn.Close()
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}

// transfer performs the operation.
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	if _, err := io.Copy(destination, source); err != nil {
		log.Errorf("transfer failed: %v", err)
	}
}
