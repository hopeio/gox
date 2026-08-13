/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hopeio/gox/container/heap"
	"github.com/hopeio/gox/log"
	"github.com/hopeio/gox/os/fs"
	timex "github.com/hopeio/gox/time"
	"golang.org/x/time/rate"
)

type Key interface {
	uint64 | string | byte | int | int32 | uint32 | int64
}

type Engine[KEY Key] struct {
	workerCount, currentWorkerCount, workingWorkerCount uint64
	waitTaskCount                                       uint64
	workers                                             []*Worker[KEY]
	// workerGroup [][]*Worker[KEY] // TODO: worker groups
	taskChanConsumer chan *Task[KEY]
	errTaskChan      chan *Task[KEY]
	readyTaskHeap    heap.Heap[*Task[KEY]]
	ctx              context.Context
	cancel           context.CancelFunc // manual stop
	wg               sync.WaitGroup     // ensures all tasks finish
	mu               sync.RWMutex
	workersMu        sync.Mutex // protects workers slice; separate from mu to avoid reentrant deadlock
	speedLimit       timex.Ticker
	rateLimiter      *rate.Limiter
	//TODO
	monitorInterval      time.Duration // global monitor interval for stuck tasks and worker panic recovery
	workerFactoryRunning atomic.Bool
	errHandlerRunning    atomic.Bool
	isRunning, isStopped bool
	enableTelemetry      bool
	EngineStatistics
	seen         map[KEY]struct{} // dedup: keys already in the heap, guarded by e.mu
	kindHandlers []*KindHandler[KEY]
	// bounded pending: cap in-flight tasks so readyTaskHeap cannot grow unboundedly
	// backpressure only affects ingest/external callers, never workers
	maxPending uint64
	pendingSem chan struct{}
	submitCh   chan *Task[KEY] // buffer for worker -> ingest subtask submit
	wakeup     chan struct{}   // wake dispatcher when new tasks enter the heap
	errHandler func(task *Task[KEY])
	onStop     []func(context.Context)
	zeroKey    KEY // field kept for performance where generics are not flexible enough
}

type KindHandler[KEY Key] struct {
	Skip        bool
	speedLimit  timex.Ticker
	rateLimiter *rate.Limiter
	// TODO: handlers keyed by Kind
	Handler TaskFunc[KEY]
}

// NewEngine creates and returns a new instance.
func NewEngine[KEY Key](workerCount uint64, opts ...Option[KEY]) *Engine[KEY] {
	engine := &Engine[KEY]{
		workerCount:      workerCount,
		ctx:              context.Background(),
		taskChanConsumer: make(chan *Task[KEY]),
		errTaskChan:      make(chan *Task[KEY], 1024),
		readyTaskHeap:    heap.Heap[*Task[KEY]]{},
		monitorInterval:  5 * time.Second,
		seen:             make(map[KEY]struct{}),
		errHandler:       func(task *Task[KEY]) { task.ErrLog() },
		wakeup:           make(chan struct{}, 1),
	}

	// opts 必须在使用 maxPending 等字段之前应用，否则 WithMaxPending/WithContext 全部失效
	for _, opt := range opts {
		opt(engine)
	}

	ctx, cancel := context.WithCancel(engine.ctx)
	engine.ctx = ctx
	engine.cancel = cancel

	if engine.maxPending > 0 {
		engine.pendingSem = make(chan struct{}, engine.maxPending)
		// Fill quota initially: pendingSem is remaining slots and must start full.
		// Otherwise the first addTasks call blocks forever waiting for a return that never comes.
		for i := uint64(0); i < engine.maxPending; i++ {
			engine.pendingSem <- struct{}{}
		}
		// Submit buffer absorbs bursts so workers are not constantly backpressured
		buf := int(engine.maxPending)
		if buf > 4096 {
			buf = 4096
		}
		if buf < 64 {
			buf = 64
		}
		engine.submitCh = make(chan *Task[KEY], buf)
		go engine.ingest()
	}
	return engine
}

// SkipKind returns the result.
func (e *Engine[KEY]) SkipKind(kinds ...Kind) *Engine[KEY] {
	length := slices.Max(kinds) + 1
	if e.kindHandlers == nil {
		e.kindHandlers = make([]*KindHandler[KEY], length)
	}
	if int(length) > len(e.kindHandlers) {
		e.kindHandlers = append(e.kindHandlers, make([]*KindHandler[KEY], int(length)-len(e.kindHandlers))...)
	}
	for _, kind := range kinds {
		if e.kindHandlers[kind] == nil {
			e.kindHandlers[kind] = &KindHandler[KEY]{Skip: true}
		} else {
			e.kindHandlers[kind].Skip = true
		}

	}
	return e
}

// MonitorInterval performs the operation.
func (e *Engine[KEY]) MonitorInterval(interval time.Duration) {
	if interval < time.Second {
		log.Warn("monitor interval min one second")
		interval = time.Second
	}

	e.monitorInterval = interval
}

// ErrHandler returns the result.
func (e *Engine[KEY]) ErrHandler(errHandler func(task *Task[KEY])) *Engine[KEY] {
	e.errHandler = errHandler
	return e
}

// ErrHandlerUtilSuccess returns the result.
func (e *Engine[KEY]) ErrHandlerUtilSuccess() *Engine[KEY] {
	log.Warn("ErrHandlerUtilSuccess will clear history exec log contains err")
	return e.ErrHandler(func(task *Task[KEY]) {
		task.errTimes = 0
		task.reExecLogs = task.reExecLogs[:0]
		e.AddOptionTasks(task.ctx, task.Priority, task)
	})
}

