/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package aop

import (
	"log"

	"bou.ke/monkey"
	_ "github.com/agiledragon/gomonkey/v2" // todo
)

func init() {
	monkey.Patch(NewDao, func() Dao {
		return &UserProxy{&User{}}
	})
}

type Dao interface {
	Save()
	Update()
}

type User struct {
}

//go:noinline
func NewDao() Dao {
	return &User{}
}

func (u *User) Save() {

}
func (u *User) Update() {

}

type UserProxy struct {
	user *User
}

func (u *UserProxy) Save() {
	log.Println("save")
	u.user.Save()
}
func (u *UserProxy) Update() {
	log.Println("update")
	u.user.Update()
}

type Proxy struct {
	p interface{}
}
