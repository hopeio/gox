/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package qyweixin

import (
	"errors"
	"fmt"
	"github.com/hopeio/gox/net/http/client"
	"strings"
)

const (
	//VERSION is SDK version
	VERSION = "0.1"

	//ROOT is the root url
	ROOT = "https://qyapi.weixin.qq.com/cgi-bin/webhook/"
)

// RobotSendMessage performs the operation.
func RobotSendMessage(key string, msg MessageType) error {
	signUrl, err := RobotUrl(key)
	if err != nil {
		return err
	}
	body := strings.NewReader(Format(msg))

	return client.Post(ROOT+signUrl, body, nil)
}

// RobotUrl performs the operation.
func RobotUrl(key string) (string, error) {
	if key == "" {
		return "", errors.New("key不能为为空")
	}

	return fmt.Sprintf("send?key=%s", key), nil
}

// RobotSendMarkDownMessage can send a text message to a group chat
func RobotSendMarkDownMessage(key string, content string) error {
	signUrl, err := RobotUrl(key)
	if err != nil {
		return err
	}
	body := strings.NewReader(MarkdownMessage(content))

	return client.Post(ROOT+signUrl, body, nil)
}

// RobotSendTextMessage performs the operation.
func RobotSendTextMessage(key, content string, mobiles []string, user ...string) error {
	msg := &Text{
		Content:             content,
		MentionedList:       user,
		MentionedMobileList: mobiles,
	}
	return RobotSendMessage(key, msg)
}

type Robot struct {
	Key string
}

// SendMessage performs the operation.
func (r *Robot) SendMessage(msg MessageType) error {
	return RobotSendMessage(r.Key, msg)
}
