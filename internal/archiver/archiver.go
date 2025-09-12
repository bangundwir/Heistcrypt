package archiver

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"bytes"
)

// ProgressCallback reports processed and total bytes during archiving
type ProgressCallback func(processed int64, total int64)

// CreateTarGz creates a compressed tar archive from a directory
func CreateTarGz(sourceDir, targetFile string, onProgress ProgressCallback) error {
	// Calculate total size first for progress reporting
	totalSize, err := calculateDirSize(sourceDir)
	if err != nil {
		return fmt.Errorf("calculate directory size: %w", err)
	}

	// Create the target file
	file, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("create target file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	processed := int64(0)

	// Walk through the source directory
	err = filepath.Walk(sourceDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return fmt.Errorf("create tar header for %s: %w", filePath, err)
		}

		// Set the name to be relative to the source directory
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}
		header.Name = filepath.ToSlash(relPath)

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("write tar header: %w", err)
		}

		// If it's a regular file, write its content
		if fileInfo.Mode().IsRegular() {
			srcFile, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("open source file %s: %w", filePath, err)
			}
			defer srcFile.Close()

			// Copy file content with progress reporting
			buf := make([]byte, 32*1024) // 32KB buffer
			for {
				n, err := srcFile.Read(buf)
				if n > 0 {
					if _, writeErr := tarWriter.Write(buf[:n]); writeErr != nil {
						return fmt.Errorf("write to tar: %w", writeErr)
					}
					processed += int64(n)
					if onProgress != nil {
						onProgress(processed, totalSize)
					}
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("read from source file: %w", err)
				}
			}
		}

		return nil
	})

	return err
}

// ExtractTarGz extracts a compressed tar archive to a directory
func ExtractTarGz(sourceFile, targetDir string, onProgress ProgressCallback) error {
	// Open the source file
	file, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer file.Close()

	// Get file size for progress reporting
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	totalSize := fileInfo.Size()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	processed := int64(0)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar header: %w", err)
		}

		targetPath := filepath.Join(targetDir, header.Name)

		// Ensure the target directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("create target directory: %w", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// Create regular file
			targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create target file %s: %w", targetPath, err)
			}

			// Copy file content with progress reporting
			buf := make([]byte, 32*1024) // 32KB buffer
			for {
				n, err := tarReader.Read(buf)
				if n > 0 {
					if _, writeErr := targetFile.Write(buf[:n]); writeErr != nil {
						targetFile.Close()
						return fmt.Errorf("write to target file: %w", writeErr)
					}
					processed += int64(n)
					if onProgress != nil {
						onProgress(processed, totalSize)
					}
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					targetFile.Close()
					return fmt.Errorf("read from tar: %w", err)
				}
			}
			targetFile.Close()
		}
	}

	return nil
}

// calculateDirSize calculates the total size of all files in a directory
func calculateDirSize(dirPath string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// IsArchive checks if a file is a tar.gz archive based on its extension
func IsArchive(filename string) bool {
	f, err := os.Open(filename)
	if err != nil { return false }
	defer f.Close()
	// Read first few bytes for gzip magic 1F 8B
	hdr := make([]byte, 3)
	if _, err := io.ReadFull(f, hdr); err != nil { return false }
	if hdr[0] != 0x1F || hdr[1] != 0x8B { return false }
	// Reset and try to create gzip reader
	f.Seek(0,0)
	gz, err := gzip.NewReader(f)
	if err != nil { return false }
	defer gz.Close()
	// Read first tar header block (512 bytes)
	block := make([]byte, 512)
	if _, err := io.ReadFull(gz, block); err != nil { return false }
	// Basic heuristic: tar header contains a null-terminated name and ustar or gnu signature optionally
	if bytes.Contains(block, []byte("ustar")) || bytes.IndexByte(block, 0) > 0 {
		return true
	}
	return false
}
