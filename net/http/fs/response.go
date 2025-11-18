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

func (res *ResponseFile) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	return res.CommonRespond(ctx, httpx.ResponseWriterWrapper{ResponseWriter: w})
}

func (res *ResponseFile) CommonRespond(ctx context.Context, w httpx.CommonResponseWriter) (int, error) {
	header := w.Header()
	header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.AttachmentTmpl, res.Name))
	n, err := res.Body.WriteTo(w)
	res.Body.Close()
	return int(n), err
}
