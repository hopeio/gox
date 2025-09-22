/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package parallel

import (
	"github.com/hopeio/gox/types"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
)

func RunIgnoreError(tasks []types.FuncReturnErr) error {
	ch := make(chan error)
	for _, task := range tasks {
		go func() {
			ch <- task()
		}()
	}
	var errs error
	for err := range ch {
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return errs
}

func Run(tasks []types.FuncReturnErr) error {
	var group errgroup.Group
	for _, task := range tasks {
		group.Go(task)
	}
	return group.Wait()
}

func RunReturnData[T any](tasks []types.FuncReturnDataOrErr[T]) ([]T, error) {
	var group errgroup.Group
	ret := make([]T, len(tasks))
	for i, task := range tasks {
		group.Go(func() error {
			data, err := task()
			ret[i] = data
			return err
		})
	}
	return ret, group.Wait()
}
