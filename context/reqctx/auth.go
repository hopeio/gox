/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package reqctx

import (
	"net/http"
)

const (
	HeaderAuthorization    = "Authorization"
	HeaderCookie           = "Cookie"
	HeaderCookieValueToken = "token"
)

type AuthInfo interface {
	GetId() string
}

type ParseToken interface {
	ParseToken(token string, secret []byte) error
}

/*
type Authorization struct {
	AuthInfo `json:"auth"`
	jwt.RegisteredClaims
	AuthRaw string `json:"-"`
}

func (x *Authorization) Validate() error {
	return nil
}

func (x *Authorization) GenerateToken(secret []byte) (string, error) {
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, x)
	token, err := tokenClaims.SignedString(secret)
	return token, err
}

func (x *Authorization) ParseToken(token string, secret []byte) error {
	_, err := jwti.ParseToken(x, token, secret)
	if err != nil {
		return err
	}
	x.ID = x.AuthInfo.GetId()
	authBytes, _ := json.Marshal(x.AuthInfo)
	x.AuthRaw = stringsx.FromBytes(authBytes)
	return nil
}
*/

func GetToken[REQ ReqCtx](r REQ) string {
	header := r.RequestHeader()
	if token := header.Get(HeaderAuthorization); token != "" {
		return token
	}
	cookie := header.Get(HeaderCookie)
	parsedCookie, err := http.ParseCookie(cookie)
	if err != nil {
		return ""
	}
	for _, v := range parsedCookie {
		if v.Name == HeaderCookieValueToken {
			return v.Value
		}
	}
	return ""
}

type Auth struct {
	AuthRaw string
	AuthID  string
	AuthInfo
}
