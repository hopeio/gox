/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"bytes"
	"crypto/md5"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/hopeio/gox/log"
)

type FileWatcher struct {
	interval time.Duration
	timer    *time.Ticker
	handlers FileWatchInfos
	mu       sync.Mutex
}

type FileWatchInfo struct {
	req         *http.Request
	lastModTime time.Time
	callback    func(file *FileInfo)
	md5value    [16]byte
}

type FileWatchInfos map[string]*FileWatchInfo

func NewFileWatcher(interval time.Duration) *FileWatcher {
	w := &FileWatcher{
		interval: interval,
		//1.map和数组做取舍
		handlers: make(map[string]*FileWatchInfo),
		timer:    time.NewTicker(interval),
		//handlers:  make(map[string]map[fsnotify.Operate]func()),
		//2.提高时间复杂度，用event做key，然后每次事件循环取值
		//handlers:  make(map[fsnotify.Event]func()),
	}

	go w.run()

	return w
}

func (w *FileWatcher) Add(url string, callback func(file *FileInfo), opts ...func(r *http.Request)) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	for _, option := range opts {
		option(req)
	}
	c := &FileWatchInfo{
		req:      req,
		callback: callback,
	}

	c.Do()
	w.handlers[req.RequestURI] = c
	return nil
}

func (w *FileWatcher) Remove(url string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.handlers, url)
	return nil
}

func (w *FileWatcher) run() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for range w.timer.C {
		for _, callback := range w.handlers {
			callback.Do()
		}
	}
}

func (w *FileWatcher) Close() {
	w.timer.Stop()
}

func (c *FileWatchInfo) Do() {
	file, err := FetchFileByRequest(c.req)
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

func (w *FileWatcher) Update(interval time.Duration) {
	w.interval = interval
	w.timer.Reset(interval)
	go w.run()
}
