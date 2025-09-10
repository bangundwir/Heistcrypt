package splitter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SizeUnit represents different size units
type SizeUnit string

const (
	UnitKiB SizeUnit = "KiB"
	UnitMiB SizeUnit = "MiB"
	UnitGiB SizeUnit = "GiB"
	UnitTiB SizeUnit = "TiB"
)

// ConvertToBytes converts size with unit to bytes
func ConvertToBytes(size int, unit SizeUnit) int64 {
	switch unit {
	case UnitKiB:
		return int64(size) * 1024
	case UnitMiB:
		return int64(size) * 1024 * 1024
	case UnitGiB:
		return int64(size) * 1024 * 1024 * 1024
	case UnitTiB:
		return int64(size) * 1024 * 1024 * 1024 * 1024
	default:
		return int64(size)
	}
}

// ProgressCallback reports splitting progress
type ProgressCallback func(processed int64, total int64)

// SplitFile splits a large file into smaller chunks
func SplitFile(inputPath string, chunkSize int64, onProgress ProgressCallback) ([]string, error) {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("open input file: %w", err)
	}
	defer inputFile.Close()

	// Get file info
	fileInfo, err := inputFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat input file: %w", err)
	}
	totalSize := fileInfo.Size()

	// If file is smaller than chunk size, no need to split
	if totalSize <= chunkSize {
		return []string{inputPath}, nil
	}

	var chunkPaths []string
	buffer := make([]byte, 64*1024) // 64KB buffer for copying
	processed := int64(0)
	chunkIndex := 0

	for processed < totalSize {
		// Create chunk file path
		chunkPath := fmt.Sprintf("%s.%03d", inputPath, chunkIndex)
		chunkPaths = append(chunkPaths, chunkPath)

		// Create chunk file
		chunkFile, err := os.Create(chunkPath)
		if err != nil {
			return nil, fmt.Errorf("create chunk file %s: %w", chunkPath, err)
		}

		// Copy data to chunk
		chunkWritten := int64(0)
		for chunkWritten < chunkSize && processed < totalSize {
			// Determine how much to read
			toRead := int64(len(buffer))
			if chunkWritten+toRead > chunkSize {
				toRead = chunkSize - chunkWritten
			}
			if processed+toRead > totalSize {
				toRead = totalSize - processed
			}

			// Read from input
			n, readErr := inputFile.Read(buffer[:toRead])
			if n > 0 {
				// Write to chunk
				if _, writeErr := chunkFile.Write(buffer[:n]); writeErr != nil {
					chunkFile.Close()
					return nil, fmt.Errorf("write to chunk file: %w", writeErr)
				}
				chunkWritten += int64(n)
				processed += int64(n)

				// Report progress
				if onProgress != nil {
					onProgress(processed, totalSize)
				}
			}

			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				chunkFile.Close()
				return nil, fmt.Errorf("read from input file: %w", readErr)
			}
		}

		chunkFile.Close()
		chunkIndex++
	}

	return chunkPaths, nil
}

// CombineFiles combines split chunks back into original file
func CombineFiles(chunkPaths []string, outputPath string, onProgress ProgressCallback) error {
	// Calculate total size
	var totalSize int64
	for _, chunkPath := range chunkPaths {
		if info, err := os.Stat(chunkPath); err == nil {
			totalSize += info.Size()
		}
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer outputFile.Close()

	buffer := make([]byte, 64*1024) // 64KB buffer
	processed := int64(0)

	// Combine chunks
	for _, chunkPath := range chunkPaths {
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("open chunk file %s: %w", chunkPath, err)
		}

		// Copy chunk to output
		for {
			n, readErr := chunkFile.Read(buffer)
			if n > 0 {
				if _, writeErr := outputFile.Write(buffer[:n]); writeErr != nil {
					chunkFile.Close()
					return fmt.Errorf("write to output file: %w", writeErr)
				}
				processed += int64(n)

				// Report progress
				if onProgress != nil {
					onProgress(processed, totalSize)
				}
			}

			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				chunkFile.Close()
				return fmt.Errorf("read from chunk file: %w", readErr)
			}
		}

		chunkFile.Close()
	}

	return nil
}

// FindChunks finds all chunks for a given base file path
func FindChunks(basePath string) ([]string, error) {
	dir := filepath.Dir(basePath)
	baseName := filepath.Base(basePath)
	
	var chunks []string
	
	// Look for files with pattern: basename.000, basename.001, etc.
	for i := 0; i < 9999; i++ {
		chunkName := fmt.Sprintf("%s.%03d", baseName, i)
		chunkPath := filepath.Join(dir, chunkName)
		
		if _, err := os.Stat(chunkPath); err == nil {
			chunks = append(chunks, chunkPath)
		} else {
			break // No more chunks
		}
	}
	
	return chunks, nil
}

// IsChunkFile checks if a file appears to be a chunk file
func IsChunkFile(filePath string) bool {
	base := filepath.Base(filePath)
	
	// Look for pattern: filename.000, filename.001, etc.
	if len(base) < 4 {
		return false
	}
	
	// Check if ends with .XXX where XXX is 3 digits
	suffix := base[len(base)-4:]
	if suffix[0] != '.' {
		return false
	}
	
	// Check if last 3 characters are digits
	for i := 1; i < 4; i++ {
		if suffix[i] < '0' || suffix[i] > '9' {
			return false
		}
	}
	
	return true
}

// GetBasePathFromChunk returns the base file path from a chunk file path
func GetBasePathFromChunk(chunkPath string) string {
	if !IsChunkFile(chunkPath) {
		return chunkPath
	}
	
	// Remove the .XXX suffix
	return chunkPath[:len(chunkPath)-4]
}

// DeleteChunks deletes all chunk files for a given base path
func DeleteChunks(basePath string) error {
	chunks, err := FindChunks(basePath)
	if err != nil {
		return err
	}
	
	for _, chunkPath := range chunks {
		if err := os.Remove(chunkPath); err != nil {
			return fmt.Errorf("delete chunk %s: %w", chunkPath, err)
		}
	}
	
	return nil
}

// GetChunkInfo returns information about chunks
func GetChunkInfo(basePath string) (int, int64, error) {
	chunks, err := FindChunks(basePath)
	if err != nil {
		return 0, 0, err
	}
	
	var totalSize int64
	for _, chunkPath := range chunks {
		if info, err := os.Stat(chunkPath); err == nil {
			totalSize += info.Size()
		}
	}
	
	return len(chunks), totalSize, nil
}
