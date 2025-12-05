/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http3

import (
	"net/http"

	"github.com/quic-go/quic-go/http3"
)

func NewClient() *http.Client {
	return &http.Client{
		Transport: &http3.Transport{},
	}
}
