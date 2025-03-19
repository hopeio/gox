/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package grpc

import (
	"crypto/tls"
	"github.com/hopeio/utils/errors/multierr"
	httpi "github.com/hopeio/utils/net/http/consts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"strings"
)

var Internal = &metadata.MD{httpi.HeaderInternal: []string{"true"}}

type clientConns map[string]*grpc.ClientConn

func (cs clientConns) Close() error {
	var err error
	for _, conn := range cs {
		err = multierr.Append(err, conn.Close())
	}
	return err
}

func NewClient(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {

	// Set up a connection to the server.
	conn, err := grpc.NewClient(addr, append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewTLSClient(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// Set up a connection to the server.
	conn, err := grpc.NewClient(addr, append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ServerName: strings.Split(addr, ":")[0], InsecureSkipVerify: true})))...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
