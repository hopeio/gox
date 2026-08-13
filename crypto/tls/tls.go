/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

// NewServerTLSConfig creates and returns a new instance.
func NewServerTLSConfig(certFile, keyFile string, clients ...string) (*tls.Config, error) {
	if certFile == "" || keyFile == "" {
		return nil, errors.New("certFile or keyFile is empty")
	}
	var err error
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	var certPool *x509.CertPool
	clientAuth := tls.NoClientCert
	if len(clients) > 0 {
		clientAuth = tls.RequireAndVerifyClientCert
		certPool = x509.NewCertPool()
		for _, client := range clients {
			ca, err := os.ReadFile(client)
			if err != nil {
				// 库函数不应 Fatal 杀进程，交由调用方处理
				return nil, fmt.Errorf("read client ca %s: %w", client, err)
			}
			if ok := certPool.AppendCertsFromPEM(ca); !ok {
				return nil, fmt.Errorf("append client ca %s: invalid PEM", client)
			}
		}
	}

	return &tls.Config{
		Certificates: certs,
		// 未提供客户端 CA 时不能开启 RequireAndVerifyClientCert，否则所有连接都被拒绝
		ClientAuth: clientAuth,
		ClientCAs:  certPool,
	}, nil
}

// NewClientTLSConfig creates and returns a new instance.
func NewClientTLSConfig(certFile, serverName string) (*tls.Config, error) {
	b, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		return nil, fmt.Errorf("credentials: failed to append certificates")
	}
	if serverName == "" {
		return &tls.Config{RootCAs: cp}, nil
	}
	return &tls.Config{ServerName: serverName, RootCAs: cp}, nil
}
