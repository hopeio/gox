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
	"strings"

	"github.com/hopeio/gox/log"
	httpx "github.com/hopeio/gox/net/http"
	stringsx "github.com/hopeio/gox/strings"
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
	chunkSize int
}

type Multipart struct {
	Param       string
	Name        string
	ContentType string
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
	return NewMultipart(param, path.Base(f.Path), contentType, f.File)
}

func NewMultipart(param, name, contentType string, reader io.Reader) *Multipart {
	return &Multipart{
		Param:       param,
		Name:        name,
		ContentType: contentType,
		Reader:      reader,
	}
}

func (f *Multipart) setHeader(header textproto.MIMEHeader) {
	var contentDispositionValue string
	if stringsx.IsEmpty(f.Name) {
		contentDispositionValue = fmt.Sprintf(httpx.FormDataFieldTmpl, escapeQuotes(f.Param))
	} else {
		contentDispositionValue = fmt.Sprintf(httpx.FormDataFileTmpl,
			escapeQuotes(f.Param), escapeQuotes(f.Name))
	}
	header.Set(httpx.HeaderContentDisposition, contentDispositionValue)

	if !stringsx.IsEmpty(f.ContentType) {
		header.Set(ContentTypeKey, f.ContentType)
	}
}

func NewUploadReq(url string) *UploadReq {
	return &UploadReq{
		ctx:      context.Background(),
		Url:      url,
		uploader: DefaultUploader,
	}
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

func (r *UploadReq) ChunkSize(chunkSize int) *UploadReq {
	if chunkSize < 512 {
		panic("buffer size should > 512")
	}
	r.chunkSize = chunkSize
	return r
}

func (r *UploadReq) UploadMultipart(formData map[string]string, files ...*Multipart) error {
	body := bufPool.Get().(*bytes.Buffer)
	w := multipart.NewWriter(body)

	if r.boundary != "" {
		if err := w.SetBoundary(r.boundary); err != nil {
			return err
		}
	}
	header := make(textproto.MIMEHeader)
	for k, v := range formData {
		header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.FormDataFieldTmpl, escapeQuotes(k)))
		part, err := w.CreatePart(header)
		if err != nil {
			return err
		}
		_, err = part.Write([]byte(v))
		if err != nil {
			return err
		}
	}

	for _, file := range files {
		if file.ContentType == "" {
			cbuf := make([]byte, 512)
			size, err := file.Reader.Read(cbuf)
			if err != nil && err != io.EOF {
				return err
			}
			file.ContentType = http.DetectContentType(cbuf[:size])
			file.setHeader(header)
			partWriter, err := w.CreatePart(header)
			if err != nil {
				return err
			}
			if _, err = partWriter.Write(cbuf[:size]); err != nil {
				return err
			}

			_, err = io.Copy(partWriter, file.Reader)
			if err != nil {
				return err
			}
		} else {
			file.setHeader(header)
			partWriter, err := w.CreatePart(header)
			if err != nil {
				return err
			}

			_, err = io.Copy(partWriter, file.Reader)
			if err != nil {
				return err
			}
		}
	}

	req, err := http.NewRequest(http.MethodPost, r.Url, body)
	if err != nil {
		return err
	}
	if r.header != nil {
		req.Header = r.header
	}
	r.header.Set(httpx.HeaderContentType, w.FormDataContentType())
	err = w.Close()
	if err != nil {
		return err
	}
	d := r.uploader
	httpx.CopyHttpHeader(req.Header, d.header)
	for _, opt := range d.httpRequestOptions {
		opt(req)
	}
	_, err = d.httpClient.Do(req)
	if err != nil {
		return err
	}
	// TODO: error handler, retry
	return nil
}

