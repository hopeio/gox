/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package mp4box

import (
	execi "github.com/hopeio/gox/os/exec"
)

// https://www.videohelp.com/software/MP4Box
const Mp4BoxCmd = `mp4box -add-image %s.hevc:primary -ab heic -new %s.heic`

// Heic performs the operation.
func Heic(filePath, dst string) error {
	_, err := execi.RunGetOut(Mp4BoxCmd)
	return err
}
