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
	g := func() {
		var executing bool
		defer func() {
			if err := recover(); err != nil {
				log.StackLogger().Error(err)
			}
			if executing {
				p.wg.Done()
			}
		}()
		for task := range taskCh {
			executing = true
			task()
			executing = false
			p.wg.Done()
		}
	}
	for range workNum {
		go g()
	}
	return p
}

// AddFunc ...
func (p *Parallel) AddFunc(task func()) {
	p.wg.Add(1)
	p.taskCh <- task
}

// Wait ...
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

// Do ...
func (t *Funcs) Do() {
	taskChain := *t
	for i := 0; i < len(taskChain); i++ {
		taskChain[i]()
	}
}
