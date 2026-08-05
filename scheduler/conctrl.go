/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hopeio/gox/idgen"
	"github.com/hopeio/gox/log"
	syncx "github.com/hopeio/gox/sync"
)

// Run executes the operation.
func (e *Engine[KEY]) Run(tasks ...*Task[KEY]) {
	// Submit initial tasks before holding the lock: addTasks acquires pending quota and pushes in a short critical section,
	// Calling while holding the lock would self-deadlock the non-reentrant mutex.
	if len(tasks) > 0 {
		e.addTasks(e.ctx, 0, tasks...)
	}
	e.mu.Lock()
	if e.isRunning {
		e.mu.Unlock()
		return
	}
	if !e.errHandlerRunning.Load() {
		go func() {
			for {
				select {
				case <-e.ctx.Done():
					// Drain leftover error tasks so every pending task calls wg.Done,
					// otherwise Run's wg.Wait blocks forever if the count never reaches zero
					for {
						select {
						case task := <-e.errTaskChan:
							_ = task
							e.taskDone()
						default:
							e.errHandlerRunning.Store(false)
							return
						}
					}
				case task := <-e.errTaskChan:
					e.taskErrHandleCount++
					e.taskDone() // Return pending quota first to avoid self-deadlock between errHandler re-add and pendingSem
					e.errHandler(task)
				}
			}
		}()
		e.errHandlerRunning.Store(true)
	}
	e.addWorker()
	if !e.isRunning {
		e.isRunning = true
		e.wg.Add(1)
		go func() {
			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()
			defer e.wg.Done() // Pairs with e.wg.Add(1) above; every exit path must Done
			var emptyTimes uint
			var readyTaskCh chan *Task[KEY]
			var readyTask *Task[KEY]

		loop:
			for {
				e.mu.Lock()
				if len(e.readyTaskHeap) > 0 && readyTask == nil {
					readyTask, _ = e.readyTaskHeap.Pop()
					readyTaskCh = e.taskChanConsumer
				}
				e.mu.Unlock()
				select {
				case readyTaskCh <- readyTask:
					readyTaskCh = nil
					readyTask = nil
				case <-e.wakeup:
					// New tasks entered the heap; jump to the top of the loop to dispatch
				case <-timer.C:
					//Check whether tasks are empty (heap len and workingWorkerCount need lock/atomic vs addTasks)
					e.mu.Lock()
					heapEmpty := len(e.readyTaskHeap) == 0
					e.mu.Unlock()
					if atomic.LoadUint64(&e.workingWorkerCount) == 0 && heapEmpty {
						e.mu.Lock()
						counter, _ := syncx.WaitGroupState(&e.wg)
						if counter == 1 {
							emptyTimes++
							if emptyTimes > 2 {
								log.NoCallerLogger().Debug("the task is about to end.")
								e.isRunning = false
								e.mu.Unlock()
								break loop
							}
						}
						e.mu.Unlock()
					}
					e.mu.Lock()
					heapLen := len(e.readyTaskHeap)
					e.mu.Unlock()
					fmt.Printf("[Running] task:R:%d,D:%d/T:%d/S:%d/H:%d/F:%d/E:%d,worker: %d/%d\r", heapLen,
						atomic.LoadUint64(&e.taskDoneCount), atomic.LoadUint64(&e.taskTotalCount), atomic.LoadUint64(&e.taskSkipCount), atomic.LoadUint64(&e.taskErrHandleCount), atomic.LoadUint64(&e.taskFailedCount), atomic.LoadUint64(&e.taskErrorTimes), atomic.LoadUint64(&e.workingWorkerCount), atomic.LoadUint64(&e.currentWorkerCount))
					timer.Reset(e.monitorInterval)
				case <-e.ctx.Done():
					if err := e.ctx.Err(); err != nil {
						log.Error(err)
					}
					// On cancel, reclaim counts for undispatched heap tasks and tasks popped but not handed to a worker,
					// otherwise those tasks never Done and Run's wg.Wait blocks forever
					e.mu.Lock()
					for len(e.readyTaskHeap) > 0 {
						if _, ok := e.readyTaskHeap.Pop(); ok {
							e.taskDone()
						}
					}
					if readyTask != nil {
						e.taskDone()
						readyTask = nil
					}
					e.mu.Unlock()
					e.isRunning = false
					close(e.taskChanConsumer)
					break loop
				}

			}
		}()
	}
	e.mu.Unlock()
	e.wg.Wait()
	log.NoCallerLogger().Infof("[END] task:D:%d/T:%d/S:%d/H:%d/F:%d/E:%d", e.taskDoneCount, e.taskTotalCount, e.taskSkipCount, e.taskErrHandleCount, e.taskFailedCount, e.taskErrorTimes)
}

