/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"testing"
	"time"
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
	go func() {
		for {
			engine.Run(genTask3("a", int(time.Now().Unix())))
			time.Sleep(time.Second)
		}
	}()

	for {
		engine.Run(genTask3("b", int(time.Now().UnixMilli())))
		time.Sleep(time.Second * 2)
	}
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
