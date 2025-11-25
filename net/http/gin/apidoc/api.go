/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package apidoc

import (
	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/net/http/apidoc"
	gin2 "github.com/hopeio/gox/net/http/gin"
	"github.com/hopeio/gox/os/fs"
)

func OpenApi(mux *gin.Engine, uriPrefix, dir string) {
	if dir != "" {
		if b := dir[len(dir)-1:]; b == "/" || b == "\\" {
			apidoc.Dir = dir
		} else {
			apidoc.Dir = dir + fs.PathSeparator
		}
	}
	if uriPrefix != "" {
		apidoc.UriPrefix = uriPrefix
	}

	mux.GET(apidoc.UriPrefix, gin2.Wrap(apidoc.DocList))
	mux.GET(apidoc.UriPrefix+"/openapi/*file", gin2.Wrap(apidoc.OpenApi))
}
