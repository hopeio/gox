/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package terminal

import (
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

// IsTerminal reports whether the condition holds.
func IsTerminal() bool {
	if _, exists := os.LookupEnv("TERM"); exists {
		return true
	}
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		return true
	}
	if terminal.IsTerminal(int(os.Stderr.Fd())) {
		return true
	}
	return false
}
