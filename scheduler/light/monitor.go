/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package lightengine

import (
	"context"
	"sync/atomic"
)

type Engine struct {
	ctx    context.Context
	cancel context.CancelFunc
	num    atomic.Int32
	ran    atomic.Bool
	onStop func()
}

func New(ctx context.Context) *Engine {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	return &Engine{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (ng *Engine) OnStop(f func()) {
	ng.onStop = f
}

type Task func() []Task

// 有没有可能父协程return后，子协程还没开始执行
// 这里有问题，存在协程快速执行完后直接执行回调的情况，此时并非所有任务完成
// 只能在任务中添加新任务
func (ng *Engine) AddTask(fns ...Task) {
	if ng.ran.Load() {
		if ng.num.Load() != 0 {
			ng.run(fns...)
		}
	} else {
		ng.run(fns...)
		ng.ran.Store(true)
	}

}

func (ng *Engine) run(fns ...Task) {
	ng.num.Add(int32(len(fns)))
	for _, fn := range fns {
		go func() {
			tasks := fn()
			if len(tasks) > 0 {
				ng.run(tasks...)
			}
			ng.num.Add(-1)
			if ng.num.Load() == 0 {
				ng.onStop()
			}
		}()
	}
}

func (ng *Engine) Context() context.Context {
	return ng.ctx
}

func (ng *Engine) Cancel() {
	ng.cancel()
}
