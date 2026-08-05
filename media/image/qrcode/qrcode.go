/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package qrcode

import (
	"crypto/md5"
	"encoding/hex"
	"image/jpeg"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

type QrCode struct {
	URL    string
	Width  int
	Height int
	Ext    string
	Level  qr.ErrorCorrectionLevel
	Mode   qr.Encoding
}

const (
	EXT_JPG = ".jpg"
)

// NewQrCode creates and returns a new instance.
func NewQrCode(url string, width, height int, level qr.ErrorCorrectionLevel, mode qr.Encoding) *QrCode {
	return &QrCode{
		URL:    url,
		Width:  width,
		Height: height,
		Level:  level,
		Mode:   mode,
		Ext:    EXT_JPG,
	}
}

/*func GetQrCodePath() string {
	return initialize.Config.Server.QrCodeSavePath
}

func GetQrCodeFullPath() string {
	return initialize.Config.Server.RuntimeRootPath + initialize.Config.Server.QrCodeSavePath
}

func GetQrCodeFullUrl(name string) string {
	return initialize.Config.Server.PrefixUrl + "/" + GetQrCodePath() + name
}*/

// GetQrCodeFileName returns the value.
func GetQrCodeFileName(value string) string {
	m := md5.New()
	m.Write([]byte(value))
	return hex.EncodeToString(m.Sum(nil))
}

// GetQrCodeExt returns the value.
func (q *QrCode) GetQrCodeExt() string {
	return q.Ext
}

// CheckEncode reports whether the condition holds.
func (q *QrCode) CheckEncode(path string) bool {
	src := path + GetQrCodeFileName(q.URL) + q.GetQrCodeExt()
	_, err := os.Stat(src)

	return !os.IsNotExist(err)
}

// Encode formats or converts the value.
func (q *QrCode) Encode(path string) (string, string, error) {
	name := GetQrCodeFileName(q.URL) + q.GetQrCodeExt()
	src := path + name
	_, err := os.Stat(src)
	if os.IsNotExist(err) == true {
		code, err := qr.Encode(q.URL, q.Level, q.Mode)
		if err != nil {
			return "", "", err
		}

		code, err = barcode.Scale(code, q.Width, q.Height)
		if err != nil {
			return "", "", err
		}

		f, err := os.Open(path + name)
		if err != nil {
			return "", "", err
		}
		defer f.Close()

		err = jpeg.Encode(f, code, nil)
		if err != nil {
			return "", "", err
		}
	}

	return name, path, nil
}
