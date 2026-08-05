/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package dingtalk

import (
	"strconv"
	"strings"

	jsonx "github.com/hopeio/gox/encoding/json"
)

const (
	MsgTypeTmpl = `{"msgtype":"%s",%s}`
)

type MessageType interface {
	MessageType() MsgType
}

type RobotConfig struct {
	Token  string
	Secret string
}

type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	At    *At    `json:"at,omitempty"`
}

// MessageType returns the result.
func (*Markdown) MessageType() MsgType {
	return MsgTypeMarkdown
}

type Text struct {
	Content string `json:"content"`
}

// MessageType returns the result.
func (Text) MessageType() MsgType {
	return MsgTypeText
}

type At struct {
	AtMobiles []string `json:"atMobiles"`
	AtUserIds []int    `json:"atUserIds"`
	IsAtAll   bool     `json:"isAtAll"`
}

type Link struct {
	Text       string `json:"text"`
	Title      string `json:"title"`
	PicUrl     string `json:"picUrl"`
	MessageUrl string `json:"messageUrl"`
}

// MessageType returns the result.
func (*Link) MessageType() MsgType {
	return MsgTypeLink
}

type ActionCard struct {
	Title          string `json:"title"`
	Text           string `json:"text"`
	BtnOrientation string `json:"btnOrientation"`
	SingleTitle    string `json:"singleTitle"`
	SingleURL      string `json:"singleURL"`
}

// MessageType returns the result.
func (*ActionCard) MessageType() MsgType {
	return MsgTypeActionCard
}

type FeedCard struct {
	Links []struct {
		Title      string `json:"title"`
		MessageURL string `json:"messageURL"`
		PicURL     string `json:"picURL"`
	} `json:"links"`
}

// MessageType returns the result.
func (*FeedCard) MessageType() MsgType {
	return MsgTypeFeedCard
}

type MsgType int

const (
	_ MsgType = iota
	MsgTypeText
	MsgTypeMarkdown
	MsgTypeLink
	MsgTypeActionCard
	MsgTypeFeedCard
)

// String returns the string representation.
func (c MsgType) String() string {
	switch c {
	case MsgTypeText:
		return "text"
	case MsgTypeMarkdown:
		return "markdown"
	case MsgTypeLink:
		return "link"
	case MsgTypeActionCard:
		return "actionCard"
	case MsgTypeFeedCard:
		return "feedCard"
	default:
		return "text"
	}
}

// TextMessage returns the result.
func TextMessage(text string) string {
	buf := strings.Builder{}
	buf.WriteString(`{"msgtype":"text","text":{"content":`)
	buf.WriteString(strconv.Quote(text))
	buf.WriteString(`}}`)
	return buf.String()
}

// Format formats or converts the value.
func Format(msg MessageType) string {
	msgType := msg.MessageType()
	buf := strings.Builder{}
	buf.WriteString(`{"msgtype":"`)
	buf.WriteString(msgType.String())
	buf.WriteString(`","`)
	buf.WriteString(msgType.String())
	buf.WriteString(`":`)
	data, _ := jsonx.Marshal(msg)
	buf.Write(data)
	buf.WriteString("}")
	return buf.String()
}
