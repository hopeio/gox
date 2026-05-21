/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package linkedlist

import "testing"

func TestNew(t *testing.T) {
	l := New[int]()
	if l == nil {
		t.Fatal("New() returned nil")
	}
	if !l.IsEmpty() {
		t.Error("New list should be empty")
	}
	if l.Len() != 0 {
		t.Errorf("Len() = %d, want 0", l.Len())
	}
}

func TestLinkedList_Append(t *testing.T) {
	l := New[int]()
	l.Append(1)
	l.Append(2)
	l.Append(3)

	if l.Len() != 3 {
		t.Errorf("Len() = %d, want 3", l.Len())
	}
	if l.IsEmpty() {
		t.Error("List should not be empty after Append")
	}
}

func TestLinkedList_InsertHead(t *testing.T) {
	l := New[int]()
	l.InsertHead(1)
	l.InsertHead(2)
	l.InsertHead(3)

	if l.Len() != 3 {
		t.Errorf("Len() = %d, want 3", l.Len())
	}
}

func TestLinkedList_InsertAfterData(t *testing.T) {
	l := New[int]()
	l.Append(1)
	l.Append(3)
	err := l.InsertAfterData(1, 2)
	if err != nil {
		t.Errorf("InsertAfterData() error: %v", err)
	}
	if l.Len() != 3 {
		t.Errorf("Len() = %d, want 3", l.Len())
	}

	// Insert after non-existent data
	err = l.InsertAfterData(99, 4)
	if err == nil {
		t.Error("InsertAfterData() with non-existent data should return error")
	}
}

func TestLinkedList_Insert(t *testing.T) {
	l := New[int]()
	l.Append(1)
	l.Append(3)

	// Insert at position 1
	err := l.Insert(1, 2)
	if err != nil {
		t.Errorf("Insert() error: %v", err)
	}
	if l.Len() != 3 {
		t.Errorf("Len() = %d, want 3", l.Len())
	}

	// Insert at head
	err = l.Insert(0, 0)
	if err != nil {
		t.Errorf("Insert(0) error: %v", err)
	}

	// Insert at tail
	err = l.Insert(l.Len(), 4)
	if err != nil {
		t.Errorf("Insert at tail error: %v", err)
	}

	// Insert with negative position
	err = l.Insert(-1, 0)
	if err == nil {
		t.Error("Insert(-1) should return error")
	}

	// Insert beyond length
	err = l.Insert(100, 0)
	if err == nil {
		t.Error("Insert(100) should return error")
	}
}

func TestLinkedList_GetNode(t *testing.T) {
	l := New[int]()
	l.Append(1)
	l.Append(2)
	l.Append(3)

	node := l.GetNode(2)
	if node == nil {
		t.Fatal("GetNode(2) returned nil")
	}

	node = l.GetNode(99)
	if node != nil {
		t.Error("GetNode(99) should return nil for non-existent data")
	}
}

func TestLinkedList_Delete(t *testing.T) {
	l := New[int]()
	l.Append(1)
	l.Append(2)
	l.Append(3)

	l.Delete(2)
	if l.Len() != 2 {
		t.Errorf("Len() after Delete = %d, want 2", l.Len())
	}

	// Delete non-existent element
	l.Delete(99)
	if l.Len() != 2 {
		t.Errorf("Len() after deleting non-existent = %d, want 2", l.Len())
	}
}

func TestLinkedList_Traverse(t *testing.T) {
	l := New[int]()
	l.Append(1)
	l.Append(2)
	l.Append(3)

	var result []int
	l.traverse(func(v int) {
		result = append(result, v)
	})

	if len(result) != 3 {
		t.Errorf("traverse visited %d items, want 3", len(result))
	}
	if result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("traverse result = %v, want [1 2 3]", result)
	}
}

func TestLinkedList_Exist(t *testing.T) {
	l := New[int]()
	l.Append(1)
	l.Append(2)

	node := l.GetNode(1)
	if !l.Exist(node) {
		t.Error("Exist() should return true for existing node")
	}
}
