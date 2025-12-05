/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package reqctx

import (
	"net/url"
	"strconv"
)

type DeviceInfo struct {
	//设备
	Device    string  `json:"device" gorm:"size:255"`
	OS        string  `json:"os" gorm:"size:255"`
	AppCode   string  `json:"appCode" gorm:"size:255"`
	AppVer    string  `json:"appVer" gorm:"size:255"`
	IP        string  `json:"ip" gorm:"size:255"`
	Lng       float64 `json:"lng" gorm:"type:numeric(10,6)"`
	Lat       float64 `json:"lat" gorm:"type:numeric(10,6)"`
	Area      string  `json:"area" gorm:"size:255"`
	UserAgent string  `json:"userAgent" gorm:"size:255"`
}

// info: device,os,appCode,appVersion
// area:xxx
// location:1.23456,2.123456
func Device(infoHeader, area, location, userAgent, ip string) *DeviceInfo {
	unknow := true
	var info DeviceInfo
	//Device-Info:device,osInfo,appCode,appVersion
	if infoHeader != "" {
		unknow = false
		var n, m int
		for i, c := range infoHeader {
			if c == ',' {
				switch n {
				case 0:
					info.Device = infoHeader[m:i]
				case 1:
					info.OS = infoHeader[m:i]
				case 2:
					info.AppCode = infoHeader[m:i]
				case 3:
					info.AppVer = infoHeader[m:i]
				}
				m = i + 1
				n++
			}
		}
	}
	// area:xxx
	// location:1.23456,2.123456
	if area != "" {
		unknow = false
		info.Area, _ = url.PathUnescape(area)
	}
	if location != "" {
		unknow = false
		var n, m int
		for i, c := range location {
			if c == ',' {
				switch n {
				case 0:
					info.Lng, _ = strconv.ParseFloat(location[m:i], 64)
				case 1:
					info.Lat, _ = strconv.ParseFloat(location[m:i], 64)
				}
				m = i + 1
				n++
			}
		}

	}

	if userAgent != "" {
		unknow = false
		info.UserAgent = userAgent
	}
	if ip != "" {
		unknow = false
		info.IP = ip
	}
	if unknow {
		return nil
	}
	return &info
}
