/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package video

import (
	"bytes"
	"encoding/binary"
	"os"
)

// BoxHeader is the box header
type BoxHeader struct {
	Size       uint32
	FourccType [4]byte
	Size64     uint64
}

// GetMP4Duration returns the value.
func GetMP4Duration(filepath string) (lengthOfTime uint32, err error) {
	reader, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	var info = make([]byte, 0x10)
	var boxHeader BoxHeader
	var offset int64 = 0
	// Get the moov box offset
	for {
		_, err = reader.ReadAt(info, offset)
		if err != nil {
			return
		}
		boxHeader = getHeaderBoxInfo(info)
		fourccType := getFourccType(boxHeader)
		if fourccType == "moov" {
			break
		}
		// Some MP4 mdat sizes are oversized and need special handling
		if fourccType == "mdat" {
			if boxHeader.Size == 1 {
				offset += int64(boxHeader.Size64)
				continue
			}
		}
		offset += int64(boxHeader.Size)
	}
	// Read the beginning of the moov box
	moovStartBytes := make([]byte, 0x100)
	_, err = reader.ReadAt(moovStartBytes, offset)
	if err != nil {
		return
	}
	// Define timeScale and Duration offsets
	timeScaleOffset := 0x1C
	durationOffest := 0x20
	timeScale := binary.BigEndian.Uint32(moovStartBytes[timeScaleOffset : timeScaleOffset+4])
	Duration := binary.BigEndian.Uint32(moovStartBytes[durationOffest : durationOffest+4])
	lengthOfTime = Duration / timeScale
	return
}

// getHeaderBoxInfo performs the operation.
func getHeaderBoxInfo(data []byte) (boxHeader BoxHeader) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &boxHeader)
	return
}

// getFourccType performs the operation.
func getFourccType(boxHeader BoxHeader) (fourccType string) {
	fourccType = string(boxHeader.FourccType[:])
	return
}
