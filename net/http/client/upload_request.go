/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	httpx "github.com/hopeio/gox/net/http"
)

const (
	defaultChunkSize = 5 * 1024 * 1024 // 每个分块的大小，这里是5MB
)

var (
	ContentTypeKey = http.CanonicalHeaderKey("Content-Type")
)

type UploadMode uint16

const (
	UModeNormal UploadMode = iota
	UModeStream
	UModeChunked
	UModeChunkedConcurrent
)

type UploadReq struct {
	Url       string
	uploader  *Uploader
	ctx       context.Context
	header    http.Header //请求级请求头
	boundary  string
	mode      UploadMode
	chunkSize int64
}

type Multipart struct {
	Name     string
	Value    string
	Filename string
	Header   textproto.MIMEHeader
	io.Reader
}

type File struct {
	Path string
	*os.File
}

func NewFile(path string) (*File, error) {
	osfile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &File{
		Path: path,
		File: osfile,
	}, nil
}

func (f *File) ToMutilPart(param string) *Multipart {
	contentType := mime.TypeByExtension(filepath.Ext(f.Path))
	return NewMultipart(param, path.Base(f.Path), textproto.MIMEHeader{httpx.HeaderContentType: []string{contentType}}, f.File)
}

func NewMultipart(name, filename string, header textproto.MIMEHeader, reader io.Reader) *Multipart {
	return &Multipart{
		Name:     name,
		Filename: filename,
		Header:   header,
		Reader:   reader,
	}
}

func NewUploadReq(url string) *UploadReq {
	return &UploadReq{
		ctx:      context.Background(),
		Url:      url,
		uploader: DefaultUploader,
	}
}

func (r *UploadReq) Context(ctx context.Context) *UploadReq {
	r.ctx = ctx
	return r
}

func (r *UploadReq) Uploader(u *Uploader) *UploadReq {
	r.uploader = u
	return r
}

func (r *UploadReq) Boundary(boundary string) *UploadReq {
	r.boundary = boundary
	return r
}

func (r *UploadReq) Mode(mode UploadMode) *UploadReq {
	r.mode = mode
	return r
}

func (r *UploadReq) ChunkSize(chunkSize int64) *UploadReq {
	if chunkSize < 512 {
		panic("buffer size should > 512")
	}
	r.chunkSize = chunkSize
	return r
}

func (r *UploadReq) UploadMultipart(formData map[string]string, files ...*Multipart) error {
	body := bufPool.Get().(*bytes.Buffer)
	defer func() {
		body.Reset()
		bufPool.Put(body)
	}()
	w := multipart.NewWriter(body)

	if r.boundary != "" {
		if err := w.SetBoundary(r.boundary); err != nil {
			return err
		}
	}

	h := make(textproto.MIMEHeader)
	for k, v := range formData {
		h.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.FormDataFieldTmpl, escapeQuotes(k)))
		part, err := w.CreatePart(h)
		if err != nil {
			return err
		}
		_, err = part.Write([]byte(v))
		if err != nil {
			return err
		}
	}

	h.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	for _, file := range files {
		h.Set(httpx.HeaderContentDisposition, multipart.FileContentDisposition(file.Name, file.Filename))
		part, err := w.CreatePart(h)
		if err != nil {
			return err
		}
		_, err = io.Copy(part, file.Reader)
		if err != nil {
			return err
		}
	}
	err := w.Close()
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, r.Url, body)
	if err != nil {
		return err
	}
	if r.header != nil {
		req.Header = r.header
	}

	d := r.uploader
	httpx.CopyHttpHeader(req.Header, d.header)
	for _, opt := range d.httpRequestOptions {
		opt(req)
	}
	req.Header.Set(httpx.HeaderContentType, w.FormDataContentType())
	_, err = d.httpClient.Do(req)
	if err != nil {
		return err
	}
	// TODO: error handler, retry
	return nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (r *UploadReq) UploadReader(reader io.Reader, name string) error {
	u := r.uploader

	req, err := http.NewRequestWithContext(r.ctx, http.MethodPost, r.Url, reader)
	if err != nil {
		return err
	}
	if r.header != nil {
		req.Header = r.header
	}
	httpx.CopyHttpHeader(req.Header, u.header)
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	req.Header.Set(httpx.HeaderContentLength, strconv.FormatInt(r.chunkSize, 10))
	req.Header.Set(httpx.HeaderContentDisposition, httpx.FormatAttachment(name))
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (r *UploadReq) UploadReaderChunked(reader io.ReaderAt, name string, total int64) error {

	var start int64
	var end int64 = r.chunkSize - 1

	u := r.uploader

	req, err := http.NewRequestWithContext(r.ctx, http.MethodPost, r.Url, nil)
	if err != nil {
		return err
	}
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	for start < total {
		body := io.NewSectionReader(reader, start, r.chunkSize)
		req.Body = io.NopCloser(body)
		req.Header.Set(httpx.HeaderContentRange, httpx.FormatContentRange(start, end, total))
		req.Header.Set(httpx.HeaderContentLength, strconv.FormatInt(r.chunkSize, 10))
		resp, err := u.httpClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
		start += r.chunkSize
		end += r.chunkSize
		if end >= total {
			end = total - 1
		}
	}
	return nil
}


func (r *UploadReq) Upload(filepath string) error {
	panic("not implemented")
}


func (r *UploadReq) ConcurrentUploadReaderChunked(reader io.ReaderAt, name string, total int64, concurrencyNum int) error {
	panic("not implemented")
}
