/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"

	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/os/fs"
)

func Upload(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		bodyfile, fileHeader, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "failed to get file", http.StatusInternalServerError)
			return
		}
		if fs.Exist(dir + fileHeader.Filename) {
			fmt.Fprintf(w, "ok", http.StatusOK)
			return
		}
		// 解析Range头部
		rangeHeader := r.Header.Get(httpx.HeaderContentRange)
		if rangeHeader == "" {
			http.Error(w, "missing Content-Range header", http.StatusBadRequest)
			return
		}

		start, end, total, err := httpx.ParseContentRange(rangeHeader)
		if err != nil {
			http.Error(w, "failed to parse range header", http.StatusBadRequest)
			return
		}

		// 打开文件准备写入，使用O_RDWR | O_CREATE | O_APPEND以追加模式打开
		file, err := os.OpenFile(dir+fileHeader.Filename+".uploading", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			http.Error(w, "failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		// 移动文件指针到开始位置
		_, err = file.Seek(start, io.SeekStart)
		if err != nil {
			http.Error(w, "failed to seek file", http.StatusInternalServerError)
			return
		}

		// 读取请求体并写入文件
		_, err = io.CopyN(file, bodyfile, end-start+1)
		if err != nil && err != io.EOF {
			http.Error(w, "failed to write to file", http.StatusInternalServerError)
			return
		}
		stats, err := file.Stat()
		if err != nil {
			http.Error(w, "failed to get file size", http.StatusInternalServerError)
			return
		}
		if stats.Size() == total {
			// 文件大小等于总大小，说明上传完成，重命名文件
			err = os.Rename(dir+fileHeader.Filename+".uploading", dir+fileHeader.Filename)
			if err != nil {
				http.Error(w, "failed to rename file", http.StatusInternalServerError)
				return
			}
		}

		// 如果一切顺利，发送成功的响应
		w.Header().Set(httpx.HeaderContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, stats.Size()))
		w.WriteHeader(http.StatusPartialContent)
		fmt.Fprintf(w, "Uploaded chunk from byte %d to %d", start, end)

	}

}
