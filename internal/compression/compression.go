package compression

import (
	"compress/flate"
	"fmt"
	"io"
)

// CompressionLevel represents compression levels
type CompressionLevel int

const (
	NoCompression      CompressionLevel = flate.NoCompression
	BestSpeed         CompressionLevel = flate.BestSpeed
	BestCompression   CompressionLevel = flate.BestCompression
	DefaultCompression CompressionLevel = flate.DefaultCompression
)

// Compressor handles data compression
type Compressor struct {
	level CompressionLevel
}

// NewCompressor creates a new compressor with specified level
func NewCompressor(level CompressionLevel) *Compressor {
	return &Compressor{
		level: level,
	}
}

// CompressStream compresses data from reader to writer
func (c *Compressor) CompressStream(src io.Reader, dst io.Writer) error {
	// Create flate writer
	writer, err := flate.NewWriter(dst, int(c.level))
	if err != nil {
		return fmt.Errorf("create flate writer: %w", err)
	}
	defer writer.Close()

	// Copy and compress data
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, readErr := src.Read(buffer)
		if n > 0 {
			if _, writeErr := writer.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("write compressed data: %w", writeErr)
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("read source data: %w", readErr)
		}
	}

	return nil
}

// DecompressStream decompresses data from reader to writer
func (c *Compressor) DecompressStream(src io.Reader, dst io.Writer) error {
	// Create flate reader
	reader := flate.NewReader(src)
	defer reader.Close()

	// Copy and decompress data
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, readErr := reader.Read(buffer)
		if n > 0 {
			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("write decompressed data: %w", writeErr)
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("read compressed data: %w", readErr)
		}
	}

	return nil
}

// CompressBytes compresses a byte slice
func (c *Compressor) CompressBytes(data []byte) ([]byte, error) {
	var compressed []byte
	
	// Use a buffer to capture compressed output
	writer, err := flate.NewWriter(&bytesWriter{&compressed}, int(c.level))
	if err != nil {
		return nil, fmt.Errorf("create flate writer: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, fmt.Errorf("write data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer: %w", err)
	}

	return compressed, nil
}

// DecompressBytes decompresses a byte slice
func (c *Compressor) DecompressBytes(data []byte) ([]byte, error) {
	reader := flate.NewReader(&bytesReader{data, 0})
	defer reader.Close()

	var decompressed []byte
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			decompressed = append(decompressed, buffer[:n]...)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read compressed data: %w", err)
		}
	}

	return decompressed, nil
}

// EstimateCompressionRatio estimates compression ratio for given data
func (c *Compressor) EstimateCompressionRatio(data []byte) float64 {
	if len(data) == 0 {
		return 1.0
	}

	compressed, err := c.CompressBytes(data)
	if err != nil {
		return 1.0 // No compression if error
	}

	return float64(len(compressed)) / float64(len(data))
}

// bytesWriter implements io.Writer for byte slices
type bytesWriter struct {
	data *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.data = append(*w.data, p...)
	return len(p), nil
}

// bytesReader implements io.Reader for byte slices
type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// DefaultCompressor returns a compressor with default settings
func DefaultCompressor() *Compressor {
	return NewCompressor(DefaultCompression)
}

// FastCompressor returns a compressor optimized for speed
func FastCompressor() *Compressor {
	return NewCompressor(BestSpeed)
}

// BestCompressor returns a compressor optimized for size
func BestCompressor() *Compressor {
	return NewCompressor(BestCompression)
}
