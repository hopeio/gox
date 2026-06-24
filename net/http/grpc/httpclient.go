/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package grpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	httpx "github.com/hopeio/gox/net/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// HTTPClient 通过 HTTP/2 调用 gRPC 服务的轻量客户端，无需 gRPC-Go 运行时。
type HTTPClient struct {
	BaseURL string
	Client  *http.Client
}

// NewHTTPClient 创建 gRPC 客户端，根据 baseURL scheme 自动选择 h2c 或 TLS。
func NewHTTPClient(baseURL string) *HTTPClient {
	var transport http.RoundTripper
	if strings.HasPrefix(baseURL, "http://") {
		tr := &http.Transport{}
		tr.Protocols = new(http.Protocols)
		tr.Protocols.SetUnencryptedHTTP2(true)
		transport = tr
	}
	return &HTTPClient{BaseURL: baseURL, Client: &http.Client{Transport: transport}}
}


// NewHTTPClientWithClient 使用自定义 http.Client 创建 gRPC 客户端。
func NewHTTPClientWithClient(baseURL string, client *http.Client) *HTTPClient {
	return &HTTPClient{BaseURL: baseURL, Client: client}
}

// Call 通过 HTTP/2 调用 gRPC unary 方法。
// method 格式: /package.Service/Method，如 /user.UserService/GetUser。
func Call[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]](
	ctx context.Context,
	c *HTTPClient,
	method string,
	req ReqPtr,
) (RespPtr, error) {
	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 构建 gRPC length-prefixed 帧: 1 字节压缩标志 + 4 字节大端长度 + payload
	frame := make([]byte, 5+len(data))
	frame[0] = 0 // 无压缩
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+method, bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set(httpx.HeaderContentType, httpx.ContentTypeGrpc+"+proto")
	httpReq.Header.Set("Te", "trailers")
	httpReq.Header.Set(httpx.HeaderUserAgent, "grpc-go/1.0")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 检查 gRPC trailer 状态
	grpcStatus := resp.Trailer.Get(httpx.HeaderGrpcStatus)
	if grpcStatus != "" && grpcStatus != "0" {
		grpcMessage := resp.Trailer.Get(httpx.HeaderGrpcMessage)
		code, _ := strconv.Atoi(grpcStatus)
		return nil, status.Error(codes.Code(code), grpcMessage)
	}

	if len(body) < 5 {
		return nil, fmt.Errorf("invalid gRPC response: body too short (%d bytes)", len(body))
	}

	if body[0] == 1 {
		return nil, status.Error(codes.Unimplemented, "compressed response not supported")
	}

	msgLength := binary.BigEndian.Uint32(body[1:5])
	if uint32(len(body)) < 5+msgLength {
		return nil, fmt.Errorf("invalid gRPC response: declared length %d exceeds body %d", msgLength, len(body)-5)
	}

	msgData := body[5 : 5+msgLength]
	var result Resp
	resultPtr := any(&result).(RespPtr)
	if err := proto.Unmarshal(msgData, resultPtr); err != nil {
		return nil, err
	}
	return resultPtr, nil
}