// newWorker creates and returns a new instance.
func (e *Engine[KEY]) newWorker(readyTask *Task[KEY]) {
	atomic.AddUint64(&e.currentWorkerCount, 1)
	//id := c.currentWorkerCount
	// Consider multiple channels; with many workers, one channel may create too many managing goroutines
	worker := &Worker[KEY]{id: uint(e.currentWorkerCount)}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				worker.canExecute = false
				log.StackLogger().Error(r, spew.Sdump(readyTask))
				atomic.AddUint64(&e.taskFailedCount, 1)
				e.taskDone()
				// create a new one
				e.newWorker(nil)
			}
			atomic.AddUint64(&e.currentWorkerCount, ^uint64(0))
		}()
		worker.canExecute = true
		if readyTask != nil {
			e.ExecTask(worker, readyTask)
		}
		var ok bool
		for {
			select {
			case readyTask, ok = <-e.taskChanConsumer:
				if !ok {
					// Channel closed (engine stopped); worker exits to avoid nil-task panic
					worker.canExecute = false
					return
				}
				e.ExecTask(worker, readyTask)
			case <-e.ctx.Done():
				worker.canExecute = false
				return
			}
		}
	}()
	e.workersMu.Lock()
	e.workers = append(e.workers, worker)
	e.workersMu.Unlock()
}

// addWorker performs the operation.
func (e *Engine[KEY]) addWorker() {
	if atomic.LoadUint64(&e.currentWorkerCount) == 0 {
		e.newWorker(nil)
	}
	if e.workerFactoryRunning.Load() {
		return
	}
	go func() {
		e.workerFactoryRunning.Store(true)
		for {
			select {
			case readyTask, ok := <-e.taskChanConsumer:
				if !ok {
					return
				}
				if atomic.LoadUint64(&e.currentWorkerCount) < atomic.LoadUint64(&e.workerCount) {
					e.newWorker(readyTask)
				} else {
					log.Info("worker count is full")
					e.mu.Lock()
					e.readyTaskHeap.Push(readyTask)
					e.mu.Unlock()
					e.workerFactoryRunning.Store(false)
					return
				}
			case <-e.ctx.Done():
				return
			}
		}
	}()

}

// addTasks performs the operation.
func (e *Engine[KEY]) addTasks(ctx context.Context, priority int, tasks ...*Task[KEY]) {
	// External/initial task entry: acquire pending quota without the lock (backpressure); block when full.
	// Acquire quota only for valid tasks, strictly 1:1 with taskDone returns.
	// Never run this while holding e.mu, or the main loop cannot take e.mu to dispatch -> deadlock.
	// Worker-produced subtasks skip this path (pending quota taken at produce in execTask),
	// so the ingest goroutine never blocks here, submitCh stays drained, no backpressure self-deadlock.
	valid := 0
	for _, task := range tasks {
		if task != nil && task.Run != nil {
			if e.maxPending > 0 && e.pendingSem != nil {
				<-e.pendingSem
			}
			valid++
		}
	}
	n := e.pushTasks(ctx, priority, tasks...)
	// Return extra pending quota for tasks skipped by dedup
	if skipped := valid - n; skipped > 0 && e.maxPending > 0 && e.pendingSem != nil {
		for i := 0; i < skipped; i++ {
			e.pendingSem <- struct{}{}
		}
	}
	atomic.AddUint64(&e.taskTotalCount, uint64(n))
	e.wg.Add(n)
}

// pushTasks returns the result.
func (e *Engine[KEY]) pushTasks(ctx context.Context, priority int, tasks ...*Task[KEY]) int {
	n := 0
	e.mu.Lock()
	for _, task := range tasks {
		if task == nil || task.Run == nil {
			continue
		}
		// Dedup on submit: the same key enters the heap only once
		if task.Key != e.zeroKey {
			if _, exists := e.seen[task.Key]; exists {
				atomic.AddUint64(&e.taskSkipCount, 1)
				continue
			}
			e.seen[task.Key] = struct{}{}
		}
		if ctx != nil {
			task.ctx = ctx
		}
		if task.ctx == nil {
			task.ctx = e.ctx
		}
		task.Priority = priority
		task.id = idgen.NewOrderedID()
		e.readyTaskHeap.Push(task)
		n++
	}
	e.mu.Unlock()
	if n > 0 {
		// Non-blocking notify that new tasks entered the heap
		select {
		case e.wakeup <- struct{}{}:
		default:
		}
	}
	return n
}

