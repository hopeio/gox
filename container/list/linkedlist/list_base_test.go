/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package linkedlist

import (
	"fmt"
	"github.com/hopeio/gox/log"
	"testing"
)

func TestList(t *testing.T) {
	// 打印链表信息
	var l BaseLinkedList[int]
	fmt.Println("###############################################")
	fmt.Println("链表长度为：", l.Len())
	fmt.Println("链表是否为空:", l.IsEmpty())
	fmt.Print("遍历链表：")
	l.traverse(func(data int) { log.Info(data, " ") })
	fmt.Println("###############################################")

}
