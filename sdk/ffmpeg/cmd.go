/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package ffmpeg

import (
	"github.com/hopeio/gox/log"
	execx "github.com/hopeio/gox/os/exec"
)

// doc: https://ffmpeg.org/ffmpeg-codecs.html
// https://ffmpeg.org/download.html

const CommonCmd = ` -i "%s" -y `

type Config struct {
	Path string
}

var ExecPath = "ffmpeg"

func Run(cmd string) error {
	cmd = ExecPath + " " + cmd
	log.Debug("exec:", cmd)
	err := execx.RunContainQuoted(cmd)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
