package fs

import (
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

func (res *ResponseFile) Respond(w http.ResponseWriter) (int, error) {
	return res.CommonRespond(httpx.CommonResponseWriter{ResponseWriter: w})
}

func (res *ResponseFile) CommonRespond(w httpx.ICommonResponseWriter) (int, error) {
	header := w.Header()
	header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.AttachmentTmpl, res.Name))
	n, err := res.Body.WriteTo(w)
	res.Body.Close()
	return int(n), err
}
