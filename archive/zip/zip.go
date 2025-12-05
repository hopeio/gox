package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func CompressDir(sourceDir, targetZip string, containRootDir bool) error {
	// 创建ZIP文件
	zipFile, err := os.Create(targetZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	rootDir := filepath.Base(sourceDir)
	// 遍历目录内容
	return filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径（排除顶层目录）
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}
		var zipPath string
		if containRootDir {
			// 构建ZIP内部路径
			zipPath = filepath.Join(rootDir, relPath)
			zipPath = filepath.ToSlash(zipPath) // 统一为斜杠路径
		} else {
			zipPath = filepath.ToSlash(relPath)
		}

		// 跳过源目录自身（relPath == "." 时表示根目录）
		if relPath == "." {
			if !containRootDir {
				return nil
			}
			zipPath = rootDir
		}

		// 创建文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipPath

		// 设置压缩方法
		if info.IsDir() {
			header.Name += "/"        // 目录需要尾部斜杠
			header.Method = zip.Store // 目录不压缩
		} else {
			header.Method = zip.Deflate // 文件压缩
		}

		// 写入文件头
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// 如果是文件则写入内容
		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
