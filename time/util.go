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

// UnixNano converts a Unix nanosecond timestamp to time.Time.
func UnixNano(nsec int64) time.Time {
	return time.Unix(0, nsec)
}

var ZeroTime = time.Time{}

// StrToIntMonth converts a month name to its numeric value.
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

// GetYMD formats the date as YYYY{sep}MM{sep}DD.
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

// GetYM formats the date as YYYY{sep}MM.
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

// GetYesterdayYMD formats yesterday as YYYY{sep}MM{sep}DD.
func GetYesterdayYMD(sep string) string {
	return GetYM(time.Now().AddDate(0, 0, -1), sep)
}

// GetTomorrowYMD formats tomorrow as YYYY{sep}MM{sep}DD.
func GetTomorrowYMD(sep string) string {
	return GetYM(time.Now().AddDate(0, 0, 1), sep)
}

// TodayZeroTime returns today's date in the local time zone with time set to 00:00:00.
func TodayZeroTime() time.Time {
	year, month, day := time.Now().Date()
	// now.Year(), now.Month(), now.Day() are based on the local time zone.
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

// YesterdayZeroTime returns the local date for yesterday at 00:00:00.
func YesterdayZeroTime() time.Time {
	return TodayZeroTime().AddDate(0, 0, -1)
}
