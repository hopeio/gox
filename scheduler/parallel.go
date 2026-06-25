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

func NewParallel(workNum uint, opts ...ParallelOption) *Parallel {
	taskCh := make(chan func(), workNum)
	p := &Parallel{taskCh: taskCh}
	g := func() {
		defer func() {
			if err := recover(); err != nil {
				log.StackLogger().Error(err)
			}
		}()
		for task := range taskCh {
			task()
			p.wg.Done()
		}
	}
	for range workNum {
		go g()
	}
	return p
}

func (p *Parallel) AddFunc(task func()) {
	p.wg.Add(1)
	p.taskCh <- task
}

func (p *Parallel) Wait() {
	p.wg.Wait()
}

func (p *Parallel) Stop() {
	p.wg.Wait()
	close(p.taskCh)
}

type ParallelOption func(p *Parallel)

type Funcs []func()

func (t *Funcs) Do()  {
	taskChain := *t
	for i := 0; i < len(taskChain); i++ {
		taskChain[i]()
		}
	return
}

