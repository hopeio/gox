/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package qyweixin

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

type MsgType int

const (
	_ MsgType = iota
	MsgTypeText
	MsgTypeMarkdown
	MsgTypeImage
	MsgTypeFile
	MsgTypeNews
	MsgTypeVoice
	MsgTypeTemplateCard
)

func (c MsgType) String() string {
	switch c {
	case MsgTypeText:
		return "text"
	case MsgTypeMarkdown:
		return "markdown"
	case MsgTypeImage:
		return "image"
	case MsgTypeFile:
		return "file"
	case MsgTypeNews:
		return "news"
	case MsgTypeVoice:
		return "voice"
	case MsgTypeTemplateCard:
		return "template_card"
	default:
		return ""
	}
}

type Markdown struct {
	Content string `json:"content"`
}

func (*Markdown) MessageType() MsgType {
	return MsgTypeMarkdown
}

type Text struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

func (*Text) MessageType() MsgType {
	return MsgTypeText
}

type Image struct {
	Base64 string `json:"base64"`
	Md5    string `json:"md5"`
}

func (Image) MessageType() MsgType {
	return MsgTypeImage
}

type News struct {
	Articles []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Url         string `json:"url"`
		Picurl      string `json:"picurl"`
	} `json:"articles"`
}

func (*News) MessageType() MsgType {
	return MsgTypeNews
}

type File struct {
	MediaId string `json:"media_id"`
}

func (File) MessageType() MsgType {
	return MsgTypeFile
}

type Voice struct {
	MediaId string `json:"media_id"`
}

func (Voice) MessageType() MsgType {
	return MsgTypeVoice
}

type TemplateCard struct {
	CardType              string                `json:"card_type"`
	Source                Source                `json:"source"`
	MainTitle             MainTitle             `json:"main_title"`
	EmphasisContent       EmphasisContent       `json:"emphasis_content"`
	QuoteArea             QuoteArea             `json:"quote_area"`
	SubTitleText          string                `json:"sub_title_text"`
	HorizontalContentList HorizontalContentList `json:"horizontal_content_list"`
	JumpList              Jump                  `json:"jump_list"`
	CardAction            CardAction            `json:"card_action"`
}

type Source struct {
	IconUrl   string `json:"icon_url"`
	Desc      string `json:"desc"`
	DescColor int    `json:"desc_color"`
}

type MainTitle struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type EmphasisContent struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type QuoteArea struct {
	Type      int    `json:"type"`
	Url       string `json:"url"`
	Appid     string `json:"appid"`
	Pagepath  string `json:"pagepath"`
	Title     string `json:"title"`
	QuoteText string `json:"quote_text"`
}
type HorizontalContentList struct {
	Keyname string `json:"keyname"`
	Value   string `json:"value"`
	Type    int    `json:"type,omitempty"`
	Url     string `json:"url,omitempty"`
	MediaId string `json:"media_id,omitempty"`
}

type Jump struct {
	Type     int    `json:"type"`
	Url      string `json:"url,omitempty"`
	Title    string `json:"title"`
	Appid    string `json:"appid,omitempty"`
	Pagepath string `json:"pagepath,omitempty"`
}

type CardAction struct {
	Type     int    `json:"type"`
	Url      string `json:"url"`
	Appid    string `json:"appid"`
	Pagepath string `json:"pagepath"`
}

func (*TemplateCard) MessageType() MsgType {
	return MsgTypeTemplateCard
}

func MarkdownMessage(text string) string {
	buf := strings.Builder{}
	buf.WriteString(`{"msgtype":"markdown","markdown":{"content":`)
	buf.WriteString(strconv.Quote(text))
	buf.WriteString(`}}`)
	return buf.String()
}

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

func Upload() {
	// 文件类型，分别有语音(voice)和普通文件(file)
	const api = "https://qyapi.weixin.qq.com/cgi-bin/webhook/upload_media?key=KEY&type=TYPE"
}

type UploadRes struct {
	Errcode   int    `json:"errcode"`
	Errmsg    string `json:"errmsg"`
	Type      string `json:"type"`
	MediaId   string `json:"media_id"`
	CreatedAt string `json:"created_at"`
}
