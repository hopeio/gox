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
	interval  time.Duration
	timer     *time.Ticker
	handlers  FileWatchInfos
	done      chan struct{}
	closeOnce sync.Once
	mu        sync.Mutex
}

type FileWatchInfo struct {
	req         *http.Request
	lastModTime time.Time
	callback    func(file *FileInfo)
	md5value    [16]byte
}

type FileWatchInfos map[string]*FileWatchInfo

// NewFileWatcher creates and returns a new instance.
func NewFileWatcher(interval time.Duration) *FileWatcher {
	w := &FileWatcher{
		interval: interval,
		handlers: make(map[string]*FileWatchInfo),
		timer:    time.NewTicker(interval),
		done:     make(chan struct{}),
	}

	go w.run()

	return w
}

// Add registers url and fetches it once immediately.
// The map key is the caller-provided url, matching Remove.
func (w *FileWatcher) Add(url string, callback func(file *FileInfo), opts ...func(r *http.Request)) error {
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

	// 首次拉取放在锁外，网络 IO 不阻塞其它 Add/tick
	c.Do()
	w.mu.Lock()
	w.handlers[url] = c
	w.mu.Unlock()
	return nil
}

// Remove removes or resets state.
func (w *FileWatcher) Remove(url string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.handlers, url)
	return nil
}

// run performs the operation.
func (w *FileWatcher) run() {
	for {
		select {
		case <-w.done:
			return
		case <-w.timer.C:
			// 锁内取快照，锁外做网络 IO
			w.mu.Lock()
			infos := make([]*FileWatchInfo, 0, len(w.handlers))
			for _, c := range w.handlers {
				infos = append(infos, c)
			}
			w.mu.Unlock()
			for _, c := range infos {
				c.Do()
			}
		}
	}
}

// Close stops the ticker and terminates the watch goroutine. Safe to call multiple times.
func (w *FileWatcher) Close() {
	w.closeOnce.Do(func() {
		w.timer.Stop()
		close(w.done)
	})
}

// Do executes the operation.
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
		} else {
			// 未变化也必须关闭响应体，否则每个 tick 泄漏一个连接
			file.Body.Close()
		}
		return
	}
	data, err := io.ReadAll(file.Body)
	file.Body.Close()
	if err != nil {
		log.Error(err)
		return
	}
	md5value := md5.Sum(data)
	if md5value != c.md5value {
		c.md5value = md5value
		c.lastModTime = file.ModTime()
		file.Body = io.NopCloser(bytes.NewReader(data))
		c.callback(file)
	}
}

// Update resets the polling interval. It must not spawn another run goroutine.
func (w *FileWatcher) Update(interval time.Duration) {
	w.mu.Lock()
	w.interval = interval
	w.mu.Unlock()
	w.timer.Reset(interval)
}
