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
	// Allow 命中即放行，优先于内置规则与 Deny（如只放行 10.0.0.0/8）。
	Allow []netip.Prefix
	// Deny 追加黑名单，优先于内置规则的放行（如拦某个公网网段）。
	Deny []netip.Prefix
	// AllowFunc 完全接管判断：非 nil 时替代内置规则（Allow/Deny 仍先于它生效）。
	// 入参是已 Unmap 的地址。
	AllowFunc func(ip netip.Addr) error
}

var (
	ErrScheme   = errors.New("only http and https URLs are allowed")
	ErrNoHost   = errors.New("URL has no host")
	ErrLoopback = errors.New("loopback addresses are not allowed")
	ErrInternal = errors.New("link-local and metadata addresses are not allowed")
	ErrPrivate  = errors.New("private addresses are not allowed")
	ErrDenied   = errors.New("address is denied by policy")
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
// 判定顺序：Allow 白名单 > Deny 黑名单 > AllowFunc（接管内置规则）> 内置规则。
func (p Policy) IPAllowed(ip netip.Addr) error {
	ip = ip.Unmap()
	if matchAny(ip, p.Allow) {
		return nil
	}
	if matchAny(ip, p.Deny) {
		return ErrDenied
	}
	if p.AllowFunc != nil {
		return p.AllowFunc(ip)
	}
	return p.builtinAllowed(ip)
}

func (p Policy) builtinAllowed(ip netip.Addr) error {
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

func matchAny(ip netip.Addr, prefixes []netip.Prefix) bool {
	if !ip.IsValid() {
		return false
	}
	for _, prefix := range prefixes {
		if prefix.Contains(ip) {
			return true
		}
	}
	return false
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