// taskDone performs the operation.
func (e *Engine[KEY]) taskDone() {
	if e.maxPending > 0 && e.pendingSem != nil {
		e.pendingSem <- struct{}{}
	}
	e.wg.Done()
}

// ingest performs the operation.
func (e *Engine[KEY]) ingest() {
	for {
		select {
		case <-e.ctx.Done():
			// Drain leftover tasks then exit
			for {
				select {
				case <-e.submitCh:
				default:
					return
				}
			}
		case task := <-e.submitCh:
			// Worker already took pending quota at produce; here only push + wg.Add.
			n := e.pushTasks(task.ctx, task.Priority, task)
			e.wg.Add(n)
		}
	}
}

// AddOptionTasks updates or inserts a value.
func (e *Engine[KEY]) AddOptionTasks(ctx context.Context, priority int, tasks ...*Task[KEY]) {
	e.addTasks(ctx, priority, tasks...)
}

// AddTasks updates or inserts a value.
func (e *Engine[KEY]) AddTasks(tasks ...*Task[KEY]) {
	e.addTasks(nil, 0, tasks...)
}

// AddWorker updates or inserts a value.
func (e *Engine[KEY]) AddWorker(num int) {
	atomic.AddUint64(&e.workerCount, uint64(num))
	e.addWorker()
}

// NewFixedWorker creates and returns a new instance.
func (e *Engine[KEY]) NewFixedWorker(interval time.Duration) int {
	taskChan := make(chan *Task[KEY])
	worker := &Worker[KEY]{id: uint(e.currentWorkerCount), typ: fixedType, taskCh: taskChan}
	e.workersMu.Lock()
	e.workers = append(e.workers, worker)
	idx := len(e.workers) - 1
	e.workersMu.Unlock()
	e.newFixedWorker(worker, interval)
	return idx
}

// fixedWorker returns the result.
func (e *Engine[KEY]) fixedWorker(workerId int) *Worker[KEY] {
	e.workersMu.Lock()
	defer e.workersMu.Unlock()
	if workerId < 0 || workerId > len(e.workers)-1 {
		return nil
	}
	return e.workers[workerId]
}

// newFixedWorker creates and returns a new instance.
func (e *Engine[KEY]) newFixedWorker(worker *Worker[KEY], interval time.Duration) {
	go func() {
		var task *Task[KEY]
		defer func() {
			if r := recover(); r != nil {
				worker.canExecute = false
				log.StackLogger().Error(r, spew.Sdump(task))
				atomic.AddUint64(&e.taskFailedCount, 1)
				e.wg.Done()
				// create a new one
				e.newFixedWorker(worker, interval)
			}
		}()
		worker.canExecute = true
		var ok bool
		for {
			select {
			case task, ok = <-worker.taskCh:
				if !ok {
					return
				}
				if interval > 0 {
					timer := time.NewTimer(interval)
					<-timer.C
					timer.Stop()
				}
				e.ExecTask(worker, task)
			case <-e.ctx.Done():
				// Exit when the engine stops so fixed workers do not block forever on taskCh
				return
			}
		}
	}()
}

// AddFixedTasks updates or inserts a value.
func (e *Engine[KEY]) AddFixedTasks(workerId int, generation int, tasks ...*Task[KEY]) error {
	err := fmt.Errorf("fixed worker with workId %d does not exist; call NewFixedWorker to add it", workerId)
	worker := e.fixedWorker(workerId)
	if worker == nil {
		return err
	}
	if worker.typ != fixedType {
		return err
	}
	l := len(tasks)
	atomic.AddUint64(&e.taskTotalCount, uint64(l))
	e.wg.Add(l)
	for _, task := range tasks {
		if task == nil || task.Run == nil {
			atomic.AddUint64(&e.taskTotalCount, ^uint64(0))
			e.wg.Done()
			continue
		}
		if task.ctx == nil {
			task.ctx = e.ctx
		}
		task.Priority += generation
		task.id = idgen.NewOrderedID()
		worker.taskCh <- task
	}
	return nil
}

