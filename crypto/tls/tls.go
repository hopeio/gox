/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hopeio/gox/log"
	"os"
	"errors"
)

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
	if len(clients) > 0 {
		certPool = x509.NewCertPool()
		for _, client := range clients {
			ca, err := os.ReadFile(client)
			if err != nil {
				log.Fatalf("ioutil.ReadFile err: %v", err)
			}
			if ok := certPool.AppendCertsFromPEM(ca); !ok {
				log.Fatalf("certPool.AppendCertsFromPEM err")
			}
		}
	}

	return &tls.Config{
		Certificates: certs,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}, nil
}

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
