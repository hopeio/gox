/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/hopeio/gox/log"
)

type Kind uint32

const (
	KindNormal = iota
)

var (
	stdTimeout time.Duration = 0
)

type execLog struct {
	execBeginAt time.Time
	execEndAt   time.Time
	err         error
}

type TaskStatistics struct {
	reExecTimes int
	errTimes    int
}

// ReExecTimes returns the result.
func (t *TaskStatistics) ReExecTimes() int {
	return t.reExecTimes
}

// ErrTimes returns the result.
func (t *TaskStatistics) ErrTimes() int {
	return t.errTimes
}

type Task[KEY Key] struct {
	ctx      context.Context
	Kind     Kind
	Key      KEY
	Priority int
	Describe string
	TaskStatistics
	Run       TaskFunc[KEY]
	id        uint64
	createdAt time.Time
	execLog
	reExecLogs []*execLog // most tasks run only once
	deadline   time.Time
	timeout    time.Duration
}

// NewTask creates and returns a new instance.
func NewTask[KEY Key](task TaskFunc[KEY]) *Task[KEY] {
	return &Task[KEY]{
		Run: task,
	}
}

// SetContext updates or inserts a value.
func (t *Task[KEY]) SetContext(ctx context.Context) *Task[KEY] {
	t.ctx = ctx
	return t
}

// SetPriority updates or inserts a value.
func (t *Task[KEY]) SetPriority(priority int) *Task[KEY] {
	t.Priority = priority
	return t
}

// SetKind updates or inserts a value.
func (t *Task[KEY]) SetKind(k Kind) *Task[KEY] {
	t.Kind = k
	return t
}

// SetKey updates or inserts a value.
func (t *Task[KEY]) SetKey(key KEY) *Task[KEY] {
	t.Key = key
	return t
}

// SetDescribe updates or inserts a value.
func (t *Task[KEY]) SetDescribe(describe string) *Task[KEY] {
	t.Describe = describe
	return t
}

// Id returns the result.
func (t *Task[KEY]) Id() uint64 {
	return t.id
}

// Compare compares values.
func (t *Task[KEY]) Compare(t2 *Task[KEY]) int {
	return t.Priority - t2.Priority
}

// Errs returns the result.
func (t *Task[KEY]) Errs() []error {
	var errs []error
	if t.err != nil {
		errs = append(errs, t.err)
	}
	for _, log := range t.reExecLogs {
		errs = append(errs, log.err)
	}
	return errs
}

// ErrLog performs the operation.
func (t *Task[KEY]) ErrLog() {
	builder := strings.Builder{}
	if t.err != nil {
		builder.WriteString("[1]{")
		builder.WriteString(t.err.Error())
		builder.WriteString("}\n")
	}
	for i, log := range t.reExecLogs {
		if log.err != nil {
			builder.WriteString("[" + strconv.Itoa(i+2) + "]{")
			builder.WriteString(log.err.Error())
			builder.WriteString("}\n")
		}
	}
	log.Error(builder.String())
}

type TaskRun[KEY Key] interface {
	Run(ctx context.Context) ([]*Task[KEY], error)
}

type TasDo[KEY Key] interface {
	Do(ctx context.Context) ([]*Task[KEY], error)
}

type TasExec[KEY Key] interface {
	Exec(ctx context.Context) ([]*Task[KEY], error)
}

type Tasks[KEY Key] []*Task[KEY]

// Less compares values.
func (tasks Tasks[KEY]) Less(i, j int) bool {
	return tasks[i].Priority > tasks[j].Priority
}

// ---------------

type ErrHandle func(context.Context, error)

type TaskFunc[KEY Key] func(ctx context.Context) ([]*Task[KEY], error)

// Run executes the operation.
func (t TaskFunc[KEY]) Run(ctx context.Context) ([]*Task[KEY], error) {
	return t(ctx)
}

// Do executes the operation.
func (t TaskFunc[KEY]) Do(ctx context.Context) ([]*Task[KEY], error) {
	return t(ctx)
}

// Exec performs the operation.
func (t TaskFunc[KEY]) Exec(ctx context.Context) ([]*Task[KEY], error) {
	return t(ctx)
}

// emptyTaskFunc performs the operation.
func emptyTaskFunc[KEY Key](ctx context.Context) ([]*Task[KEY], error) {
	return nil, nil
}
