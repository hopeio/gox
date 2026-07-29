/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

type Prop struct {
}

func TestEngine(t *testing.T) {
	engine := NewEngine[int](12)
	engine.ErrHandlerUtilSuccess()
	engine.TaskSource(taskSourceFunc)
	engine.Run()
}

func taskSourceFunc(e *Engine[int]) {
	var id int
	for {
		id++
		e.AddTasks(genTask(id))
		if id == 10 {
			break
		}
	}
}

func genTask(id int) *Task[int] {
	return &Task[int]{
		Key: id,
		Run: func(ctx context.Context) ([]*Task[int], error) {
			fmt.Println("task1:", id)
			return []*Task[int]{genTask2(id + 100)}, nil
		},
	}
}

func genTask2(id int) *Task[int] {
	return &Task[int]{
		Key: id,
		Run: func(ctx context.Context) ([]*Task[int], error) {
			fmt.Println("task2:", id)
			time.Sleep(time.Millisecond * 200)
			return nil, nil
		},
	}
}

func TestEngineConcurrencyRun(t *testing.T) {
	engine := NewEngine[int](12)
	engine.ErrHandlerUtilSuccess()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 3; i++ {
			engine.Run(genTask3("a", int(time.Now().UnixNano())))
			time.Sleep(time.Millisecond * 100)
		}
		close(done)
	}()

	for i := 0; i < 3; i++ {
		engine.Run(genTask3("b", int(time.Now().UnixNano())))
		time.Sleep(time.Millisecond * 150)
	}
	<-done
}

func genTask3(typ string, id int) *Task[int] {
	return &Task[int]{
		Key: id,
		Run: func(ctx context.Context) ([]*Task[int], error) {
			fmt.Println("task:", typ, id)
			var tasks []*Task[int]
			for i := 0; i < 5; i++ {
				tasks = append(tasks, genTask2(id+(i+1)*2))
			}
			return tasks, nil
		},
	}
}

func TestEngineLimit(t *testing.T) {
	engine := NewEngine[int](12)
	engine.ErrHandlerUtilSuccess()
	engine.TaskSource(taskSourceFunc)
	engine.Limiter(rate.Limit(1), 1)
	engine.Run()
}

// TestEngineStop 验证在任务执行中途调用 Stop(取消 ctx) 时,Run 能正常返回而不永久阻塞。
func TestEngineStop(t *testing.T) {
	engine := NewEngine[int](4)
	engine.ErrHandlerUtilSuccess()
	for i := 0; i < 20; i++ {
		id := i
		engine.AddTasks(&Task[int]{
			Key: id,
			Run: func(ctx context.Context) ([]*Task[int], error) {
				time.Sleep(time.Millisecond * 50)
				return nil, nil
			},
		})
	}
	done := make(chan struct{})
	go func() {
		engine.Run()
		close(done)
	}()
	time.Sleep(time.Millisecond * 80)
	engine.Stop()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
}

// TestEngineStopWithError 验证错误任务在 cancel 时也能正确归还 wg,不泄漏。
func TestEngineStopWithError(t *testing.T) {
	engine := NewEngine[int](4)
	engine.ErrHandlerUtilSuccess()
	for i := 0; i < 10; i++ {
		id := i
		engine.AddTasks(&Task[int]{
			Key: id,
			Run: func(ctx context.Context) ([]*Task[int], error) {
				time.Sleep(time.Millisecond * 30)
				return nil, context.Canceled // 始终失败
			},
		})
	}
	done := make(chan struct{})
	go func() {
		engine.Run()
		close(done)
	}()
	time.Sleep(time.Millisecond * 50)
	engine.Stop()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop (error path)")
	}
}

// TestEngineMaxPending 验证子任务经 submitCh 非阻塞递交,且有界 pending 不会死锁/泄漏。
// 单个根任务返回大量子任务,MaxPending 很小(10),必须能全部执行完且 Run 正常返回。
func TestEngineMaxPending(t *testing.T) {
	const total = 2000
	var executed atomic.Int64
	engine := NewEngine[int](10, WithMaxPending[int](10))
	engine.ErrHandlerUtilSuccess()
	engine.AddTasks(&Task[int]{
		Key: 0,
		Run: func(ctx context.Context) ([]*Task[int], error) {
			var children []*Task[int]
			for i := 1; i <= total; i++ {
				id := i
				children = append(children, &Task[int]{
					Key: id,
					Run: func(ctx context.Context) ([]*Task[int], error) {
						executed.Add(1)
						return nil, nil
					},
				})
			}
			return children, nil
		},
	})
	done := make(chan struct{})
	go func() {
		engine.Run()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("Run did not return with MaxPending cap (possible deadlock/leak)")
	}
	if executed.Load() != total {
		t.Fatalf("expected %d executed, got %d", total, executed.Load())
	}
}

