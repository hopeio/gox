/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fsnotify

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Callback struct {
	LastModTime time.Time
	Callback    func(fsnotify.Event)
}

type Handlers map[string]*Callback

type Watch struct {
	*fsnotify.Watcher
	interval   time.Duration
	mu         sync.Mutex // 保护 handlers：Add 与 run goroutine 并发访问
	handlers   Handlers
	errHandler func(error)
}

type Option func(*Watch)

// WithWatcher updates or inserts a value.
func WithWatcher(watcher *fsnotify.Watcher) Option {
	return func(watch *Watch) {
		watch.Watcher = watcher
	}
}

// WithInterval updates or inserts a value.
func WithInterval(interval time.Duration) Option {
	return func(watch *Watch) {
		watch.interval = interval
	}
}

// WithErrHandler updates or inserts a value.
func WithErrHandler(errHandler func(error)) Option {
	return func(watch *Watch) {
		watch.errHandler = errHandler
	}
}

// New creates a new instance.
func New(opts ...Option) (*Watch, error) {
	watch := &Watch{
		handlers: make(Handlers),
	}
	for _, opt := range opts {
		opt(watch)
	}
	if watch.Watcher == nil {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
		watch.Watcher = watcher
	}
	if watch.interval == 0 {
		watch.interval = time.Second
	}
	go watch.run()
	return watch, nil
}

// Add updates or inserts a value.
func (w *Watch) Add(path string, callback func(fsnotify.Event)) error {
	path = filepath.Clean(path)
	w.mu.Lock()
	defer w.mu.Unlock()
	handler, ok := w.handlers[path]
	if !ok {
		err := w.Watcher.Add(path)
		if err != nil {
			return err
		}
		handler = &Callback{}
		w.handlers[path] = handler
	}
	handler.Callback = callback
	return nil
}

// run performs the operation.
func (w *Watch) run() {
	for {
		select {
		case event, ok := <-w.Watcher.Events:
			if !ok {
				return
			}
			var cb func(fsnotify.Event)
			w.mu.Lock()
			if handle, ok := w.handlers[event.Name]; ok {
				now := time.Now()
				if now.Sub(handle.LastModTime) >= w.interval {
					handle.LastModTime = now
					cb = handle.Callback
				}
			}
			w.mu.Unlock()
			// 回调在锁外执行，避免回调内 Add/Remove 自死锁
			if cb != nil {
				cb(event)
			}
		case err, ok := <-w.Watcher.Errors:
			if !ok {
				return
			}
			if w.errHandler != nil {
				w.errHandler(err)
			}
		}
	}
}

// Close closes and releases resources.
// Events/Errors 归 fsnotify.Watcher 所有，由它的 Close 负责关闭；
// 在这里 close 会与其内部 goroutine 形成 double-close panic。
func (w *Watch) Close() error {
	return w.Watcher.Close()
}
