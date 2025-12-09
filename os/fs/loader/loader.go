/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package loader

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hopeio/gox/log"
)

type Loader struct {
	// 间隔大于1秒采用timer定时加载，小于1秒用fsnotify
	ReloadInterval time.Duration `comment:"0:not reload; < 1s: fsnotify; >= 1s: polling"`
	Paths          []string
	watcher        *fsnotify.Watcher
	timer          *time.Ticker
	mu             sync.RWMutex
}

type ReloadType int

const (
	ReloadTypeNone ReloadType = iota
	ReloadTypeFsNotify
	ReloadTypeTimer
)

const (
	ReloadTypeNoneName     = "none"
	ReloadTypeFsNotifyName = "fsnotify"
	ReloadTypeTimerName    = "timer"
)

func (t ReloadType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}
func (t *ReloadType) UnmarshalText(data []byte) error {
	switch string(data) {
	case ReloadTypeNoneName:
		*t = ReloadTypeNone
	case ReloadTypeFsNotifyName:
		*t = ReloadTypeFsNotify
	case ReloadTypeTimerName:
		*t = ReloadTypeTimer
	default:
		*t = ReloadTypeNone
	}
	return nil
}

func (t *ReloadType) UnmarshalJSON(data []byte) error {
	data = data[1 : len(data)-1]
	return t.UnmarshalText(data)
}

func (t ReloadType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t ReloadType) String() string {
	switch t {
	case ReloadTypeNone:
		return ReloadTypeNoneName
	case ReloadTypeFsNotify:
		return ReloadTypeFsNotifyName
	case ReloadTypeTimer:
		return ReloadTypeTimerName
	default:
		return ReloadTypeNoneName
	}
}

// New initialize a Loader
func New(interval time.Duration, filepaths ...string) *Loader {
	return &Loader{ReloadInterval: interval, Paths: filepaths}
}

func (ld *Loader) Add(paths ...string) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	if len(paths) == 0 {
		return nil
	}
	ld.Paths = append(ld.Paths, paths...)
	if ld.watcher != nil {
		for _, filepath := range paths {
			err := ld.watcher.Add(filepath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ld *Loader) Remove(paths ...string) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	if len(paths) == 0 {
		return nil
	}
	for i, filepath := range paths {
		if slices.Contains(paths, filepath) {
			if ld.watcher != nil {
				err := ld.watcher.Remove(filepath)
				if err != nil {
					return err
				}
			}
			ld.Paths = append(ld.Paths[:i], paths[i+1:]...)
		}
	}

	return nil
}

func (ld *Loader) Close() error {
	if ld.watcher != nil {
		return ld.watcher.Close()
	}
	if ld.timer != nil {
		ld.timer.Stop()
	}
	return nil
}

// Load will unmarshal configurations to struct from files that you provide
func (ld *Loader) Handler(handler func(io.Reader) error) (err error) {
	if len(ld.Paths) == 0 {
		return errors.New("empty local config path")
	}
	for i, path := range ld.Paths {
		ld.Paths[i], err = filepath.Abs(path)
		if err != nil {
			return err
		}
		err = load(handler, ld.Paths[i])
		if err != nil {
			return err
		}
	}

	if ld.ReloadInterval != 0 {
		if ld.ReloadInterval >= time.Second {
			ld.timer = time.NewTicker(ld.ReloadInterval)
			go ld.watchTimer(handler)
		} else {
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				return err
			}
			for _, filepath := range ld.Paths {
				err = watcher.Add(filepath)
				if err != nil {
					return err
				}
			}
			ld.watcher = watcher
			go ld.watchNotify(handler)
		}
	}

	return
}

func (ld *Loader) watchNotify(handler func(reader io.Reader) error) {
	interval := make(map[string]time.Time)
	for {
		select {
		case event, ok := <-ld.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				now := time.Now()
				if now.Sub(interval[event.Name]) < time.Second {
					continue
				}
				interval[event.Name] = now
				if err := loads(handler, event.Name); err != nil {
					log.Errorf("failed to reload data from %v, got error %v\n", ld.Paths, err)
				}
			}
		case err, ok := <-ld.watcher.Errors:
			if !ok {
				return
			}
			log.Error(err)
		}
	}
}

func (ld *Loader) watchTimer(handle func(reader io.Reader) error) {
	var fileModTimes map[string]time.Time
	for i := range ld.Paths {
		file := ld.Paths[i]

		// check configuration
		if fileInfo, err := os.Stat(file); err == nil && fileInfo.Mode().IsRegular() {
			fileModTimes[file] = fileInfo.ModTime()
		}
	}

	for range ld.timer.C {
		for i := range ld.Paths {
			file := ld.Paths[i]

			// check configuration
			if fileInfo, err := os.Stat(file); err == nil && fileInfo.Mode().IsRegular() {
				if fileInfo.ModTime().After(fileModTimes[file]) {
					fileModTimes[file] = fileInfo.ModTime()
					if err := loads(handle, file); err != nil {
						log.Error("failed to reload data from %v, got error %v\n", ld.Paths, err)
					}
				}
			}
		}
	}
}

func loads(handler func(io.Reader) error, filepaths ...string) (err error) {
	for _, filepath := range filepaths {
		err := load(handler, filepath)
		if err != nil {
			return err
		}
	}
	return err
}

func load(handler func(io.Reader) error, filepath string) (err error) {
	log.Debugf("load data from: '%v'", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	err = handler(file)
	if err != nil {
		return err
	}
	return file.Close()
}