// ErrHandlerRetryTimes returns the result.
func (e *Engine[KEY]) ErrHandlerRetryTimes(times int) *Engine[KEY] {
	return e.ErrHandler(func(task *Task[KEY]) {
		if task.reExecTimes < times {
			task.errTimes = 0
			task.reExecLogs = task.reExecLogs[:0]
			e.AddOptionTasks(task.ctx, task.Priority, task)
		} else {
			task.ErrLog()
		}

	})
}

// ErrHandlerWriteToFile returns the result.
func (e *Engine[KEY]) ErrHandlerWriteToFile(path string) *Engine[KEY] {
	file, err := fs.Create(path)
	if err != nil {
		panic(err)
	}
	e.OnStop(func(context.Context) {
		file.Close()
	})
	return e.ErrHandler(func(task *Task[KEY]) {
		spew.Fdump(file, task)
	})
}

// OnStop returns the result.
func (e *Engine[KEY]) OnStop(callBack func(context.Context)) *Engine[KEY] {
	e.onStop = append(e.onStop, callBack)
	return e
}

// SpeedLimited returns the result.
func (e *Engine[KEY]) SpeedLimited(interval time.Duration) *Engine[KEY] {
	e.speedLimit = timex.NewTicker(interval)
	return e
}

// RandSpeedLimited returns the result.
func (e *Engine[KEY]) RandSpeedLimited(minInterval, maxInterval time.Duration) *Engine[KEY] {
	e.speedLimit = timex.NewRandTicker(minInterval, maxInterval)
	return e
}

// KindSpeedLimit returns the result.
func (e *Engine[KEY]) KindSpeedLimit(kind Kind, interval time.Duration) *Engine[KEY] {
	limiter := timex.NewRandTicker(interval, interval)
	e.kindSpeedLimit(kind, limiter)
	return e
}

// KindRandSpeedLimit returns the result.
func (e *Engine[KEY]) KindRandSpeedLimit(kind Kind, minInterval, maxInterval time.Duration) *Engine[KEY] {
	limiter := timex.NewRandTicker(minInterval, maxInterval)
	e.kindSpeedLimit(kind, limiter)
	return e
}

// kindSpeedLimit returns the result.
func (e *Engine[KEY]) kindSpeedLimit(kind Kind, limiter timex.Ticker) *Engine[KEY] {
	if e.kindHandlers == nil {
		e.kindHandlers = make([]*KindHandler[KEY], int(kind)+1)
	}
	if int(kind)+1 > len(e.kindHandlers) {
		e.kindHandlers = append(e.kindHandlers, make([]*KindHandler[KEY], int(kind)+1-len(e.kindHandlers))...)
	}
	if e.kindHandlers[kind] == nil {
		e.kindHandlers[kind] = &KindHandler[KEY]{speedLimit: limiter}
	} else {
		e.kindHandlers[kind].speedLimit = limiter
	}
	return e
}

// KindGroupSpeedLimit returns the result.
func (e *Engine[KEY]) KindGroupSpeedLimit(interval time.Duration, kinds ...Kind) *Engine[KEY] {
	limiter := timex.NewRandTicker(interval, interval)
	for _, kind := range kinds {
		e.kindSpeedLimit(kind, limiter)
	}
	return e
}

// KindGroupRandSpeedLimit returns the result.
func (e *Engine[KEY]) KindGroupRandSpeedLimit(minInterval, maxInterval time.Duration, kinds ...Kind) *Engine[KEY] {
	limiter := timex.NewRandTicker(minInterval, maxInterval)
	for _, kind := range kinds {
		e.kindSpeedLimit(kind, limiter)
	}
	return e
}

// Limiter returns the result.
func (e *Engine[KEY]) Limiter(r rate.Limit, b int) *Engine[KEY] {
	e.rateLimiter = rate.NewLimiter(r, b)
	return e
}

// KindLimiter returns the result.
func (e *Engine[KEY]) KindLimiter(kind Kind, r rate.Limit, b int) *Engine[KEY] {
	e.kindLimiter(kind, r, b)
	return e
}

// kindLimiter performs the operation.
func (e *Engine[KEY]) kindLimiter(kind Kind, r rate.Limit, b int) {
	if e.kindHandlers == nil {
		e.kindHandlers = make([]*KindHandler[KEY], int(kind)+1)
	}
	if int(kind)+1 > len(e.kindHandlers) {
		e.kindHandlers = append(e.kindHandlers, make([]*KindHandler[KEY], int(kind)+1-len(e.kindHandlers))...)
	}
	if e.kindHandlers[kind] == nil {
		e.kindHandlers[kind] = &KindHandler[KEY]{rateLimiter: rate.NewLimiter(r, b)}
	} else {
		e.kindHandlers[kind].rateLimiter = rate.NewLimiter(r, b)
	}
}

type AddTask[KEY Key] func(ctx context.Context, priority int, task ...*Task[KEY])

// TaskSource performs the operation.
func (e *Engine[KEY]) TaskSource(taskSource func(addTask *Engine[KEY])) {
	e.wg.Add(1)
	go func() {
		taskSource(e)
		e.wg.Done()
	}()
}

type Option[KEY Key] func(engine *Engine[KEY])

// WithMaxPending updates or inserts a value.
func WithMaxPending[KEY Key](n uint64) Option[KEY] {
	return func(c *Engine[KEY]) { c.maxPending = n }
}

// WithContext updates or inserts a value.
func WithContext[KEY Key](ctx context.Context) Option[KEY] {
	return func(c *Engine[KEY]) { c.ctx = ctx }
}