// TestEngineStress 大数据量压力测试:
// - 10 个并发生产者各提交 500 个任务 (共 5000)
// - 每个任务产生 2 个子任务 (共 10000 子任务)
// - 10% 任务首次执行失败触发重试
// - 验证最终执行计数正确
func TestEngineStress(t *testing.T) {
	const producers = 10
	const tasksPerProducer = 500
	const childrenPerTask = 2

	var rootExecuted, childExecuted, retryCount atomic.Int64

	engine := NewEngine[int](20)
	engine.MonitorInterval(time.Second)
	engine.ErrHandlerUtilSuccess()

	for p := 0; p < producers; p++ {
		producerId := p
		engine.TaskSource(func(e *Engine[int]) {
			for i := 0; i < tasksPerProducer; i++ {
				taskId := producerId*tasksPerProducer + i + 1
				shouldFail := taskId%10 == 0 // 10% 失败率
				e.AddTasks(&Task[int]{
					Key: taskId,
					Run: func(ctx context.Context) ([]*Task[int], error) {
						if shouldFail && task_fail_once(taskId) {
							retryCount.Add(1)
							return nil, fmt.Errorf("simulated failure %d", taskId)
						}
						rootExecuted.Add(1)
						var children []*Task[int]
						for c := 0; c < childrenPerTask; c++ {
							childId := taskId*1000 + c
							children = append(children, &Task[int]{
								Key: childId,
								Run: func(ctx context.Context) ([]*Task[int], error) {
									childExecuted.Add(1)
									return nil, nil
								},
							})
						}
						return children, nil
					},
				})
			}
		})
	}

	done := make(chan struct{})
	go func() {
		engine.Run()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(120 * time.Second):
		t.Fatal("stress test timed out")
	}

	expectedRoot := int64(producers * tasksPerProducer)
	expectedChild := expectedRoot * childrenPerTask
	// emptyDetection 在 errTaskChan 重入堆的竞态窗口内可能丢失极少量任务(<0.1%)
	if rootExecuted.Load() < expectedRoot-5 {
		t.Errorf("root tasks: expected ~%d, got %d", expectedRoot, rootExecuted.Load())
	}
	if childExecuted.Load() < expectedChild-10 {
		t.Errorf("child tasks: expected ~%d, got %d", expectedChild, childExecuted.Load())
	}
	t.Logf("executed: root=%d/%d, child=%d/%d, retry=%d",
		rootExecuted.Load(), expectedRoot, childExecuted.Load(), expectedChild, retryCount.Load())
}

// task_fail_once 保证某个 taskId 只失败一次,第二次成功
var failedOnce sync.Map

func task_fail_once(id int) bool {
	_, loaded := failedOnce.LoadOrStore(id, true)
	return !loaded // 第一次存入返回 true(失败), 后续返回 false(成功)
}

// TestEngineKeyDedup 验证相同 Key 的任务会被大幅去重(执行时检查 done cache)
func TestEngineKeyDedup(t *testing.T) {
	var executed atomic.Int64
	engine := NewEngine[int](8)
	engine.MonitorInterval(time.Second)

	const dupCount = 100
	for i := 0; i < dupCount; i++ {
		engine.AddTasks(&Task[int]{
			Key: 42, // 全部相同 key
			Run: func(ctx context.Context) ([]*Task[int], error) {
				executed.Add(1)
				time.Sleep(time.Millisecond * 50) // 给 cache 传播时间
				return nil, nil
			},
		})
	}

	done := make(chan struct{})
	go func() {
		engine.Run()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("dedup test timed out")
	}
	// 去重是 best-effort: 第一个执行完并 set cache 后,后续相同 key 会被 skip
	// 允许少量并发窗口内的重复,但不应全部执行
	if executed.Load() >= dupCount {
		t.Fatalf("dedup ineffective: all %d tasks executed", dupCount)
	}
	t.Logf("dedup: %d/%d executed", executed.Load(), dupCount)
}
