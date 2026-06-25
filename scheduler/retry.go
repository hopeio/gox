/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import "go.uber.org/multierr"


type Retrier interface {
	Do(times uint) (retry bool)
}


func RetryRunTimes(times int, f func(int) error) error {
	var errs error
	for i := 0; i < times; i++ {
		err := f(i)
		if err == nil {
			return nil
		}
		errs = multierr.Append(errs, err)
	}

	return errs
}

func RetryRun(f func(int) bool) {
	for i := 0; ; i++ {
		if !f(i) {
			break
		}
	}
}
