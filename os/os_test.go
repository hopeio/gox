/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package os

import "testing"

func TestHostname(t *testing.T) {
	hostname := Hostname()
	if hostname == "" {
		t.Error("Hostname() returned empty string")
	}
}

func TestSplit_Simple(t *testing.T) {
	result := Split("echo hello world")
	if len(result) != 3 {
		t.Fatalf("Split() = %v, want 3 elements", result)
	}
	if result[0] != "echo" {
		t.Errorf("result[0] = %q, want 'echo'", result[0])
	}
}

func TestSplit_Quoted(t *testing.T) {
	result := Split(`echo "hello world"`)
	if len(result) != 2 {
		t.Fatalf("Split() = %v, want 2 elements", result)
	}
	if result[0] != "echo" {
		t.Errorf("result[0] = %q, want 'echo'", result[0])
	}
	if result[1] != "hello world" {
		t.Errorf("result[1] = %q, want 'hello world'", result[1])
	}
}
