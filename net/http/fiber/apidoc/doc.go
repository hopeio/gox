/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package apidoc

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"

	"github.com/gofiber/fiber/v3"
	"github.com/hopeio/gox/net/http/apidoc"
	"github.com/hopeio/gox/os/fs"
)

func Openapi(ctx fiber.Ctx) error {
	prefixUri := apidoc.UriPrefix + "/" + apidoc.TypeOpenapi + "/"
	requestURI := string(ctx.Request().URI().RequestURI())
	if requestURI[len(requestURI)-5:] == ".json" {
		b, err := os.ReadFile(apidoc.Dir + requestURI[len(prefixUri):])
		if err != nil {
			return err
		}
		ctx.Response().Header.Set("Content-Type", "application/json; charset=utf-8")
		ctx.Response().SetStatusCode(http.StatusOK)
		ctx.Write(b)
		return nil
	}
	mod := requestURI[len(prefixUri):]

	opts := apidoc.RedocOpts{
		BasePath: prefixUri,
		SpecURL:  path.Join(prefixUri+mod, mod+apidoc.OpenapiEXT),
		Path:     mod,
	}
	opts.EnsureDefaults()
	pth := path.Join(opts.BasePath, opts.Path)
	tmpl := template.Must(template.New("redoc").Parse(opts.Template))
	assets := bytes.NewBuffer(nil)
	if err := tmpl.Execute(assets, opts); err != nil {
		panic(fmt.Errorf("cannot execute template: %w", err))
	}
	if path.Clean(ctx.Path()) == pth {
		ctx.Response().Header.Set("Content-Type", "text/html; charset=utf-8")
		ctx.Response().SetStatusCode(http.StatusOK)
		_, _ = ctx.Write(assets.Bytes())
	}
	return nil
}
func DocList(ctx fiber.Ctx) error {
	fileInfos, err := os.ReadDir(apidoc.Dir)
	if err != nil {
		return err
	}
	requestURI := string(ctx.Request().URI().RequestURI())
	var buff bytes.Buffer
	for i := range fileInfos {
		if fileInfos[i].IsDir() {
			buff.Write([]byte(`<a href="` + requestURI + "/openapi/" + fileInfos[i].Name() + `"> openapi: ` + fileInfos[i].Name() + `</a><br>`))
		}
	}
	_, err = ctx.Write(buff.Bytes())
	return err
}

func ApiDoc(mux *fiber.App, uriPrefix, dir string) {
	if uriPrefix != "" {
		apidoc.UriPrefix = uriPrefix
	}
	if dir != "" {
		if b := dir[len(dir)-1:]; b == "/" || b == "\\" {
			apidoc.Dir = dir
		} else {
			apidoc.Dir = dir + fs.PathSeparator
		}
	}
	mux.Get(apidoc.UriPrefix, DocList)
	mux.Get(apidoc.UriPrefix+"/openapi/", Openapi)
}
