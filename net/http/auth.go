/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"encoding/base64"
	"net/http"
)

func SetBasicAuth(header http.Header, username, password string) {
	header.Set(HeaderAuthorization, "Basic "+BasicAuth(username, password))
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func GetToken(r *http.Request) string {
	if token := r.Header.Get(HeaderAuthorization); token != "" {
		return token
	}
	if cookie, _ := r.Cookie(HeaderCookieValueToken); cookie != nil {
		return cookie.Value
	}
	return ""
}
