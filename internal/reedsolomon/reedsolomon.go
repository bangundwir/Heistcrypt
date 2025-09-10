package reedsolomon

import (
	"fmt"
	"io"
)

// ReedSolomon provides error correction capabilities
type ReedSolomon struct {
	dataShards   int
	parityShards int
	encoder      RSEncoder
}

// RSEncoder interface for Reed-Solomon encoding/decoding
type RSEncoder interface {
	Encode(data []byte) ([]byte, error)
	Decode(data []byte) ([]byte, error)
}

// SimpleRSEncoder is a basic implementation for demonstration
// In production, you would use a proper Reed-Solomon library
type SimpleRSEncoder struct {
	dataShards   int
	parityShards int
}

// NewSimpleRSEncoder creates a new simple Reed-Solomon encoder
func NewSimpleRSEncoder(dataShards, parityShards int) *SimpleRSEncoder {
	return &SimpleRSEncoder{
		dataShards:   dataShards,
		parityShards: parityShards,
	}
}

// Encode adds Reed-Solomon parity data
func (rs *SimpleRSEncoder) Encode(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	
	// Simple XOR-based parity for demonstration
	// In production, use a proper Reed-Solomon implementation
	chunkSize := 128
	result := make([]byte, 0, len(data)*138/128) // ~8 bytes per 128 bytes
	
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		
		chunk := data[i:end]
		result = append(result, chunk...)
		
		// Add simple parity bytes (8 bytes per 128-byte chunk)
		parity := rs.generateParity(chunk)
		result = append(result, parity...)
	}
	
	return result, nil
}

// Decode attempts to correct errors and recover original data
func (rs *SimpleRSEncoder) Decode(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	
	// Calculate expected chunk size with parity
	chunkWithParity := 136 // 128 data + 8 parity
	result := make([]byte, 0, len(data)*128/136)
	
	for i := 0; i < len(data); i += chunkWithParity {
		end := i + chunkWithParity
		if end > len(data) {
			// Handle last incomplete chunk
			remaining := len(data) - i
			if remaining <= 8 {
				break // Only parity bytes left
			}
			
			dataLen := remaining - 8
			if dataLen > 0 {
				result = append(result, data[i:i+dataLen]...)
			}
			break
		}
		
		chunk := data[i : i+128]
		parity := data[i+128 : end]
		
		// Simple error detection/correction
		expectedParity := rs.generateParity(chunk)
		if rs.parityMatches(parity, expectedParity) {
			result = append(result, chunk...)
		} else {
			// Attempt simple correction
			corrected, err := rs.attemptCorrection(chunk, parity, expectedParity)
			if err != nil {
				// Use original data if correction fails
				result = append(result, chunk...)
			} else {
				result = append(result, corrected...)
			}
		}
	}
	
	return result, nil
}

// generateParity creates simple parity bytes for a chunk
func (rs *SimpleRSEncoder) generateParity(data []byte) []byte {
	parity := make([]byte, 8)
	
	for i, b := range data {
		parity[i%8] ^= b
	}
	
	// Add some redundancy
	for i := 0; i < 8; i++ {
		parity[i] ^= byte(len(data) + i)
	}
	
	return parity
}

// parityMatches checks if parity bytes match
func (rs *SimpleRSEncoder) parityMatches(actual, expected []byte) bool {
	if len(actual) != len(expected) {
		return false
	}
	
	for i := range actual {
		if actual[i] != expected[i] {
			return false
		}
	}
	return true
}

// attemptCorrection tries to correct single-bit errors
func (rs *SimpleRSEncoder) attemptCorrection(data, actualParity, expectedParity []byte) ([]byte, error) {
	// Simple single-bit error correction attempt
	corrected := make([]byte, len(data))
	copy(corrected, data)
	
	// Try flipping each bit to see if it fixes the parity
	for i := 0; i < len(data); i++ {
		for bit := 0; bit < 8; bit++ {
			// Flip bit
			corrected[i] ^= (1 << bit)
			
			// Check if parity is now correct
			testParity := rs.generateParity(corrected)
			if rs.parityMatches(testParity, expectedParity) {
				return corrected, nil
			}
			
			// Flip bit back
			corrected[i] ^= (1 << bit)
		}
	}
	
	return nil, fmt.Errorf("unable to correct errors")
}

// New creates a new Reed-Solomon encoder
func New(dataShards, parityShards int) *ReedSolomon {
	return &ReedSolomon{
		dataShards:   dataShards,
		parityShards: parityShards,
		encoder:      NewSimpleRSEncoder(dataShards, parityShards),
	}
}

// EncodeStream encodes data from reader to writer with Reed-Solomon error correction
func (rs *ReedSolomon) EncodeStream(src io.Reader, dst io.Writer) error {
	buffer := make([]byte, 32*1024) // 32KB buffer
	
	for {
		n, err := src.Read(buffer)
		if n > 0 {
			encoded, encErr := rs.encoder.Encode(buffer[:n])
			if encErr != nil {
				return fmt.Errorf("encode error: %w", encErr)
			}
			
			if _, writeErr := dst.Write(encoded); writeErr != nil {
				return fmt.Errorf("write error: %w", writeErr)
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}
	}
	
	return nil
}

// DecodeStream decodes data from reader to writer with error correction
func (rs *ReedSolomon) DecodeStream(src io.Reader, dst io.Writer) error {
	buffer := make([]byte, 32*1024*138/128) // Larger buffer for encoded data
	
	for {
		n, err := src.Read(buffer)
		if n > 0 {
			decoded, decErr := rs.encoder.Decode(buffer[:n])
			if decErr != nil {
				return fmt.Errorf("decode error: %w", decErr)
			}
			
			if _, writeErr := dst.Write(decoded); writeErr != nil {
				return fmt.Errorf("write error: %w", writeErr)
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}
	}
	
	return nil
}
