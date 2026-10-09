/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

// Package safedial adds an address gate to outbound HTTP requests whose
// target URLs are supplied by users or operators.
//
// Probing, callbacks, and fetching remote resources hand the URL straight
// to http.Client and feed the response body back, which opens a hole into
// the server's internal network (cloud metadata at 169.254.169.254 being
// the classic case). Validating the URL alone is not enough: redirects and
// DNS rebinding both bypass it. The gate therefore sits in Dialer.Control,
// where the address is the IP actually about to be dialed.
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

// Policy decides which destination addresses may be connected.
type Policy struct {
	// AllowPrivate permits RFC1918 / ULA ranges. Probing agents on a LAN
	// usually needs it on; loopback and link-local are always blocked.
	AllowPrivate bool
	// Allow entries pass immediately, overriding the built-in rules and
	// Deny (e.g. allow only 10.0.0.0/8).
	Allow []netip.Prefix
	// Deny entries are an extra blocklist that wins over built-in allow
	// (e.g. block a public range).
	Deny []netip.Prefix
	// AllowFunc fully replaces the built-in rules when non-nil
	// (Allow/Deny still take precedence over it).
	// It receives the Unmapped address.
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

// ValidateURL checks only what can be decided statically: scheme and host presence.
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

// IPAllowed reports whether a specific IP may be connected.
// Decision order: Allow allowlist > Deny blocklist > AllowFunc (replaces
// built-in rules) > built-in rules.
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

// Control is passed to net.Dialer.Control and blocks disallowed targets
// before the connection is established.
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

// Client returns an http.Client with the address gate that does not
// follow redirects.
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
