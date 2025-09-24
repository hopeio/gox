/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package ringqueue

import "testing"

func TestQueue(t *testing.T) {
	queue := New[int](6)
	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)
	queue.Enqueue(4)
	queue.Enqueue(5)
	t.Log(queue.LookAll())
	queue.Dequeue()
	queue.Dequeue()
	queue.Dequeue()
	t.Log(queue.LookAll())
	queue.Enqueue(6)
	t.Log(queue.LookAll())
	queue.Enqueue(7)
	t.Log(queue.LookAll())
	queue.Enqueue(8)
	t.Log(queue.LookAll())
	queue.Enqueue(9)
	t.Log(queue.LookAll())
	queue.Enqueue(10)
	t.Log(queue.LookAll())
	queue.Dequeue()
	t.Log(queue.LookAll())
	queue.Dequeue()
	t.Log(queue.LookAll())
	queue.Dequeue()
	t.Log(queue.LookAll())
	queue.Dequeue()
	t.Log(queue.LookAll())
	queue.Dequeue()
	t.Log(queue.LookAll())
	queue.Dequeue()
	t.Log(queue.LookAll())
}