// ExecTask performs the operation.
func (e *Engine[KEY]) ExecTask(worker *Worker[KEY], task *Task[KEY]) {
	if task == nil {
		return
	}
	atomic.AddUint64(&e.workingWorkerCount, 1)
	worker.isExecuting = true
	worker.currentTask = task
	defer func() {
		atomic.AddUint64(&e.workingWorkerCount, ^uint64(0))
		worker.isExecuting = false
	}()
	if e.execTask(task) {
		e.taskDone()
	}
}

// execTask reports whether the condition holds.
func (e *Engine[KEY]) execTask(task *Task[KEY]) bool {

	if e.speedLimit != nil {
		e.speedLimit.Wait()
	}

	if e.rateLimiter != nil {
		err := e.rateLimiter.Wait(task.ctx)
		if err != nil {
			log.Warnf("rate limit err:%v", err)
		}
	}

	var kindHandler *KindHandler[KEY]
	if e.kindHandlers != nil && int(task.Kind) < len(e.kindHandlers) {
		kindHandler = e.kindHandlers[task.Kind]
	}

	if kindHandler != nil {
		if kindHandler.Skip {
			atomic.AddUint64(&e.taskSkipCount, 1)
			return true
		}

		if kindHandler.speedLimit != nil {
			kindHandler.speedLimit.Wait()
		}
		if kindHandler.rateLimiter != nil {
			err := kindHandler.rateLimiter.Wait(task.ctx)
			if err != nil {
				log.Warnf("kind rate limit err:%v", err)
			}
		}
	}

	if task.reExecTimes > 0 {
		task.reExecLogs = append(task.reExecLogs, &execLog{
			execBeginAt: time.Now(),
		})
	} else {
		task.execBeginAt = time.Now()
	}
	tasks, err := task.Run.Run(task.ctx)
	if task.reExecTimes > 0 {
		task.reExecLogs[len(task.reExecLogs)-1].execEndAt = time.Now()
	} else {
		task.execEndAt = time.Now()
	}

	if err != nil {
		atomic.AddUint64(&e.taskErrorTimes, 1)
		task.errTimes++
		if task.reExecTimes > 0 {
			task.reExecLogs[len(task.reExecLogs)-1].err = err
		} else {
			task.err = err
		}

		// If ctx is canceled, skip retry/error queue, return after wg, avoid Run hang from piled tasks
		if e.ctx.Err() != nil {
			e.taskDone()
			return false
		}

		if task.errTimes < 5 {
			task.reExecTimes++
			log.Warnf("%v failed: %v; will retry for the %d time(s)", task.Key, err, task.reExecTimes)
			task.Priority++
			e.mu.Lock()
			e.readyTaskHeap.Push(task)
			e.mu.Unlock()
		} else {
			log.Warn(task.Key, "failed repeatedly:", err, "; running error handler")
			select {
			case e.errTaskChan <- task:
			case <-e.ctx.Done():
				e.taskDone()
			}
		}

		return false
	}
	if len(tasks) > 0 && e.ctx.Err() == nil {
		if e.submitCh != nil {
			// Worker takes pending quota when producing subtasks (backpressure slows production),
			// then non-blocking send to submitCh. ingest only pushes the heap and never takes quota, so it always drains submitCh,
			// workers are never blocked by ingest. Quota returns are 1:1 with taskDone.
			for _, c := range tasks {
				if e.maxPending > 0 && e.pendingSem != nil {
					select {
					case <-e.pendingSem:
					case <-e.ctx.Done():
						return true // On cancel, drop not-yet-queued subtasks (not in wg; no leak)
					}
				}
				select {
				case e.submitCh <- c:
				case <-e.ctx.Done():
					return true // On cancel, drop not-yet-queued subtasks (not in wg; no leak)
				}
			}
		} else {
			e.AddOptionTasks(task.ctx, task.Priority+1, tasks...)
		}
	}
	atomic.AddUint64(&e.taskDoneCount, 1)
	return true
}

// Stop closes and releases resources.
func (e *Engine[KEY]) Stop() {
	e.cancel()
	if e.speedLimit != nil {
		e.speedLimit.Stop()
	}
	for _, kindHandler := range e.kindHandlers {
		if kindHandler != nil {
			if kindHandler.speedLimit != nil {
				kindHandler.speedLimit.Stop()
			}
			if kindHandler.rateLimiter != nil {
				kindHandler.rateLimiter = nil
			}
		}
	}

	for _, callback := range e.onStop {
		callback(e.ctx)
	}
	e.isStopped = true
}

// StopAfter closes and releases resources.
func (e *Engine[KEY]) StopAfter(interval time.Duration) *Engine[KEY] {
	time.AfterFunc(interval, e.Stop)
	return e
}
