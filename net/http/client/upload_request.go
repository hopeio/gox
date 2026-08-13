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
	defaultChunkSize = 5 * 1024 * 1024 // chunk size, here 5MB
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
	header    http.Header //request-level headers
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

// NewFile creates and returns a new instance.
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

// ToMutilPart converts the value.
func (f *File) ToMutilPart(param string) *Multipart {
	contentType := mime.TypeByExtension(filepath.Ext(f.Path))
	return NewMultipart(param, path.Base(f.Path), textproto.MIMEHeader{httpx.HeaderContentType: []string{contentType}}, f.File)
}

// NewMultipart creates and returns a new instance.
func NewMultipart(name, filename string, header textproto.MIMEHeader, reader io.Reader) *Multipart {
	return &Multipart{
		Name:     name,
		Filename: filename,
		Header:   header,
		Reader:   reader,
	}
}

// NewUploadReq creates and returns a new instance.
func NewUploadReq(url string) *UploadReq {
	return &UploadReq{
		ctx:      context.Background(),
		Url:      url,
		uploader: DefaultUploader,
	}
}

// Context returns the result.
func (r *UploadReq) Context(ctx context.Context) *UploadReq {
	r.ctx = ctx
	return r
}

// Uploader returns the result.
func (r *UploadReq) Uploader(u *Uploader) *UploadReq {
	r.uploader = u
	return r
}

// Boundary returns the result.
func (r *UploadReq) Boundary(boundary string) *UploadReq {
	r.boundary = boundary
	return r
}

// Mode returns the result.
func (r *UploadReq) Mode(mode UploadMode) *UploadReq {
	r.mode = mode
	return r
}

// ChunkSize returns the result.
func (r *UploadReq) ChunkSize(chunkSize int64) *UploadReq {
	if chunkSize < 512 {
		panic("buffer size should > 512")
	}
	r.chunkSize = chunkSize
	return r
}

// UploadMultipart performs the operation.
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
	req, err := http.NewRequestWithContext(r.ctx, http.MethodPost, r.Url, body)
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
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status:%s %s", resp.Status, string(data))
	}
	return nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// escapeQuotes returns the result.
func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// UploadReader performs the operation.
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

// UploadReaderChunked performs the operation.
func (r *UploadReq) UploadReaderChunked(reader io.ReaderAt, name string, total int64) error {
	var start int64
	u := r.uploader

	req, err := http.NewRequestWithContext(r.ctx, http.MethodPost, r.Url, nil)
	if err != nil {
		return err
	}
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	for start < total {
		// 末块长度不足 chunkSize：Content-Length 必须按实际字节数，
		// 否则服务端会一直等满长度导致挂起
		chunk := r.chunkSize
		if start+chunk > total {
			chunk = total - start
		}
		body := io.NewSectionReader(reader, start, chunk)
		req.Body = io.NopCloser(body)
		req.ContentLength = chunk
		req.Header.Set(httpx.HeaderContentRange, httpx.FormatContentRange(start, start+chunk-1, total))
		req.Header.Set(httpx.HeaderContentLength, strconv.FormatInt(chunk, 10))
		resp, err := u.httpClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("upload chunk %d-%d failed: %s", start, start+chunk-1, resp.Status)
		}
		start += chunk
	}
	return nil
}

// Upload performs the operation.
func (r *UploadReq) Upload(filepath string) error {
	panic("not implemented")
}

// ConcurrentUploadReaderChunked performs the operation.
func (r *UploadReq) ConcurrentUploadReaderChunked(reader io.ReaderAt, name string, total int64, concurrencyNum int) error {
	panic("not implemented")
}
