package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// CompressDir performs the operation.
func CompressDir(sourceDir, targetZip string, containRootDir bool) error {
	// Create the ZIP file
	zipFile, err := os.Create(targetZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	rootDir := filepath.Base(sourceDir)
	// Walk directory contents
	return filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Relative path excluding the top directory
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}
		var zipPath string
		if containRootDir {
			// Build the path inside the ZIP
			zipPath = filepath.Join(rootDir, relPath)
			zipPath = filepath.ToSlash(zipPath) // Normalize to slash paths
		} else {
			zipPath = filepath.ToSlash(relPath)
		}

		// Skip the source directory itself (relPath == "." is the root)
		if relPath == "." {
			if !containRootDir {
				return nil
			}
			zipPath = rootDir
		}

		// Create the file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipPath

		// Set the compression method
		if info.IsDir() {
			header.Name += "/"        // Directories need a trailing slash
			header.Method = zip.Store // Do not compress directories
		} else {
			header.Method = zip.Deflate // Compress files
		}

		// Write the file header
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// If it is a file, write its contents
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
