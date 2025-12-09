/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package timer

import (
	"os"
	"path/filepath"
	"time"

	"github.com/hopeio/gox/log"
	"github.com/hopeio/gox/os/fs/watch"
)

// only support Create,Remove,Write
type Watch struct {
	interval time.Duration
	handlers watch.Handlers
	timer    *time.Ticker
}

func New(interval time.Duration) (*Watch, error) {
	return &Watch{
		interval: interval,
		handlers: make(watch.Handlers),
		timer:    time.NewTicker(interval),
	}, nil
}

func (w *Watch) Add(path string, op watch.Op, callback func(string)) error {
	path = filepath.Clean(path)
	var modTime time.Time
	info, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		modTime = info.ModTime()
	}

	handler, ok := w.handlers[path]
	if !ok {
		handler = &watch.Callback{
			LastModTime: modTime,
		}
		w.handlers[path] = handler
	}
	handler.Callbacks[op-1] = callback
	return nil
}

func (w *Watch) run() {

	for range w.timer.C {

		for path, handler := range w.handlers {
			var modTime time.Time
			info, err := os.Stat(path)
			if err != nil && !os.IsNotExist(err) {
				log.Error(err)
			}
			if err == nil {
				modTime = info.ModTime()
			}

			if !handler.LastModTime.IsZero() {
				if modTime.IsZero() {
					handler.LastModTime = modTime
					handler.Callbacks[watch.Remove-1](path)
				} else {
					if modTime.Sub(handler.LastModTime) > time.Second {
						handler.LastModTime = modTime
						handler.Callbacks[watch.Write-1](path)
					}
				}
			} else {
				if modTime.After(handler.LastModTime) {
					handler.LastModTime = modTime
					handler.Callbacks[watch.Create-1](path)
				}
			}
		}
	}
}

func (w *Watch) Close() error {
	w.timer.Stop()
	return nil
}
