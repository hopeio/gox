package fs

import (
	"fmt"
	"io"
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
)

type ResponseFile struct {
	Name string        `json:"name"`
	Body io.ReadCloser `json:"body,omitempty"`
}

func (res *ResponseFile) Response(w http.ResponseWriter) (int, error) {
	return res.CommonResponse(httpx.CommonResponseWriter{ResponseWriter: w})
}

func (res *ResponseFile) CommonResponse(w httpx.ICommonResponseWriter) (int, error) {
	header := w.Header()
	header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.AttachmentTmpl, res.Name))
	n, err := io.Copy(w, res.Body)
	res.Body.Close()
	return int(n), err
}

type ResponseFileWriteTo struct {
	Name string               `json:"name"`
	Body httpx.WriterToCloser `json:"body,omitempty"`
}

func (res *ResponseFileWriteTo) Response(w http.ResponseWriter) (int, error) {
	return res.CommonResponse(httpx.CommonResponseWriter{ResponseWriter: w})
}

func (res *ResponseFileWriteTo) CommonResponse(w httpx.ICommonResponseWriter) (int, error) {
	header := w.Header()
	header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.AttachmentTmpl, res.Name))
	n, err := res.Body.WriteTo(w)
	res.Body.Close()
	return int(n), err
}
