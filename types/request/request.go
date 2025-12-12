/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package request

type OauthReq struct {
	ResponseType   string `json:"responseType,omitempty"`
	ClientID       string `json:"clientID,omitempty"`
	Scope          string `json:"scope,omitempty"`
	RedirectURI    string `json:"redirectURI,omitempty"`
	State          string `json:"state,omitempty"`
	UserID         string `json:"userID,omitempty"`
	AccessTokenExp int64  `json:"accessTokenExp,omitempty"`
	ClientSecret   string `json:"clientSecret,omitempty"`
	Code           string `json:"code,omitempty"`
	RefreshToken   string `json:"refreshToken,omitempty"`
	GrantType      string `json:"grantType,omitempty"`
	AccessType     string `json:"accessType,omitempty"`
	LoginURI       string `json:"loginURI,omitempty"`
}

type Client struct {
	ID     string `json:"id,omitempty"`
	Secret string `json:"secret,omitempty"`
	Domain string `json:"domain,omitempty"`
	UserID string `json:"userID,omitempty"`
}
