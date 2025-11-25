/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package dingtalk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hopeio/gox/net/http/client"
)

const (
	//VERSION is SDK version
	VERSION = "0.1"

	//ROOT is the root url
	ROOT = "https://oapi.dingtalk.com/"
)

func RobotSendMessage(accessToken, secret string, msg MessageType) error {
	signUrl, err := RobotUrl(accessToken, secret)
	if err != nil {
		return err
	}
	body := strings.NewReader(Format(msg))

	return client.Post(ROOT+signUrl, body, nil)
}

func RobotUrl(accessToken, secret string) (string, error) {
	if accessToken == "" {
		return "", errors.New("token不能为为空")
	}
	if secret != "" {
		// 密钥加签处理
		now := time.Now().UnixNano() / int64(time.Millisecond)
		timestampStr := strconv.FormatInt(now, 10)
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(timestampStr + "\n" + secret))
		sum := h.Sum(nil)
		return fmt.Sprintf("robot/send?access_token=%s&timestamp=%s&sign=%s", accessToken, timestampStr, url.QueryEscape(base64.StdEncoding.EncodeToString(sum))), nil
	}
	return fmt.Sprintf("robot/send?access_token=%s", accessToken), nil
}

// RobotSendTextMessage can send a text message to a group chat
func RobotSendTextMessage(accessToken string, content string) error {
	return RobotSendTextMessageWithSecret(accessToken, "", content)
}

func RobotSendTextMessageWithSecret(accessToken, secret, content string) error {
	signUrl, err := RobotUrl(accessToken, secret)
	if err != nil {
		return err
	}
	body := strings.NewReader(TextMessage(content))

	return client.Post(ROOT+signUrl, body, nil)
}

func RobotSendMarkDownMessage(token, title, content string, at *At) error {
	msg := &Markdown{
		Title: title,
		Text:  content,
		At:    at,
	}
	return RobotSendMessage(token, "", msg)
}

func RobotSendMarkDownMessageWithSecret(token, secret, title, content string, at *At) error {
	msg := &Markdown{
		Title: title,
		Text:  content,
		At:    at,
	}
	return RobotSendMessage(token, secret, msg)
}

type Robot struct {
	AccessToken string
	Secret      string
}

func (r *Robot) SendMessage(msg MessageType) error {
	return RobotSendMessage(r.AccessToken, r.Secret, msg)
}
