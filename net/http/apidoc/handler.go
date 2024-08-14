package apidoc

import (
	"bytes"
	"github.com/go-openapi/runtime/middleware"
	"github.com/hopeio/utils/fs"
	httpi "github.com/hopeio/utils/net/http"
	"net/http"
	"os"
	"path"
)

// 目录结构 ./api/mod/mod.swagger.json ./api/mod/mod.apidoc.md
// 请求路由 /apidoc /apidoc/swagger/mod/mod.swagger.json /apidoc/markdown/mod/mod.apidoc.md
var UriPrefix = "/apidoc"
var Dir = "./apidoc/"

const TypeSwagger = "swagger"
const TypeMarkdown = "markdown"
const SwaggerEXT = ".swagger.json"
const MarkDownEXT = ".apidoc.md"
const rootModName = "root"

func Swagger(w http.ResponseWriter, r *http.Request) {
	prefixUri := UriPrefix + "/" + TypeSwagger + "/"
	if r.RequestURI[len(r.RequestURI)-5:] == ".json" {
		b, err := os.ReadFile(Dir + r.RequestURI[len(prefixUri):])
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set(httpi.HeaderContentType, "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
		return
	}
	mod := r.RequestURI[len(prefixUri):]
	if mod == rootModName {
		middleware.Redoc(middleware.RedocOpts{
			BasePath: prefixUri,
			SpecURL:  path.Join(prefixUri, rootModName+SwaggerEXT),
			Path:     mod,
		}, http.NotFoundHandler()).ServeHTTP(w, r)
		return
	}
	middleware.Redoc(middleware.RedocOpts{
		BasePath: prefixUri,
		SpecURL:  path.Join(prefixUri+mod, mod+SwaggerEXT),
		Path:     mod,
	}, http.NotFoundHandler()).ServeHTTP(w, r)
}

func Markdown(w http.ResponseWriter, r *http.Request) {
	prefixUri := UriPrefix + "/" + TypeMarkdown + "/"
	mod := r.RequestURI[len(prefixUri):]
	path := Dir + mod + "/" + mod + MarkDownEXT
	if mod == rootModName {
		path = Dir + rootModName + MarkDownEXT
	}
	b, err := os.ReadFile(path)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return
}

func DocList(w http.ResponseWriter, r *http.Request) {
	fileInfos, err := os.ReadDir(Dir)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var buff bytes.Buffer
	for i := range fileInfos {
		if fileInfos[i].Name() == "root.swagger.json" {
			// TODO: 解决root重名 /apidoc=root /apidoc/root
			buff.Write([]byte(`<a href="` + r.RequestURI + "/swagger/" + rootModName + `"> swagger: ` + fileInfos[i].Name() + `</a><br>`))
		}
		if fileInfos[i].Name() == "root.markdown.json" {
			buff.Write([]byte(`<a href="` + r.RequestURI + "/markdown/" + rootModName + `"> markdown: ` + fileInfos[i].Name() + `</a><br>`))
		}
		if fileInfos[i].IsDir() {
			buff.Write([]byte(`<a href="` + r.RequestURI + "/swagger/" + fileInfos[i].Name() + `"> swagger: ` + fileInfos[i].Name() + `</a><br>`))
			buff.Write([]byte(`<a href="` + r.RequestURI + "/markdown/" + fileInfos[i].Name() + `"> markdown: ` + fileInfos[i].Name() + `</a><br>`))
		}
	}
	w.Write(buff.Bytes())
}

func OpenApi(mux *http.ServeMux, uriPrefix, dir string) {
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
	mux.Handle(UriPrefix, http.HandlerFunc(DocList))
	mux.Handle(UriPrefix+"/markdown/", http.HandlerFunc(Markdown))
	mux.Handle(UriPrefix+"/swagger/", http.HandlerFunc(Swagger))
}
