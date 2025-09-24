/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package engine

import (
	"context"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/hopeio/gox/container/heap"
	"time"
)

type Config[KEY Key] struct {
	WorkerCount     uint64
	MonitorInterval time.Duration // 全局检测定时器间隔时间，任务的卡住检测，worker panic recover都可以用这个检测
	DoneCache       ristretto.Config[KEY, struct{}]
	EnableTelemetry bool
}

func (c *Config[KEY]) NewEngine() *Engine[KEY] {
	return c.NewEngineWithContext(context.Background())
}

func (c *Config[KEY]) NewEngineWithContext(ctx context.Context) *Engine[KEY] {
	c.Init()
	ctx, cancel := context.WithCancel(ctx)
	cache, _ := ristretto.NewCache(&ristretto.Config[KEY, struct{}]{
		NumCounters:        1e4,   // number of keys to track frequency of (10M).
		MaxCost:            1e3,   // maximum cost of cache (MaxCost * 1MB).
		BufferItems:        64,    // number of keys per Get buffer.
		Metrics:            false, // number of keys per Get buffer.
		IgnoreInternalCost: true,
	})
	engine := &Engine[KEY]{
		workerCount:      c.WorkerCount,
		ctx:              ctx,
		cancel:           cancel,
		taskChanConsumer: make(chan *Task[KEY]),
		errTaskChan:      make(chan *Task[KEY]),
		readyTaskHeap:    heap.Heap[*Task[KEY]]{},
		monitorInterval:  c.MonitorInterval,
		done:             cache,
		errHandler:       func(task *Task[KEY]) { task.ErrLog() },
	}
	return engine
}

func (c *Config[KEY]) Init() {
	if c.WorkerCount == 0 {
		c.WorkerCount = 10
	}
	if c.MonitorInterval == 0 {
		c.MonitorInterval = 5 * time.Second
	}
	if c.DoneCache.NumCounters == 0 {
		c.DoneCache.NumCounters = 1e4
	}
	if c.DoneCache.MaxCost == 0 {
		c.DoneCache.MaxCost = 1e3
	}
	if c.DoneCache.BufferItems == 0 {
		c.DoneCache.BufferItems = 64
	}

}

func NewConfig[KEY Key](opts ...Option[KEY]) *Config[KEY] {
	c := &Config[KEY]{
		WorkerCount:     10,
		MonitorInterval: 5 * time.Second,
		DoneCache: ristretto.Config[KEY, struct{}]{
			NumCounters:        1e4,   // number of keys to track frequency of (10M).
			MaxCost:            1e3,   // maximum cost of cache (MaxCost * 1MB).
			BufferItems:        64,    // number of keys per Get buffer.
			Metrics:            false, // number of keys per Get buffer.
			IgnoreInternalCost: true,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option[KEY Key] func(engine *Config[KEY])
