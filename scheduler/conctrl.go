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

func (e *Engine[KEY]) Run(tasks ...*Task[KEY]) {
	// 在持锁前递交初始任务:addTasks 内部会获取 pending 额度并在短临界区内 push,
	// 若持锁调用会触发非重入锁自死锁。
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
					// 排空残留错误任务,确保每个 pending 任务都 wg.Done,
					// 否则 Run 的 wg.Wait 会因计数未归零而永久阻塞
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
					e.taskDone() // 先归还 pending 额度,避免 errHandler 内 re-add 与 pendingSem 形成自死锁
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
			defer e.wg.Done() // 与上面 e.wg.Add(1) 配对,所有退出路径都必须 Done
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
					// 新任务入堆,立即回到循环顶部尝试分发
				case <-timer.C:
					//检测任务是否已空(堆长度与 workingWorkerCount 的读取都需加锁/原子,避免与 addTasks 竞争)
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
					// 取消时回收堆中尚未分发、以及已出堆但未成功投递给 worker 的任务计数,
					// 否则这些任务永不 Done,Run 的 wg.Wait 会永久阻塞
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

func (e *Engine[KEY]) newWorker(readyTask *Task[KEY]) {
	atomic.AddUint64(&e.currentWorkerCount, 1)
	//id := c.currentWorkerCount
	// 这里考虑回复多channel,worker数量多起来的时候,channel维护的goroutine数量太多
	worker := &Worker[KEY]{id: uint(e.currentWorkerCount)}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				worker.canExecute = false
				log.StackLogger().Error(r, spew.Sdump(readyTask))
				atomic.AddUint64(&e.taskFailedCount, 1)
				e.taskDone()
				// 创建一个新的
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
					// channel 已被关闭(引擎停止),worker 退出,避免收到 nil 任务空指针 panic
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

func (e *Engine[KEY]) addTasks(ctx context.Context, priority int, tasks ...*Task[KEY]) {
	// 外部/初始任务入口:先在不持锁的情况下获取 pending 额度(背压点),额度满则阻塞调用者。
	// 仅对有效任务获取额度,与 taskDone 的归还严格 1:1。
	// 注意:绝不可在持有 e.mu 时执行这里,否则主循环无法拿到 e.mu 分发任务 -> 死锁。
	// worker 产出的子任务不走此路径(pending 额度由 worker 在 produce 时占用,见 execTask),
	// 因此 ingest 协程永不在此阻塞,submitCh 始终被及时消费,不存在背压自死锁。
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
	// 去重跳过的任务归还多占的 pending 额度
	if skipped := valid - n; skipped > 0 && e.maxPending > 0 && e.pendingSem != nil {
		for i := 0; i < skipped; i++ {
			e.pendingSem <- struct{}{}
		}
	}
	atomic.AddUint64(&e.taskTotalCount, uint64(n))
	e.wg.Add(n)
}

// pushTasks 仅把任务 push 进堆并初始化,不触碰 pendingSem / wg。
// 返回实际加入的有效任务数,供调用者配对 wg.Add。
// 调用者负责在调用前已占用 pending 额度(外部经 addTasks,worker 经 execTask)。
func (e *Engine[KEY]) pushTasks(ctx context.Context, priority int, tasks ...*Task[KEY]) int {
	n := 0
	e.mu.Lock()
	for _, task := range tasks {
		if task == nil || task.Run == nil {
			continue
		}
		// 提交时去重:相同 key 只入堆一次
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
		// 非阻塞通知 dispatcher 有新任务入堆
		select {
		case e.wakeup <- struct{}{}:
		default:
		}
	}
	return n
}

// taskDone 在任务终态时归还 pending 额度并递减 wg,与额度获取严格 1:1 配对。
func (e *Engine[KEY]) taskDone() {
	if e.maxPending > 0 && e.pendingSem != nil {
		e.pendingSem <- struct{}{}
	}
	e.wg.Done()
}

// ingest 专门负责把子任务送入堆。它只 push 堆、不获取 pendingSem(额度已由 worker 在
// produce 时占用),因此本协程永远能及时消费 submitCh,绝不阻塞 worker。
// 取消时排空 submitCh 中残留任务后退出,worker 侧已有 ctx.Done 分支不会永久阻塞。
func (e *Engine[KEY]) ingest() {
	for {
		select {
		case <-e.ctx.Done():
			// 排空残留任务后退出
			for {
				select {
				case <-e.submitCh:
				default:
					return
				}
			}
		case task := <-e.submitCh:
			// worker 已在 produce 时占用 pending 额度,这里只负责 push + wg.Add。
			n := e.pushTasks(task.ctx, task.Priority, task)
			e.wg.Add(n)
		}
	}
}

func (e *Engine[KEY]) AddOptionTasks(ctx context.Context, priority int, tasks ...*Task[KEY]) {
	e.addTasks(ctx, priority, tasks...)
}

func (e *Engine[KEY]) AddTasks(tasks ...*Task[KEY]) {
	e.addTasks(nil, 0, tasks...)
}

func (e *Engine[KEY]) AddWorker(num int) {
	atomic.AddUint64(&e.workerCount, uint64(num))
	e.addWorker()
}

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

func (e *Engine[KEY]) fixedWorker(workerId int) *Worker[KEY] {
	e.workersMu.Lock()
	defer e.workersMu.Unlock()
	if workerId < 0 || workerId > len(e.workers)-1 {
		return nil
	}
	return e.workers[workerId]
}

func (e *Engine[KEY]) newFixedWorker(worker *Worker[KEY], interval time.Duration) {
	go func() {
		var task *Task[KEY]
		defer func() {
			if r := recover(); r != nil {
				worker.canExecute = false
				log.StackLogger().Error(r, spew.Sdump(task))
				atomic.AddUint64(&e.taskFailedCount, 1)
				e.wg.Done()
				// 创建一个新的
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
				// 引擎停止时退出,避免 fixed worker 永久阻塞在 taskCh 上
				return
			}
		}
	}()
}

func (e *Engine[KEY]) AddFixedTasks(workerId int, generation int, tasks ...*Task[KEY]) error {
	err := fmt.Errorf("不存在workId为%d的fixed worker,请调用NewFixedWorker添加", workerId)
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

		// ctx 已取消时不再重试/投递错误队列,归还 wg 后返回,避免 cancel 后任务堆积导致 Run 永久阻塞
		if e.ctx.Err() != nil {
			e.taskDone()
			return false
		}

		if task.errTimes < 5 {
			task.reExecTimes++
			log.Warnf("%v执行失败:%v,将第%d次执行", task.Key, err, task.reExecTimes)
			task.Priority++
			e.mu.Lock()
			e.readyTaskHeap.Push(task)
			e.mu.Unlock()
		} else {
			log.Warn(task.Key, "多次执行失败:", err, "将执行错误处理")
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
			// worker 在 produce 子任务时先占用 pending 额度(背压在此生效,减慢子任务生成速率),
			// 再非阻塞发往 submitCh。ingest 协程只 push 堆、不获取额度,因此始终及时消费 submitCh,
			// worker 不会被 ingest 阻塞。额度与 taskDone 的归还严格 1:1。
			for _, c := range tasks {
				if e.maxPending > 0 && e.pendingSem != nil {
					select {
					case <-e.pendingSem:
					case <-e.ctx.Done():
						return true // 取消时丢弃尚未入队的子任务(未计入 wg,无泄漏)
					}
				}
				select {
				case e.submitCh <- c:
				case <-e.ctx.Done():
					return true // 取消时丢弃尚未入队的子任务(未计入 wg,无泄漏)
				}
			}
		} else {
			e.AddOptionTasks(task.ctx, task.Priority+1, tasks...)
		}
	}
	atomic.AddUint64(&e.taskDoneCount, 1)
	return true
}

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

func (e *Engine[KEY]) StopAfter(interval time.Duration) *Engine[KEY] {
	time.AfterFunc(interval, e.Stop)
	return e
}
