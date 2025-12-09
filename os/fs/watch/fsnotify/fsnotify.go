/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fsnotify

import (
	"path/filepath"
	"time"

	"github.com/hopeio/gox/os/fs/watch"

	"github.com/fsnotify/fsnotify"
	"github.com/hopeio/gox/log"
)

type Watch struct {
	*fsnotify.Watcher
	interval time.Duration
	handlers watch.Handlers
}

func New(interval time.Duration) (*Watch, error) {
	watcher, err := fsnotify.NewWatcher()
	w := &Watch{
		Watcher:  watcher,
		interval: interval,
		//1.map和数组做取舍
		handlers: make(watch.Handlers),
		//Handlers:  make(map[string]map[fsnotify.Op]func()),
		//2.提高时间复杂度，用event做key，然后每次事件循环取值
		//Handlers:  make(map[fsnotify.Event]func()),
	}

	if err == nil {
		go w.run()
	}

	return w, err
}

func (w *Watch) Add(path string, op fsnotify.Op, callback func(string)) error {
	path = filepath.Clean(path)
	handler, ok := w.handlers[path]
	if !ok {
		err := w.Watcher.Add(path)
		if err != nil {
			return err
		}
		handler = &watch.Callback{}
		w.handlers[path] = handler
	}
	handler.Callbacks[op-1] = callback
	return nil
}

func (w *Watch) run() {
	ev := &fsnotify.Event{}
	for {
		select {
		case event, ok := <-w.Watcher.Events:
			if !ok {
				return
			}
			ev = &event
			if handle, ok := w.handlers[event.Name]; ok {
				now := time.Now()
				if now.Sub(handle.LastModTime) < w.interval && event == *ev {
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
			log.Error("error:", err)

		}
	}
}

func (w *Watch) Close() error {
	close(w.Events)
	close(w.Errors)
	return w.Watcher.Close()
}
