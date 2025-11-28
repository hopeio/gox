/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package apidoc

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	http2 "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/os/fs"
)

// 目录结构 ./api/mod/mod.openapi.json
// 请求路由 /apidoc /apidoc/openapi/mod/mod.openapi.json
var UriPrefix = "/apidoc"
var Dir = "./apidoc/"

const TypeOpenapi = "openapi"
const OpenapiEXT = ".openapi.json"
const SwaggerEXT = ".swagger.json"
const JsonEXT = ".json"

func OpenApi(w http.ResponseWriter, r *http.Request) {
	prefixUri := UriPrefix + "/" + TypeOpenapi + "/"
	if r.RequestURI[len(r.RequestURI)-5:] == ".json" {
		b, err := os.ReadFile(Dir + r.RequestURI[len(prefixUri):])
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set(http2.HeaderContentType, "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
		return
	}
	mod := r.RequestURI[len(prefixUri):]
	Redoc(RedocOpts{
		BasePath: prefixUri,
		SpecURL:  path.Join(prefixUri, mod+JsonEXT),
		Path:     mod,
	}, http.NotFoundHandler()).ServeHTTP(w, r)
}
func DocList(w http.ResponseWriter, r *http.Request) {
	fileInfos, err := os.ReadDir(Dir)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var buff bytes.Buffer
	for i := range fileInfos {

		if strings.HasSuffix(fileInfos[i].Name(), JsonEXT) {
			mod := strings.TrimSuffix(fileInfos[i].Name(), JsonEXT)
			buff.Write([]byte(`<a href="` + r.RequestURI + "/openapi/" + mod + `"> ` + mod + `</a><br>`))
		}
	}
	w.Write(buff.Bytes())
}

func ApiDoc(mux *http.ServeMux, uriPrefix, dir string) {
	if dir != "" {
		if b := dir[len(dir)-1:]; b == "/" || b == "\\" {
			Dir = dir
		} else {
			Dir = dir + fs.PathSeparator
		}
	}
	if uriPrefix != "" {
		UriPrefix = uriPrefix
	}
	mux.HandleFunc(UriPrefix, DocList)
	mux.HandleFunc(UriPrefix+"/openapi/{file...}", OpenApi)
}

func WriteToFile(docDir, modName string, doc *openapi3.T) error {
	if doc == nil {
		return errors.New("doc is nil")
	}
	if docDir == "" {
		docDir = Dir
	}

	err := os.MkdirAll(docDir, os.ModePerm)
	if err != nil {
		return err
	}

	path := filepath.Join(docDir, modName+OpenapiEXT)

	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}
	var file *os.File
	file, err = os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(doc)
	if err != nil {
		return err
	}
	return nil
}
