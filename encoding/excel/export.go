package excel

import (
	"context"
	"fmt"
	"io"
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
	"github.com/xuri/excelize/v2"
)

type Export struct {
	Name string
	*excelize.File
	Options excelize.Options
}

func (e *Export) WriteTo(w io.Writer) (int64, error) {
	return e.File.WriteTo(w, e.Options)
}

func (res *Export) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *Export) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	if wx, ok := w.(httpx.ResponseWriter); ok {
		header := wx.HeaderX()
		header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
		header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.AttachmentTmpl, res.Name))
	} else {
		header := w.Header()
		header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
		header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.AttachmentTmpl, res.Name))
	}
	n, err := res.File.WriteTo(w, res.Options)
	res.File.Close()
	return int(n), err
}
