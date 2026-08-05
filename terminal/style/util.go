/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package style

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

var supportLevel int

// init initializes package state.
func init() {
	if EnableColor() {
		colorDepth, _ := ColorDepth()
		if runtime.GOOS == "windows" || colorDepth >= 256 {
			supportLevel = 3
		} else if colorDepth == 16 {
			supportLevel = 2
		} else {
			supportLevel = 1
		}
	}
}

// EnableColor reports whether the condition holds.
func EnableColor() bool {
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}
	forceColor := os.Getenv("FORCE_COLOR")
	if forceColor == "0" {
		return false
	}
	term := os.Getenv("TERM")
	if term == "dump" {
		return false
	}
	return true
}

// ColorDepth performs the operation.
func ColorDepth() (int, error) {
	colorTerm := os.Getenv("COLORTERM")
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return 1<<24 - 1, nil
	}

	term := os.Getenv("TERM")
	if strings.HasSuffix(term, "-256color") || strings.HasSuffix(term, "256") {
		return 256, nil
	} else if strings.Contains(term, "color") {
		return 16, nil
	}

	cmd := exec.Command("tput", "colors")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	colors, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}
	return colors, nil
}
