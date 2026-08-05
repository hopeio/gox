/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package aop

import (
	"log"
	"testing"
)

var foo1 = func() {
	log.Println("foo1")
}

func foo2() {
	log.Println("foo1")
}
func before() { log.Println("before") }
func after()  { log.Println("after") }

func TestAop(t *testing.T) {
	Invoke(before, &foo1, after)
	foo1()

	// aop() depends on gomonkey and is marked unavailable; skip to avoid stack overflow.
	t.Skip("aop() via gomonkey is unsupported")
	log.Println("----------------------------------------")
	aop(before, foo2, after)
	foo2()
}
