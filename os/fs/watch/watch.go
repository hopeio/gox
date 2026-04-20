/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fsnotify

import (
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Callback struct {
	LastModTime time.Time
	Callbacks   [5]func(string)
}

type Handlers map[string]*Callback

type Watch struct {
	*fsnotify.Watcher
	interval time.Duration
	handlers Handlers
	errHandler func(error)
}

type Option func(*Watch)

func WithWatcher(watcher *fsnotify.Watcher) Option {
	return func(watch *Watch) {
		watch.Watcher = watcher
	}
}

func WithInterval(interval time.Duration) Option {
	return func(watch *Watch) {
		watch.interval = interval
	}
}

func WithErrHandler(errHandler func(error)) Option {
	return func(watch *Watch) {
		watch.errHandler = errHandler
	}
}

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

func (w *Watch) Add(path string, op fsnotify.Op, callback func(string)) error {
	path = filepath.Clean(path)
	handler, ok := w.handlers[path]
	if !ok {
		err := w.Watcher.Add(path)
		if err != nil {
			return err
		}
		handler = &Callback{}
		w.handlers[path] = handler
	}
	handler.Callbacks[op-1] = callback
	return nil
}

func (w *Watch) run() {
	for {
		select {
		case event, ok := <-w.Watcher.Events:
			if !ok {
				return
			}
			if handle, ok := w.handlers[event.Name]; ok {
				now := time.Now()
				if now.Sub(handle.LastModTime) < w.interval  {
					continue
				}
				handle.LastModTime = now
				for i := range handle.Callbacks {
					if event.Op&fsnotify.Op(i+1) == fsnotify.Op(i+1) && handle.Callbacks[i] != nil {
						handle.Callbacks[i](event.Name)
					}
				}
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

func (w *Watch) Close() error {
	close(w.Events)
	close(w.Errors)
	return w.Watcher.Close()
}
