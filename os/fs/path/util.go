/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package path

// 该文件仅供示例

import (
	stringsx "github.com/hopeio/gox/strings"
	timex "github.com/hopeio/gox/time"
	"os"
	"strconv"
	"strings"
	"time"
)

const PathSeparator = string(os.PathSeparator)

type Path interface {
	Path() string
}

// userId/year/time_userId_key_filename
type ByUId struct {
	UserId    int       `json:"userId"`
	UserIdStr string    `json:"userIdStr"`
	Id        int       `json:"id"`
	IdStr     string    `json:"idStr"`
	Time      time.Time `json:"pubAt" gorm:"type:timestamptz(0);default:0001-01-01 00:00:00"`
	TimeStr   string    `json:"pubAtStr" comment:"20230321"`
	FileName  string    `json:"fileName"`
}

func (d *ByUId) PreHandle() {
	if d.IdStr == "" {
		d.IdStr = strconv.Itoa(d.Id)
	}
	if d.TimeStr == "" {
		d.TimeStr = d.Time.Format(timex.LayoutCompactTime)
	}
	if d.UserIdStr == "" {
		d.UserIdStr = strconv.Itoa(d.UserId)
	}
	d.TimeStr = stringsx.RemoveRunes(d.TimeStr, '-', ' ', ':')
}

func (d *ByUId) Path() string {
	d.PreHandle()
	filepath := strings.Join([]string{d.UserIdStr, d.TimeStr[:4], strings.Join([]string{d.TimeStr, d.UserIdStr, d.IdStr, d.FileName}, "_")}, PathSeparator)
	return filepath
}

// year/key_date_title/filename
type ById struct {
	Id    int    `json:"id"`
	IdStr string `json:"idStr"`
	Title string
	//PrePath   string    `json:"prePath" comment:""`
	TimeStr  string
	FileName string `json:"fileName"`
}

func (d *ById) Path() string {
	if d.IdStr == "" {
		d.IdStr = strconv.Itoa(d.Id)
	}
	filepath := strings.Join([]string{d.TimeStr[:4], d.IdStr + "_" + d.TimeStr + "_" + d.Title, d.FileName}, "/")
	return filepath
}

// path
type ByPath string

func (d ByPath) Path() string {
	return string(d)
}

// userId/date_key_title/time_userId_key_filename
type ByUIdTitle struct {
	ByUId
	Title string
	//PrePath   string    `json:"prePath" comment:""`
}

func (d *ByUIdTitle) Path() string {
	d.PreHandle()
	filepath := strings.Join([]string{d.UserIdStr, d.TimeStr + "_" + d.Title + "_" + d.IdStr, strings.Join([]string{d.TimeStr, d.UserIdStr, d.IdStr, d.FileName}, "_")}, "/")
	return filepath
}
