/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	DefaultExcludedExtensions = NewExcludedExtensions([]string{
		".png", ".gif", ".jpeg", ".jpg",
	})
	DefaultGzipOptions = &GzipOptions{
		ExcludedExtensions: DefaultExcludedExtensions,
	}
)

const (
	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

type GzipOptions struct {
	ExcludedExtensions ExcludedExtensions
	ExcludedPaths      ExcludedPaths
	ExcludedPathsRegex ExcludedPathsRegex
	Handler            http.HandlerFunc
}

type gzipWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
	size   int
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del("Content-Length")
	n, err := g.writer.Write(data)
	if err != nil {
		return n, err
	}
	g.size += n
	return n, nil
}

func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

type gzipHandler struct {
	*GzipOptions
	gzPool *sync.Pool
}

func NewGzipHandler(level int, options *GzipOptions) *gzipHandler {
	var gzPool sync.Pool
	gzPool.New = func() interface{} {
		gz, err := gzip.NewWriterLevel(ioutil.Discard, level)
		if err != nil {
			panic(err)
		}
		return gz
	}
	if options == nil {
		options = DefaultGzipOptions
	}
	handler := &gzipHandler{
		GzipOptions: options,
		gzPool:      &gzPool,
	}
	return handler
}

func (g *gzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if !g.shouldCompress(r) {
		return
	}

	gz := g.gzPool.Get().(*gzip.Writer)
	defer g.gzPool.Put(gz)
	defer gz.Reset(ioutil.Discard)
	gz.Reset(w)
	header := w.Header()
	header.Set("Content-Encoding", "gzip")
	header.Set("Vary", "Accept-Encoding")
	gw := &gzipWriter{w, gz, 0}
	g.GzipOptions.Handler(w, r)
	gz.Close()
	header.Set("Content-Length", strconv.Itoa(gw.size))

}

func (g *gzipHandler) shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Content-Type"), "text/event-stream") {

		return false
	}

	extension := filepath.Ext(req.URL.Path)
	if g.ExcludedExtensions.Contains(extension) {
		return false
	}

	if g.ExcludedPaths.Contains(req.URL.Path) {
		return false
	}
	if g.ExcludedPathsRegex.Contains(req.URL.Path) {
		return false
	}

	return true
}

func GzipBody(r *http.Request) (io.ReadCloser, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader %w", err)
		}
		return reader, nil
	}
	return r.Body, nil
}
