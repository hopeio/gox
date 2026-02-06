/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import "testing"

func TestLog(t *testing.T) {
	Info("test")
}

func TestLogStack(t *testing.T) {
	stackLogger.Debug("test")
}

func TestLogNoCaller(t *testing.T) {
	noCallerLogger.Debug("test")
}
