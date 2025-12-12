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

// DeviceInfo device info
type DeviceInfo struct {
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

// Device get device info
// device: device,os
// app: appCode,appVersion
// area: xxx
// location: 1.23456,2.123456
func Device(device, app, area, location, userAgent, ip string) *DeviceInfo {
	info := new(DeviceInfo)
	unknow := true
	//Device:device,osInfo
	if device != "" {
		unknow = false
		var n, m int
		for i, c := range device {
			if c == ',' {
				switch n {
				case 0:
					info.Device = device[m:i]
				case 1:
					info.OS = device[m:i]
				}
				m = i + 1
				n++
			}
		}
	}
	// App:appCode,appVersion
	if app != "" {
		unknow = false
		var n, m int
		for i, c := range app {
			if c == ',' {
				switch n {
				case 0:
					info.AppCode = app[m:i]
				case 1:
					info.AppVer = app[m:i]
				}
				m = i + 1
				n++
			}
		}
	}
	// area:xxx
	if area != "" {
		unknow = false
		info.Area, _ = url.PathUnescape(area)
	}
	// location:1.23456,2.123456
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
	return info
}
