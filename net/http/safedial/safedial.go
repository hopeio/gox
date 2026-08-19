/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

// Package safedial 给“服务端按用户/运维填写的 URL 主动出站”加地址闸门。
//
// 探活、回调、拉远端资源若把 URL 直接交给 http.Client，再把响应体回吐，
// 就等于开了一个读服务器内网的口子（云厂商元数据 169.254.169.254 最典型）。
// 只校验 URL 不够：重定向和 DNS rebinding 都能绕过。闸门放在 Dialer.Control——
// 那时拿到的才是真正要连的 IP。
package safedial

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"syscall"
	"time"
)

// Policy 决定哪些目标地址可以连。
type Policy struct {
	// AllowPrivate 允许 RFC1918 / ULA。局域网 Agent 探活常要打开；
	// 回环与链路本地无论如何都拦。
	AllowPrivate bool
}

var (
	ErrScheme   = errors.New("only http and https URLs are allowed")
	ErrNoHost   = errors.New("URL has no host")
	ErrLoopback = errors.New("loopback addresses are not allowed")
	ErrInternal = errors.New("link-local and metadata addresses are not allowed")
	ErrPrivate  = errors.New("private addresses are not allowed")
)

// ValidateURL 只做能静态判断的部分：协议与是否带 host。
func ValidateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrScheme
	}
	if u.Host == "" {
		return ErrNoHost
	}
	return nil
}

// IPAllowed 判断某个具体 IP 是否可连。
func (p Policy) IPAllowed(ip netip.Addr) error {
	ip = ip.Unmap()
	switch {
	case !ip.IsValid(), ip.IsUnspecified():
		return ErrInternal
	case ip.IsLoopback():
		return ErrLoopback
	case ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast(), ip.IsMulticast(), ip.IsInterfaceLocalMulticast():
		return ErrInternal
	case ip.IsPrivate():
		if p.AllowPrivate {
			return nil
		}
		return ErrPrivate
	}
	return nil
}

// Control 传给 net.Dialer.Control，在建立连接前拦下不允许的目标。
func (p Policy) Control(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("split %q: %w", address, err)
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return fmt.Errorf("parse %q: %w", host, err)
	}
	return p.IPAllowed(ip)
}

// Client 返回一个带地址闸门、且不跟随重定向的 http.Client。
func (p Policy) Client(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: timeout, Control: p.Control}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, addr)
			},
			DisableKeepAlives: true,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
