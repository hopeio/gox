/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import "time"

const (
	LayoutTimeMacro   = "2006-01-02 15:04:05.999999"
	LayoutCompactDate = "20060102"
	LayoutCompactTime = "20060102150405"
	LayoutRFC1        = "2006/01/02 - 15:04:05"
)

const (
	DayEndTime            = "23:59:59"
	DayEndTimeWithSpace   = " 23:59:59"
	DayBeginTime          = "00:00:00"
	DayBeginTimeWithSpace = " 00:00:00"
)

const (
	Day             = time.Hour * 24
	MonthDay30      = Day * 30
	MonthDay31      = Day * 31
	MonthDay28      = Day * 28
	MonthDay29      = Day * 29
	Month           = MonthDay30
	YearDay365      = Day * 365
	YearDay366      = Day * 366
	Year            = YearDay365
	SecondsOfDay    = HoursOfDay * SecondsOfHour
	SecondsOfMinute = 60
	SecondsOfHour   = MinutesOfHour * SecondsOfMinute
	MinutesOfHour   = 60
	MinutesOfDay    = MinutesOfHour * HoursOfDay
	HoursOfDay      = 24
)

const (
	January   = "January"
	February  = "February"
	March     = "March"
	April     = "April"
	May       = "May"
	June      = "June"
	July      = "July"
	August    = "August"
	September = "September"
	October   = "October"
	November  = "November"
	December  = "December"
)

const (
	Monday    = "Monday"
	Tuesday   = "Tuesday"
	Wednesday = "Wednesday"
	Thursday  = "Thursday"
	Friday    = "Friday"
	Saturday  = "Saturday"
	Sunday    = "Sunday"
)
