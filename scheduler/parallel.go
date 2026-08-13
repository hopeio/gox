/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"sync"

	"github.com/hopeio/gox/log"
)

type Parallel struct {
	taskCh chan func()
	wg     sync.WaitGroup
}

// NewParallel creates and returns a new instance.
func NewParallel(workNum uint, opts ...ParallelOption) *Parallel {
	taskCh := make(chan func(), workNum)
	p := &Parallel{taskCh: taskCh}
	for _, opt := range opts {
		opt(p)
	}
	// recover 放在单任务粒度：曾放在 worker 粒度，一个 panic 任务会永久杀死一个 worker
	runTask := func(task func()) {
		defer func() {
			if err := recover(); err != nil {
				log.StackLogger().Error(err)
			}
			p.wg.Done()
		}()
		task()
	}
	g := func() {
		for task := range taskCh {
			runTask(task)
		}
	}
	for range workNum {
		go g()
	}
	return p
}

// AddFunc updates or inserts a value.
func (p *Parallel) AddFunc(task func()) {
	p.wg.Add(1)
	p.taskCh <- task
}

// Wait performs the operation.
func (p *Parallel) Wait() {
	p.wg.Wait()
}

// Stop closes and releases resources.
func (p *Parallel) Stop() {
	p.wg.Wait()
	close(p.taskCh)
}

type ParallelOption func(p *Parallel)

type Funcs []func()

// Do executes the operation.
func (t *Funcs) Do() {
	taskChain := *t
	for i := 0; i < len(taskChain); i++ {
		taskChain[i]()
	}
}
