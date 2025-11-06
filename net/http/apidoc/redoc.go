package apidoc

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path"

	http2 "github.com/hopeio/gox/net/http"
)

// RedocOpts configures the Redoc middlewares
type RedocOpts struct {
	// BasePath for the UI, defaults to: /
	BasePath string

	// Path combines with BasePath to construct the path to the UI, defaults to: "docs".
	Path string

	// SpecURL is the URL of the spec document.
	//
	// Defaults to: /openapi.json
	SpecURL string

	// Title for the documentation site, default to: API documentation
	Title string

	// Template specifies a custom template to serve the UI
	Template string

	// RedocURL points to the js that generates the redoc site.
	//
	// Defaults to: https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js
	RedocURL string
}

// EnsureDefaults in case some options are missing
func (r *RedocOpts) EnsureDefaults() {
	if r.BasePath == "" {
		r.BasePath = "/"
	}
	if r.Path == "" {
		r.Path = defaultDocsPath
	}
	if r.SpecURL == "" {
		r.SpecURL = defaultDocsURL
	}
	if r.Title == "" {
		r.Title = defaultDocsTitle
	}

	// redoc-specifics
	if r.RedocURL == "" {
		r.RedocURL = redocLatest
	}
	if r.Template == "" {
		r.Template = redocTemplate
	}
}

// Redoc creates a middleware to serve a documentation site for a swagger spec.
//
// This allows for altering the spec before starting the http listener.
func Redoc(opts RedocOpts, next http.Handler) http.Handler {
	opts.EnsureDefaults()

	pth := path.Join(opts.BasePath, opts.Path)
	tmpl := template.Must(template.New("redoc").Parse(opts.Template))
	assets := bytes.NewBuffer(nil)
	if err := tmpl.Execute(assets, opts); err != nil {
		panic(fmt.Errorf("cannot execute template: %w", err))
	}

	return serveUI(pth, assets.Bytes(), next)
}

const (
	// constants that are common to all UI-serving middlewares
	defaultDocsPath  = "docs"
	defaultDocsURL   = "/openapi.json"
	defaultDocsTitle = "API Documentation"
	redocLatest      = "https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js"
	redocTemplate    = `<!DOCTYPE html>
<html>
  <head>
    <title>{{ .Title }}</title>
		<!-- needed for adaptive design -->
		<meta charset="utf-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">

    <!--
    ReDoc doesn't change outer page styles
    -->
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc spec-url='{{ .SpecURL }}'></redoc>
    <script src="{{ .RedocURL }}"> </script>
  </body>
</html>
`
)

const ()

// serveUI creates a middleware that serves a templated asset as text/html.
func serveUI(pth string, assets []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if path.Clean(r.URL.Path) == pth {
			rw.Header().Set(http2.HeaderContentType, "text/html; charset=utf-8")
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write(assets)

			return
		}

		if next != nil {
			next.ServeHTTP(rw, r)

			return
		}

		rw.Header().Set(http2.HeaderContentType, "text/plain")
		rw.WriteHeader(http.StatusNotFound)
		_, _ = rw.Write([]byte(fmt.Sprintf("%q not found", pth)))
	})
}
