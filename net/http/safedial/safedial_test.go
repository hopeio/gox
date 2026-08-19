/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package safedial

import (
	"errors"
	"net/netip"
	"testing"
)

func TestValidateURLRejectsNonHTTPSchemes(t *testing.T) {
	for _, raw := range []string{"file:///etc/passwd", "gopher://x", "ftp://host/x"} {
		if err := ValidateURL(raw); !errors.Is(err, ErrScheme) {
			t.Fatalf("ValidateURL(%q) = %v, want ErrScheme", raw, err)
		}
	}
}

func TestValidateURLAcceptsHTTP(t *testing.T) {
	for _, raw := range []string{"http://agent.local:9000", "https://agent.example.com/x"} {
		if err := ValidateURL(raw); err != nil {
			t.Fatalf("ValidateURL(%q) = %v, want nil", raw, err)
		}
	}
}

func TestLoopbackIsAlwaysBlocked(t *testing.T) {
	for _, p := range []Policy{{}, {AllowPrivate: true}} {
		for _, s := range []string{"127.0.0.1", "::1", "127.5.5.5"} {
			if err := p.IPAllowed(netip.MustParseAddr(s)); !errors.Is(err, ErrLoopback) {
				t.Fatalf("IPAllowed(%q) with AllowPrivate=%v = %v, want ErrLoopback", s, p.AllowPrivate, err)
			}
		}
	}
}

func TestCloudMetadataIsAlwaysBlocked(t *testing.T) {
	p := Policy{AllowPrivate: true}
	if err := p.IPAllowed(netip.MustParseAddr("169.254.169.254")); !errors.Is(err, ErrInternal) {
		t.Fatalf("metadata endpoint reachable: %v", err)
	}
	if err := p.IPAllowed(netip.MustParseAddr("fe80::1")); !errors.Is(err, ErrInternal) {
		t.Fatalf("link-local v6 reachable: %v", err)
	}
	if err := p.IPAllowed(netip.MustParseAddr("0.0.0.0")); !errors.Is(err, ErrInternal) {
		t.Fatalf("unspecified address reachable: %v", err)
	}
}

func TestPrivateRangesFollowThePolicy(t *testing.T) {
	private := []string{"10.0.0.1", "192.168.31.5", "172.16.0.9", "fd00::1"}
	for _, s := range private {
		if err := (Policy{}).IPAllowed(netip.MustParseAddr(s)); !errors.Is(err, ErrPrivate) {
			t.Fatalf("IPAllowed(%q) = %v, want ErrPrivate", s, err)
		}
		if err := (Policy{AllowPrivate: true}).IPAllowed(netip.MustParseAddr(s)); err != nil {
			t.Fatalf("IPAllowed(%q) with AllowPrivate = %v, want nil", s, err)
		}
	}
}

func TestPublicAddressesPass(t *testing.T) {
	for _, s := range []string{"8.8.8.8", "2400:3200::1"} {
		if err := (Policy{}).IPAllowed(netip.MustParseAddr(s)); err != nil {
			t.Fatalf("IPAllowed(%q) = %v, want nil", s, err)
		}
	}
}

func TestControlBlocksByResolvedAddress(t *testing.T) {
	p := Policy{AllowPrivate: true}
	if err := p.Control("tcp4", "127.0.0.1:80", nil); !errors.Is(err, ErrLoopback) {
		t.Fatalf("Control = %v, want ErrLoopback", err)
	}
	if err := p.Control("tcp4", "10.0.0.4:9000", nil); err != nil {
		t.Fatalf("Control = %v, want nil", err)
	}
}

func TestMappedV4IsUnwrapped(t *testing.T) {
	if err := (Policy{AllowPrivate: true}).IPAllowed(netip.MustParseAddr("::ffff:127.0.0.1")); !errors.Is(err, ErrLoopback) {
		t.Fatalf("mapped loopback slipped through: %v", err)
	}
}
