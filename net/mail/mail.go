/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"

	"github.com/hopeio/gox/text/template"

	"net"
	"net/smtp"
)

// 550,Mailbox not found or access denied.是因为收件邮箱不存在
type Mail struct {
	Addr                                          string
	FromName, From, Subject, ContentType, Content string
	To                                            []string
	Auth                                          smtp.Auth
}

const msg = `{{define "mail"}}To: {{join .To ",\n\t"}}
From: {{.FromName}} <{{.From}}>
Subject: {{.Subject}}
Content-Type: {{if .ContentType}}{{.ContentType}}{{- else}}text/html; charset=UTF-8{{end}}

{{.Content}}{{end}}
`

func init() {
	template.Parse(msg)
}

func (m *Mail) GenMsg() ([]byte, error) {
	var buf = new(bytes.Buffer)
	err := template.Execute(buf, "mail", m)
	if err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}
	return buf.Bytes(), nil
}
func (m *Mail) SendMail() error {
	msg, err := m.GenMsg()
	if err != nil {
		return err
	}
	return smtp.SendMail(m.Addr, m.Auth, m.From, m.To, msg)
}

func (m *Mail) SendMailTLS() error {
	client, err := createSMTPClient(m.Addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if m.Auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(m.Auth); err != nil {
				return err
			}
		}
	}
	if err := client.Mail(m.From); err != nil {
		return err
	}
	for _, addr := range m.To {
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	msg, err := m.GenMsg()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return client.Quit()
}

func createSMTPClient(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}
