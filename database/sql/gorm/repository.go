/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gorm

import (
	"github.com/hopeio/gox/database/sql/gorm/scope"
	"gorm.io/gorm"
)

type Repository[T any] struct {
	*gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db}
}

func (r *Repository[T]) Create(t *T) error {
	return r.DB.Create(t).Error
}

func (r *Repository[T]) Retrieve(id any) (*T, error) {
	var t T
	err := r.DB.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository[T]) Update(t *T) error {
	return r.DB.Updates(&t).Error
}

func (r *Repository[T]) Delete(id any) error {
	var t T
	return r.DB.Delete(&t, id).Error
}

type ChainRepository[T any] struct {
	scope.ChainScope
	*gorm.DB
}

func NewChainRepository[T any](db *gorm.DB) *ChainRepository[T] {
	return &ChainRepository[T]{
		DB: db,
	}
}
