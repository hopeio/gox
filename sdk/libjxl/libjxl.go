/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package libjxl

import (
	"fmt"
	execi "github.com/hopeio/gox/os/exec"
	"strings"
)

// https://github.com/libjxl/libjxl/releases
// windows support: https://github.com/saschanaz/jxl-winthumb/releases administrator regsvr32 jxl_winthumb.dll
const ImgToJxlCmd = `cjxl %s %s.jxl -q %d --lossless_jpeg=0`
const JxlImgToOtherCmd = `djxl %s %s`

// ImgToJxl ...
func ImgToJxl(filePath, dst string, quality int) error {
	if strings.HasSuffix(dst, ".jxl") {
		dst = dst[:len(dst)-4]
	}
	_, err := execi.RunGetOut(fmt.Sprintf(ImgToJxlCmd, filePath, dst, quality))
	return err
}
