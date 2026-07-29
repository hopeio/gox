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
	// workerGroup [][]*Worker[KEY] //TODO 工作组概念
	taskChanConsumer chan *Task[KEY]
	errTaskChan      chan *Task[KEY]
	readyTaskHeap    heap.Heap[*Task[KEY]]
	ctx              context.Context
	cancel           context.CancelFunc // 手动停止执行
	wg               sync.WaitGroup     // 控制确保所有任务执行完
	mu               sync.RWMutex
	workersMu        sync.Mutex // 保护 workers 切片,与 mu 分离避免重入死锁
	speedLimit       timex.Ticker
	rateLimiter      *rate.Limiter
	//TODO
	monitorInterval      time.Duration // 全局检测定时器间隔时间，任务的卡住检测，worker panic recover都可以用这个检测
	workerFactoryRunning atomic.Bool
	errHandlerRunning    atomic.Bool
	isRunning, isStopped bool
	enableTelemetry      bool
	EngineStatistics
	seen         map[KEY]struct{} // 去重:记录已入堆的 key,由 e.mu 保护
	kindHandlers []*KindHandler[KEY]
	// 有界 pending:限制"已加入未结束"的任务总数,防止 readyTaskHeap 无界膨胀
	// 背压只作用在 ingest 协程/外部调用者上,绝不阻塞 worker
	maxPending uint64
	pendingSem chan struct{}
	submitCh   chan *Task[KEY] // worker -> ingest 的子任务提交缓冲
	wakeup     chan struct{}   // 通知 dispatcher 有新任务入堆
	errHandler func(task *Task[KEY])
	onStop     []func(context.Context)
	zeroKey    KEY // 泛型不够强大,又为了性能妥协的字段
}

type KindHandler[KEY Key] struct {
	Skip        bool
	speedLimit  timex.Ticker
	rateLimiter *rate.Limiter
	// TODO 指定Kind的Handler
	Handler TaskFunc[KEY]
}

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

	ctx, cancel := context.WithCancel(engine.ctx)
	engine.ctx = ctx
	engine.cancel = cancel

	if engine.maxPending > 0 {
		engine.pendingSem = make(chan struct{}, engine.maxPending)
		// 初始填满额度:pendingSem 是"剩余可加入名额"的信号量,必须从满开始。
		// 否则 addTasks 的首个任务就会因无人归还额度而永久阻塞。
		for i := uint64(0); i < engine.maxPending; i++ {
			engine.pendingSem <- struct{}{}
		}
		// 提交缓冲足以吸收突发,避免 worker 在递交子任务时频繁被背压
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

func (e *Engine[KEY]) MonitorInterval(interval time.Duration) {
	if interval < time.Second {
		log.Warn("monitor interval min one second")
		interval = time.Second
	}

	e.monitorInterval = interval
}

func (e *Engine[KEY]) ErrHandler(errHandler func(task *Task[KEY])) *Engine[KEY] {
	e.errHandler = errHandler
	return e
}

func (e *Engine[KEY]) ErrHandlerUtilSuccess() *Engine[KEY] {
	log.Warn("ErrHandlerUtilSuccess will clear history exec log contains err")
	return e.ErrHandler(func(task *Task[KEY]) {
		task.errTimes = 0
		task.reExecLogs = task.reExecLogs[:0]
		e.AddOptionTasks(task.ctx, task.Priority, task)
	})
}

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

func (e *Engine[KEY]) OnStop(callBack func(context.Context)) *Engine[KEY] {
	e.onStop = append(e.onStop, callBack)
	return e
}

func (e *Engine[KEY]) SpeedLimited(interval time.Duration) *Engine[KEY] {
	e.speedLimit = timex.NewTicker(interval)
	return e
}

func (e *Engine[KEY]) RandSpeedLimited(minInterval, maxInterval time.Duration) *Engine[KEY] {
	e.speedLimit = timex.NewRandTicker(minInterval, maxInterval)
	return e
}

func (e *Engine[KEY]) KindSpeedLimit(kind Kind, interval time.Duration) *Engine[KEY] {
	limiter := timex.NewRandTicker(interval, interval)
	e.kindSpeedLimit(kind, limiter)
	return e
}

func (e *Engine[KEY]) KindRandSpeedLimit(kind Kind, minInterval, maxInterval time.Duration) *Engine[KEY] {
	limiter := timex.NewRandTicker(minInterval, maxInterval)
	e.kindSpeedLimit(kind, limiter)
	return e
}

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

// 多个kind共用一个timer
func (e *Engine[KEY]) KindGroupSpeedLimit(interval time.Duration, kinds ...Kind) *Engine[KEY] {
	limiter := timex.NewRandTicker(interval, interval)
	for _, kind := range kinds {
		e.kindSpeedLimit(kind, limiter)
	}
	return e
}

func (e *Engine[KEY]) KindGroupRandSpeedLimit(minInterval, maxInterval time.Duration, kinds ...Kind) *Engine[KEY] {
	limiter := timex.NewRandTicker(minInterval, maxInterval)
	for _, kind := range kinds {
		e.kindSpeedLimit(kind, limiter)
	}
	return e
}

func (e *Engine[KEY]) Limiter(r rate.Limit, b int) *Engine[KEY] {
	e.rateLimiter = rate.NewLimiter(r, b)
	return e
}

func (e *Engine[KEY]) KindLimiter(kind Kind, r rate.Limit, b int) *Engine[KEY] {
	e.kindLimiter(kind, r, b)
	return e
}

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

// TaskSource,参数为添加任务的函数，直到该函数运行结束，任务引擎才会检测任务是否结束
func (e *Engine[KEY]) TaskSource(taskSource func(addTask *Engine[KEY])) {
	e.wg.Add(1)
	go func() {
		taskSource(e)
		e.wg.Done()
	}()
}

type Option[KEY Key] func(engine *Engine[KEY])

// WithMaxPending 限制"已加入未结束"的任务总数,防止任务堆无界膨胀。
// 背压只作用在 ingest 协程/外部调用者,不会阻塞 worker。建议 >= WorkerCount。
func WithMaxPending[KEY Key](n uint64) Option[KEY] {
	return func(c *Engine[KEY]) { c.maxPending = n }
}

func WithContext[KEY Key](ctx context.Context) Option[KEY] {
	return func(c *Engine[KEY]) { c.ctx = ctx }
}
