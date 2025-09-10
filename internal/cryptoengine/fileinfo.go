package cryptoengine

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// ExtractCommentsFromFile extracts comments from an encrypted file header
func ExtractCommentsFromFile(inputPath string) (string, error) {
	in, err := os.Open(inputPath)
	if err != nil {
		return "", err
	}
	defer in.Close()

	// Read and validate header
	header := make([]byte, 4)
	if _, err := io.ReadFull(in, header); err != nil {
		return "", err
	}
	if string(header) != fileMagic {
		return "", fmt.Errorf("not a HadesCrypt file")
	}

	ver := make([]byte, 1)
	if _, err := io.ReadFull(in, ver); err != nil {
		return "", err
	}
	if ver[0] != fileVersion {
		return "", fmt.Errorf("unsupported version: %d", ver[0])
	}

	// Read encryption mode (skip it)
	modeBytes := make([]byte, 1)
	if _, err := io.ReadFull(in, modeBytes); err != nil {
		return "", err
	}

	// Skip salt and nonce prefix
	salt := make([]byte, saltLengthBytes)
	if _, err := io.ReadFull(in, salt); err != nil {
		return "", err
	}
	noncePrefix := make([]byte, noncePrefixLen)
	if _, err := io.ReadFull(in, noncePrefix); err != nil {
		return "", err
	}

	// Skip chunk size and total size
	var tmp4 [4]byte
	if _, err := io.ReadFull(in, tmp4[:]); err != nil {
		return "", err
	}
	var tmp8 [8]byte
	if _, err := io.ReadFull(in, tmp8[:]); err != nil {
		return "", err
	}

	// Read comment length
	var commentLen [4]byte
	if _, err := io.ReadFull(in, commentLen[:]); err != nil {
		return "", err
	}
	commentLength := binary.BigEndian.Uint32(commentLen[:])

	if commentLength == 0 {
		return "", nil // No comments
	}

	if commentLength > 1024*1024 { // Sanity check: max 1MB comment
		return "", fmt.Errorf("comment too large: %d bytes", commentLength)
	}

	// Read comment
	commentBytes := make([]byte, commentLength)
	if _, err := io.ReadFull(in, commentBytes); err != nil {
		return "", err
	}

	return string(commentBytes), nil
}

// ExtractEncryptionModeFromFile extracts the encryption mode from a HadesCrypt file
func ExtractEncryptionModeFromFile(inputPath string) (EncryptionMode, error) {
	in, err := os.Open(inputPath)
	if err != nil {
		return ModeAES256GCM, err
	}
	defer in.Close()

	// Read and validate header
	header := make([]byte, 4)
	if _, err := io.ReadFull(in, header); err != nil {
		return ModeAES256GCM, err
	}
	if string(header) != fileMagic {
		return ModeAES256GCM, fmt.Errorf("not a HadesCrypt file")
	}

	ver := make([]byte, 1)
	if _, err := io.ReadFull(in, ver); err != nil {
		return ModeAES256GCM, err
	}
	if ver[0] != fileVersion {
		return ModeAES256GCM, fmt.Errorf("unsupported version: %d", ver[0])
	}

	// Read encryption mode
	modeBytes := make([]byte, 1)
	if _, err := io.ReadFull(in, modeBytes); err != nil {
		return ModeAES256GCM, err
	}
	
	return EncryptionMode(modeBytes[0]), nil
}

// GetFileInfo returns information about an encrypted file
func GetFileInfo(inputPath string) (map[string]interface{}, error) {
	info := make(map[string]interface{})
	
	// Get basic file info
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, err
	}
	
	info["size"] = fileInfo.Size()
	info["modified"] = fileInfo.ModTime()
	info["name"] = fileInfo.Name()
	
	// Try to extract HadesCrypt specific info
	comments, err := ExtractCommentsFromFile(inputPath)
	if err == nil {
		info["format"] = "HadesCrypt"
		info["comments"] = comments
		
		// Try to get encryption mode
		mode, modeErr := ExtractEncryptionModeFromFile(inputPath)
		if modeErr == nil {
			info["encryption_mode"] = mode
			info["encryption_mode_name"] = GetEncryptionModeName(mode)
		}
	} else {
		// Check if it's a GnuPG file
		if IsGnuPGFile(inputPath) {
			info["format"] = "GnuPG/OpenPGP"
			info["comments"] = "" // GnuPG doesn't store comments in the same way
			info["encryption_mode_name"] = "GnuPG/OpenPGP"
		} else {
			info["format"] = "Unknown"
			info["comments"] = ""
		}
	}
	
	return info, nil
}

// GetEncryptionModeName returns human-readable name for encryption mode
func GetEncryptionModeName(mode EncryptionMode) string {
	switch mode {
	case ModeAES256GCM:
		return "AES-256-GCM"
	case ModeChaCha20:
		return "ChaCha20-Poly1305"
	case ModeParanoid:
		return "Paranoid (AES-256 + ChaCha20)"
	case ModePostQuantumKyber768:
		return "Post-Quantum: Kyber-768"
	case ModePostQuantumDilithium3:
		return "Post-Quantum: Dilithium-3"
	case ModePostQuantumSPHINCS:
		return "Post-Quantum: SPHINCS+"
	case ModeGnuPG:
		return "GnuPG/OpenPGP"
	default:
		return "Unknown"
	}
}

// IsGnuPGFile checks if a file is in GnuPG/OpenPGP format
func IsGnuPGFile(filePath string) bool {
	// Check by extension first
	lowerPath := strings.ToLower(filePath)
	if strings.HasSuffix(lowerPath, ".gpg") || strings.HasSuffix(lowerPath, ".pgp") {
		return true
	}
	
	// Check by file content (OpenPGP magic bytes)
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()
	
	// Read first few bytes to check for OpenPGP format
	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		return false
	}
	
	// OpenPGP files typically start with specific packet headers
	if len(header) > 0 {
		firstByte := header[0]
		// Check for OpenPGP packet format (high bit set indicates OpenPGP packet)
		if (firstByte&0x80) != 0 {
			return true
		}
	}
	
	return false
}

// FormatFileSize formats file size in human readable format
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	
	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

// FormatTime formats time in user-friendly format
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
