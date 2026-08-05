/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package ffmpeg

import (
	"fmt"
)

type PerSet string

const (
	Ultrafast PerSet = "ultrafast"
	SuperFast PerSet = "superfast"
	VeryFast  PerSet = "veryfast"
	Faster    PerSet = "faster"
	Fast      PerSet = "fast"
	Medium    PerSet = "medium"
	Slow      PerSet = "slow"
	Slower    PerSet = "slower"
	VerySlow  PerSet = "veryslow"
	Placebo   PerSet = "placebo"
)

const param = "-global_quality 20"

const H264ToH265ByIntelGPUCmd = `ffmpeg -hwaccel qsv -c:v h264_qsv -i %s -c:v hevc_qsv -preset veryslow -g 60 -gpu_copy 1 -c:a copy "%s"`

const cmd1 = `preset=veryslow,profile=main,look_ahead=1,global_quality=18`

// H264ToH265ByIntelGPU ...
func H264ToH265ByIntelGPU(filePath, dst string) error {
	return Run(fmt.Sprintf(H264ToH265ByIntelGPUCmd, filePath, dst))
}

// libaom-av1
const ToAv1Libaomav1Cmd = CommonCmd + `-c:v libaom-av1 -crf %d -cpu-used %d -row-mt 1 -y "%s"`

// ToAV1ByLibaomav1 ...
func ToAV1ByLibaomav1(filePath, dst string, crf, cpuUsed int) error {
	return Run(fmt.Sprintf(ToAv1Libaomav1Cmd, filePath, crf, cpuUsed, dst))
}

// libsvtav1
// librav1e

// libx264
const ToH264Cmd = CommonCmd + `-c:v libx264 -profile high -preset %s -crf %d -y "%s"`

// ToH264ByXlib264 ...
func ToH264ByXlib264(filePath, dst string, crf int, perset PerSet) error {
	return Run(fmt.Sprintf(ToH264Cmd, filePath, perset, crf, dst))
}

// libvpx

// libx265
const ToH265Cmd = CommonCmd + `-c:v libx265 -preset %s -crf %d -y "%s"`

// ToH265ByXlib265 ...
func ToH265ByXlib265(filePath, dst string, crf int, perset PerSet) error {
	return Run(fmt.Sprintf(ToH265Cmd, filePath, perset, crf, dst))
}

// libvpx
