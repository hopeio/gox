/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fiber

import (
	"net/url"

	"github.com/hopeio/gox/net/http"
	stringsx "github.com/hopeio/gox/strings"
	"github.com/valyala/fasthttp"
)

func GetToken(req *fasthttp.Request) string {
	if token := stringsx.BytesToString(req.Header.Peek(http.HeaderAuthorization)); token != "" {
		return token
	}
	if cookie := stringsx.BytesToString(req.Header.Cookie(http.HeaderCookieValueToken)); len(cookie) > 0 {
		token, _ := url.QueryUnescape(cookie)
		return token
	}
	return ""
}