// 默认单文件
func (r *UploadReq) UploadMultipartChunked(formData map[string]string, file Multipart) error {
	u := r.uploader
	body := bufPool.Get().(*bytes.Buffer)
	w := multipart.NewWriter(body)
	var start, total int64
	var end int64 = -1

	req, err := http.NewRequest(http.MethodPost, r.Url, nil)
	if err != nil {
		return err
	}

	if r.boundary != "" {
		if err := w.SetBoundary(r.boundary); err != nil {
			return err
		}
	}
	header := make(textproto.MIMEHeader)
	for k, v := range formData {
		header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.FormDataFieldTmpl, escapeQuotes(k)))
		part, err := w.CreatePart(header)
		if err != nil {
			return err
		}
		_, err = part.Write([]byte(v))
		if err != nil {
			return err
		}
	}
	fieldSize := body.Len()
	for {
		body.Reset()
		body.Truncate(fieldSize)
		buf := make([]byte, chunkSize)
		size, er := file.Reader.Read(buf)
		if er != nil && er != io.EOF {
			return er
		}
		if size > 0 {
			if file.ContentType == "" {
				file.ContentType = http.DetectContentType(buf[:min(size, 512)])
			}

			file.setHeader(header)
			partWriter, err := w.CreatePart(header)
			if err != nil {
				return err
			}
			if _, err = partWriter.Write(buf[:size]); err != nil {
				return err
			}

			end += int64(size)
			if er == io.EOF {
				total = end + 1
			}
			req.Body = io.NopCloser(bytes.NewReader(buf[0:size]))
			req.Header.Set(httpx.HeaderContentRange, httpx.FormatContentRange(start, end, total))
			resp, err := u.httpClient.Do(req)
			if err != nil {
				return err
			}
			resp.Body.Close()
		}
		if er == io.EOF {
			return nil
		}
	}
}

func (r *UploadReq) UploadStream(oReader io.Reader) error {
	u := r.uploader

	// 创建一个HTTP请求
	req, err := http.NewRequest(http.MethodPost, r.Url, nil)
	if err != nil {
		return err
	}
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	req.Header.Set(httpx.HeaderTransferEncoding, httpx.HeaderTransferEncodingChunked)
	// 使用io.Pipe创建一个管道，用于流式传输文件内容
	reader, writer := io.Pipe()

	// 创建一个goroutine来读取文件内容并写入管道
	go func() {
		defer writer.Close()
		_, err = io.Copy(writer, oReader)
		if err != nil && err != io.EOF {
			log.Error("error copying file to pipe: ", err)
		}
	}()

	// 将管道的读取端作为请求体发送到服务器
	req.Body = reader

	// 发送请求
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (r *UploadReq) UploadRaw(reader io.Reader, name string) error {
	u := r.uploader

	req, err := http.NewRequest(http.MethodPost, r.Url, reader)
	if err != nil {
		return err
	}
	if r.header != nil {
		req.Header = r.header
	}
	httpx.CopyHttpHeader(req.Header, u.header)
	name = escapeQuotes(name)
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	req.Header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.FormDataFileTmpl,
		name, name))
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (r *UploadReq) UploadRawChunked(reader io.Reader, name string) error {

	var start, total int64
	var end int64 = -1

	u := r.uploader
	buf := make([]byte, r.chunkSize)
	req, err := http.NewRequest(http.MethodPost, r.Url, nil)
	if err != nil {
		return err
	}
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeOctetStream)
	name = escapeQuotes(name)
	req.Header.Set(httpx.HeaderContentDisposition, fmt.Sprintf(httpx.FormDataFileTmpl,
		name, name))
	for {
		nr, er := reader.Read(buf)
		if er != nil && er != io.EOF {
			return er
		}
		if nr > 0 {
			end += int64(nr)
			if er == io.EOF {
				total = end + 1
			}
			req.Body = io.NopCloser(bytes.NewReader(buf[0:nr]))
			req.Header.Set(httpx.HeaderContentRange, httpx.FormatContentRange(start, end, total))
			resp, err := u.httpClient.Do(req)
			if err != nil {
				return err
			}
			resp.Body.Close()
		}
		if er == io.EOF {
			return nil
		}
	}
}
