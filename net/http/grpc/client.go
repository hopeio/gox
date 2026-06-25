/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package grpc

import (
	"crypto/tls"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)


func NewClient(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))...)
}

func NewClientTLS(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ServerName: strings.Split(addr, ":")[0], InsecureSkipVerify: true})))...)
}
