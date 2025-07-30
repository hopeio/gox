/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package watch

import (
	"bytes"
	"crypto/md5"
	"github.com/hopeio/gox/log"
	http_fs "github.com/hopeio/gox/net/http/fs"
	"io"
	"net/http"
	"time"
)

type Watch struct {
	interval time.Duration
	timer    *time.Ticker
	handler  Handler
}

type Callback struct {
	req         *http.Request
	lastModTime time.Time
	callback    func(file *http_fs.FileInfo)
	md5value    [16]byte
}

type Handler map[string]*Callback

func New(interval time.Duration) *Watch {
	w := &Watch{
		interval: interval,
		//1.map和数组做取舍
		handler: make(map[string]*Callback),
		timer:   time.NewTicker(interval),
		//handler:  make(map[string]map[fsnotify.Op]func()),
		//2.提高时间复杂度，用event做key，然后每次事件循环取值
		//handler:  make(map[fsnotify.Event]func()),
	}

	go w.run()

	return w
}

func (w *Watch) Add(url string, callback func(file *http_fs.FileInfo), opts ...func(r *http.Request)) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	for _, option := range opts {
		option(req)
	}
	c := &Callback{
		req:      req,
		callback: callback,
	}

	c.Do()
	w.handler[req.RequestURI] = c
	return nil
}

func (w *Watch) Remove(url string) error {
	delete(w.handler, url)
	return nil
}

func (w *Watch) run() {

	for range w.timer.C {
		for _, callback := range w.handler {
			callback.Do()
		}
	}
}

func (w *Watch) Close() {
	w.timer.Stop()
}

func (c *Callback) Do() {
	file, err := http_fs.FetchFileByRequest(c.req)
	if err != nil {
		log.Error(err)
		return
	}
	if !file.ModTime().IsZero() {
		if file.ModTime().After(c.lastModTime) {
			c.lastModTime = file.ModTime()
			c.callback(file)
		}
		return
	}
	data, err := io.ReadAll(file.Body)
	if err != nil {
		log.Error(err)
		return
	}
	file.Body.Close()
	md5value := md5.Sum(data)
	if md5value != c.md5value {
		c.md5value = md5value
		c.lastModTime = file.ModTime()
		file.Body = io.NopCloser(bytes.NewReader(data))
		c.callback(file)
	}
}

func (w *Watch) Update(interval time.Duration) {
	w.interval = interval
	w.timer.Reset(interval)
	go w.run()
}
