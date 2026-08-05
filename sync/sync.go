/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package sync

type noCopy struct{}

// Lock provides mutual exclusion.
func (*noCopy) Lock() {}

// Unlock provides mutual exclusion.
func (*noCopy) Unlock() {}
