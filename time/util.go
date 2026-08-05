/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import (
	"strconv"
	"time"
)

// UnixNano ...
func UnixNano(nsec int64) time.Time {
	return time.Unix(0, nsec)
}

var ZeroTime = time.Time{}

// StrToIntMonth ...
func StrToIntMonth(month string) int {
	var data = map[string]int{
		January:   1,
		February:  2,
		March:     3,
		April:     4,
		May:       5,
		June:      6,
		July:      7,
		August:    8,
		September: 9,
		October:   10,
		November:  11,
		December:  12,
	}
	return data[month]
}

// GetYMD ...
func GetYMD(time time.Time, sep string) string {
	year, month, day := time.Date()

	var monthStr string
	var dateStr string
	if month < 10 {
		monthStr = "0" + strconv.Itoa(int(month))
	} else {
		monthStr = strconv.Itoa(int(month))
	}

	if day < 10 {
		dateStr = "0" + strconv.Itoa(day)
	} else {
		dateStr = strconv.Itoa(day)
	}
	return strconv.Itoa(year) + sep + monthStr + sep + dateStr
}

// GetYM ...
func GetYM(time time.Time, sep string) string {
	year, month, _ := time.Date()

	var monthStr string
	if month < 10 {
		monthStr = "0" + strconv.Itoa(int(month))
	} else {
		monthStr = strconv.Itoa(int(month))
	}
	return strconv.Itoa(year) + sep + monthStr
}

// GetYesterdayYMD ...
func GetYesterdayYMD(sep string) string {
	return GetYM(time.Now().AddDate(0, 0, -1), sep)
}

// GetTomorrowYMD ...
func GetTomorrowYMD(sep string) string {
	return GetYM(time.Now().AddDate(0, 0, 1), sep)
}

// TodayZeroTime ...
func TodayZeroTime() time.Time {
	year, month, day := time.Now().Date()
	// now.Year(), now.Month(), now.Day() 是以本地时区为参照的年、月、日
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

// YesterdayZeroTime ...
func YesterdayZeroTime() time.Time {
	return TodayZeroTime().AddDate(0, 0, -1)
}
