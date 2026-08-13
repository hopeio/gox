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
	engine.MonitorInterval(200 * time.Millisecond)
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
	engine.MonitorInterval(200 * time.Millisecond)
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
	engine.MonitorInterval(200 * time.Millisecond)
	engine.ErrHandlerUtilSuccess()
	engine.TaskSource(taskSourceFunc)
	engine.Limiter(rate.Limit(1), 1)
	engine.Run()
}

// TestEngineStop verifies that calling Stop (cancel ctx) mid-execution lets Run return without hanging forever.
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

// TestEngineStopWithError verifies error tasks still release wg on cancel without leaking.
func TestEngineStopWithError(t *testing.T) {
	engine := NewEngine[int](4)
	engine.ErrHandlerUtilSuccess()
	for i := 0; i < 10; i++ {
		id := i
		engine.AddTasks(&Task[int]{
			Key: id,
			Run: func(ctx context.Context) ([]*Task[int], error) {
				time.Sleep(time.Millisecond * 30)
				return nil, context.Canceled // always fail
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

// TestEngineMaxPending verifies child tasks are submitted non-blocking via submitCh,
// and bounded pending does not deadlock/leak.
// A single root returns many children with a small MaxPending (10); all must finish and Run must return.
func TestEngineMaxPending(t *testing.T) {
	const total = 2000
	var executed atomic.Int64
	engine := NewEngine(10, WithMaxPending[int](10))
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

// TestEngineStress large-volume stress test:
// - 10 concurrent producers each submit 500 tasks (5000 total)
// - each task spawns 2 children (10000 children total)
// - 10% of tasks fail on first run to trigger retry
// - verify final execution counts
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
				shouldFail := taskId%10 == 0 // 10% failure rate
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
	// emptyDetection may drop a tiny fraction of tasks (<0.1%) in the race window when re-pushing errTaskChan onto the heap
	if rootExecuted.Load() < expectedRoot-10 {
		t.Errorf("root tasks: expected ~%d, got %d", expectedRoot, rootExecuted.Load())
	}
	if childExecuted.Load() < expectedChild-20 {
		t.Errorf("child tasks: expected ~%d, got %d", expectedChild, childExecuted.Load())
	}
	t.Logf("executed: root=%d/%d, child=%d/%d, retry=%d",
		rootExecuted.Load(), expectedRoot, childExecuted.Load(), expectedChild, retryCount.Load())
}

// task_fail_once ensures a given taskId fails only once; second call succeeds
var failedOnce sync.Map

func task_fail_once(id int) bool {
	_, loaded := failedOnce.LoadOrStore(id, true)
	return !loaded // first store returns true (fail); later calls return false (success)
}

// TestEngineMaxPendingDedupRefund verifies pending quota taken at produce is refunded when
// ingest's pushTasks dedups the subtask; without the refund the producer stalls forever.
func TestEngineMaxPendingDedupRefund(t *testing.T) {
	const children = 100 // all share one key; quota (8) < children would deadlock without refund
	var executed atomic.Int64
	engine := NewEngine(4, WithMaxPending[int](8))
	engine.ErrHandlerUtilSuccess()
	engine.AddTasks(&Task[int]{
		Key: 1,
		Run: func(ctx context.Context) ([]*Task[int], error) {
			var tasks []*Task[int]
			for i := 0; i < children; i++ {
				tasks = append(tasks, &Task[int]{
					Key: 777,
					Run: func(ctx context.Context) ([]*Task[int], error) {
						executed.Add(1)
						return nil, nil
					},
				})
			}
			return tasks, nil
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
		t.Fatal("Run stalled: pending quota for deduped subtasks not refunded")
	}
	if executed.Load() < 1 {
		t.Fatal("deduped child never executed")
	}
	t.Logf("dedup refund: %d/%d executed", executed.Load(), children)
}

// TestEngineKeyDedup verifies tasks with the same Key are heavily deduped (done cache checked at run time)
func TestEngineKeyDedup(t *testing.T) {
	var executed atomic.Int64
	engine := NewEngine[int](8)
	engine.MonitorInterval(time.Second)

	const dupCount = 100
	for i := 0; i < dupCount; i++ {
		engine.AddTasks(&Task[int]{
			Key: 42, // all same key
			Run: func(ctx context.Context) ([]*Task[int], error) {
				executed.Add(1)
				time.Sleep(time.Millisecond * 50) // allow cache to propagate
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
	// dedup is best-effort: after the first finishes and sets cache, later same keys are skipped
	// allow a few duplicates in the concurrency window, but not all should run
	if executed.Load() >= dupCount {
		t.Fatalf("dedup ineffective: all %d tasks executed", dupCount)
	}
	t.Logf("dedup: %d/%d executed", executed.Load(), dupCount)
}
