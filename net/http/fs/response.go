package fs

import (
	"context"
	"fmt"
	"io"
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
)

type ReadCloser interface {
	io.ReadCloser
	Name() string
}

type ResponseFile struct {
	Name string               `json:"name"`
	Body httpx.WriterToCloser `json:"body,omitempty"`
}

func (res *ResponseFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *ResponseFile) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	if wx, ok := w.(httpx.ResponseWriter); ok {
		header := wx.HeaderX()
		header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
		header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.AttachmentTmpl, res.Name))
	} else {
		header := w.Header()
		header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
		header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.AttachmentTmpl, res.Name))
	}
	n, err := res.Body.WriteTo(w)
	res.Body.Close()
	return int(n), err
}
