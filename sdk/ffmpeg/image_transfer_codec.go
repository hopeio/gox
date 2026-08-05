/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package ffmpeg

import (
	"fmt"
	"github.com/hopeio/gox/sdk/mp4box"
	"strings"
)

// webp 无损模式
const ImgToWebpLosslessCmd = CommonCmd + `-c:v libwebp -lossless 1 -quality 100 -compression_level 6 "%s.webp"`

// ImgToWebpLossless ...
func ImgToWebpLossless(filePath, dst string) error {
	if strings.HasSuffix(dst, ".webp") {
		dst = dst[:len(dst)-5]
	}
	return Run(fmt.Sprintf(ImgToWebpLosslessCmd, filePath, dst))
}

const ImgToWebpCmd = CommonCmd + `-c:v libwebp -quality %d "%s.webp"`

// ImgToWebp ...
func ImgToWebp(filePath, dst string, quality int) error {
	if strings.HasSuffix(dst, ".webp") {
		dst = dst[:len(dst)-5]
	}
	return Run(fmt.Sprintf(ImgToWebpCmd, filePath, quality, dst))
}

const ImgToTAvifCmd = CommonCmd + `-c:v libaom-av1 -crf %d -cpu-used %d -row-mt 1 "%s.avif"`

// 多次压缩后avif会出现明显色差,比webp略好
// -cpu-used 3 会加速，但是图片大小会变大,质量变差,<=3比较好,推荐2
// More encoding options are available: -b 700k -tile-columns 600 -tile-rows 800 - example for the bitrate and tales.

// ImgToAvif ...
func ImgToAvif(filePath, dst string, crf, cpuUsed int) error {
	if strings.HasSuffix(dst, ".avif") {
		dst = dst[:len(dst)-5]
	}
	return Run(fmt.Sprintf(ImgToTAvifCmd, filePath, crf, cpuUsed, dst))
}

const ImgToHeicCmd = CommonCmd + `-crf 20 -c:v libx265 -preset veryslow %s.mp4`
const ImgToHeicCmd2 = CommonCmd + `-hide_banner -r 1 -vf "scale=trunc(iw/2)*2:trunc(ih/2)*2,zscale=m=170m:r=pc" -pix_fmt yuv420p -frames 1 -c:v libx265 -preset veryslow -crf 20 -x265-params range=full:colorprim=smpte170m "%s.hevc"`
const ImgToHeicCmd3 = CommonCmd + `-hide_banner -r 1 -vf "scale=trunc(iw/2)*2:trunc(ih/2)*2,zscale=m=170m:r=pc" -pix_fmt yuv420p -frames 1 -c:v libx265 -preset veryslow -crf 20 -x265-params range=full:colorprim=smpte170m:aq-strength=1.2 -deblock -2:-2 "%s.hevc"
`

// ImgToHeic ...
func ImgToHeic(filePath, dst string) error {
	if strings.HasSuffix(dst, ".heic") {
		dst = dst[:len(dst)-5]
	}
	err := Run(fmt.Sprintf(ImgToHeicCmd, filePath, dst))
	if err != nil {
		return err
	}

	return mp4box.Heic(dst+".mp4", dst)
}

const ImgToJxlCmd = CommonCmd + `-c:v libjxl "%s.jxl"`

// ImgToJxl ...
func ImgToJxl(filePath, dst string) error {
	if strings.HasSuffix(dst, ".jxl") {
		dst = dst[:len(dst)-4]
	}

	return Run(fmt.Sprintf(ImgToJxlCmd, filePath, dst))
}
