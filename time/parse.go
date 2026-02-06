/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import (
	"fmt"
	"time"
)

func FormatRelativeTime(fromTime time.Time) string {
	now := time.Now()
	duration := now.Sub(fromTime)

	days := int(duration.Hours() / 24)
	weeks := days / 7
	months := int(duration.Hours() / (24 * 30)) // 简化计算，实际月份天数有变化
	years := months / 12

	switch {
	case duration.Minutes() < 1:
		return "刚刚"
	case duration.Hours() < 1:
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	case days < 1:
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	case days < 7:
		return fmt.Sprintf("%d天前", days)
	case weeks < 1:
		return fmt.Sprintf("%d周前", weeks)
	case months < 1:
		return fmt.Sprintf("%d个月前", months)
	default:
		return fmt.Sprintf("%d年前", years)
	}
}

func ParseTime(t string) (time.Time, error) {
	return time.Parse(time.TimeOnly, t)
}

func ParseDateTime(t string) (time.Time, error) {
	return time.Parse(time.DateTime, t)
}

func ParseDate(t string) (time.Time, error) {
	return time.Parse(time.DateOnly, t)
}
