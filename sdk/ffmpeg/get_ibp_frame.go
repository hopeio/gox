/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package ffmpeg

import (
	"fmt"
	"path/filepath"

	"github.com/hopeio/gox/os/fs"
)

type Frame int

func (f Frame) String() string {
	switch f {
	case I:
		return "I"
	case P:
		return "P"
	case B:
		return "B"
	}
	return "unknown"
}

const (
	I Frame = iota
	P
	B
)

const GetFrameCmd = CommonCmd + `-vf "select=eq(pict_type\,%s)" -fps_mode vfr -qscale:v 2 -f image2 %s/%%03d.jpg`

func GetFrame(path, outPutDir string, f Frame) error {
	//cmd := `ffmpeg -i ` + path + ` -vf "select=eq(pict_type\,` + f.String() + `)" -vsync vfr -qscale:v 2 -f image2 ` + dst + `/%03d.jpg`
	dst := filepath.Clean(outPutDir)
	fs.Mkdir(dst)
	cmd := fmt.Sprintf(GetFrameCmd, path, f.String(), dst)
	return Run(cmd)
}
