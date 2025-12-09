/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package watch

import "time"

type Callback struct {
	LastModTime time.Time
	Callbacks   [5]func(string)
}

type Handlers map[string]*Callback

// Op describes a set of file operations.
type Op uint32

const (
	// A new pathname was created.
	Create Op = 1 << iota

	// The pathname was written to; this does *not* mean the write has finished,
	// and a write can be followed by more writes.
	Write

	// The path was removed; any watches on it will be removed. Some "remove"
	// operations may trigger a Rename if the file is actually moved (for
	// example "remove to trash" is often a rename).
	Remove

	// The path was renamed to something else; any watched on it will be
	// removed.
	Rename

	// File attributes were changed.
	//
	// It's generally not recommended to take action on this event, as it may
	// get triggered very frequently by some software. For example, Spotlight
	// indexing on macOS, anti-virus software, backup software, etc.
	Chmod
)
