/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"time"
)

type Type uint32

const (
	normalType Type = iota
	fixedType
)

type Worker[KEY Key] struct {
	id                      uint
	typ                     Type
	kind                    Kind
	taskCh                  chan *Task[KEY]
	createdAt               time.Time
	currentTask             *Task[KEY]
	isExecuting, canExecute bool
}

// workStatistics holds worker stats
type workStatistics struct {
	timeCost                                                                          time.Duration
	taskTotalCount, taskDoneCount, taskSkipCount, taskErrHandleCount, taskFailedCount uint64
	taskRepeatTimes, taskErrorTimes, taskTimeoutTimes                                 uint64
}

// EngineStatistics holds basic engine stats
type EngineStatistics struct {
	workStatistics
}
