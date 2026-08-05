/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package net

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
)

// IPStrToUint32 returns the result.
func IPStrToUint32(ipStr string) uint32 {
	return IPv4ToUint32(net.ParseIP(ipStr))
}

// IPv4ToUint32 returns the result.
func IPv4ToUint32(ip net.IP) uint32 {
	if ip.To4() == nil {
		return 0
	}

	ipBytes := ip.To4()
	return uint32(ipBytes[0])<<24 | uint32(ipBytes[1])<<16 | uint32(ipBytes[2])<<8 | uint32(ipBytes[3])
}

// Uint32ToIPv4 returns the result.
func Uint32ToIPv4(ipInt uint32) net.IP {
	if ipInt == 0 {
		return nil
	}
	ip := make(net.IP, 4)
	ip[0] = byte(ipInt >> 24)
	ip[1] = byte(ipInt >> 16)
	ip[2] = byte(ipInt >> 8)
	ip[3] = byte(ipInt)
	return ip
}

// ExternalIPStr returns the result.
func ExternalIPStr() string {
	ip, _ := ExternalIP()
	return ip.String()
}

// ExternalIP performs the operation.
func ExternalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := AddrToIP(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("network error")
}

// AddrToIP updates or inserts a value.
func AddrToIP(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}

	return ip
}

// CommonIPV4 performs the operation.
func CommonIPV4() (string, error) {
	res, err := http.Get("http://txt.go.sohu.com/ip/soip")
	if err != nil {
		return "", errors.New("network error")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	reg := regexp.MustCompile(`\d+.\d+.\d+.\d+`)
	return string(reg.Find(body)), nil
}

// CommonIPv6 performs the operation.
func CommonIPv6() (string, error) {
	resp, err := http.Get("https://api64.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var ip string
	if _, err := fmt.Fscanf(resp.Body, "%s", &ip); err != nil {
		return "", err
	}

	return ip, nil
}

// LocalIPv4s performs the operation.
func LocalIPv4s() ([]net.IP, error) {
	var ipv4Addrs []net.IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipv4Addrs = append(ipv4Addrs, ipnet.IP)
			}
		}
	}
	return ipv4Addrs, nil
}

// IPv4s performs the operation.
func IPv4s() ([]net.IP, error) {
	var ipv4Addrs []net.IP
	address, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, a := range address {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsPrivate() && ipNet.IP.IsGlobalUnicast() {
			if ipNet.IP.To4() != nil {
				ipv4Addrs = append(ipv4Addrs, ipNet.IP)
			}
		}
	}
	return ipv4Addrs, nil
}

// LocalIPv6s performs the operation.
func LocalIPv6s() ([]net.IP, error) {
	var ipv6Addrs []net.IP

	// Get addresses of all network interfaces
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		// Check whether the address family is an IP interface address
		if ipNet, ok := addr.(*net.IPNet); ok {
			// Check whether it is an IPv6 address
			if ipNet.IP.To4() == nil && !ipNet.IP.IsLoopback() {
				ipv6Addrs = append(ipv6Addrs, ipNet.IP)
			}
		}
	}

	return ipv6Addrs, nil
}

// IPv6s performs the operation.
func IPv6s() ([]net.IP, error) {
	var ipv6s []net.IP

	// Get addresses of all network interfaces
	address, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range address {
		// Check whether the address family is an IP interface address
		if ipNet, ok := addr.(*net.IPNet); ok {
			// Check whether it is an IPv6 address
			if ipNet.IP.To4() == nil && !ipNet.IP.IsPrivate() && ipNet.IP.IsGlobalUnicast() {
				ipv6s = append(ipv6s, ipNet.IP)
			}
		}
	}

	return ipv6s, nil
}

// PrivateIPv4 performs the operation.
func PrivateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if ip.IsPrivate() {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}
