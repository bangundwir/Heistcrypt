package keyfiles

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Keyfile represents a single keyfile
type Keyfile struct {
	Path string
	Hash [32]byte
}

// KeyfileManager manages multiple keyfiles
type KeyfileManager struct {
	Keyfiles    []Keyfile
	RequireOrder bool
}

// NewKeyfileManager creates a new keyfile manager
func NewKeyfileManager() *KeyfileManager {
	return &KeyfileManager{
		Keyfiles:    []Keyfile{},
		RequireOrder: false,
	}
}

// AddKeyfile adds a keyfile to the manager
func (km *KeyfileManager) AddKeyfile(path string) error {
	// Read and hash the keyfile
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open keyfile %s: %w", path, err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("hash keyfile %s: %w", path, err)
	}

	var hash [32]byte
	copy(hash[:], hasher.Sum(nil))

	keyfile := Keyfile{
		Path: path,
		Hash: hash,
	}

	km.Keyfiles = append(km.Keyfiles, keyfile)
	return nil
}

// RemoveKeyfile removes a keyfile by path
func (km *KeyfileManager) RemoveKeyfile(path string) {
	for i, kf := range km.Keyfiles {
		if kf.Path == path {
			km.Keyfiles = append(km.Keyfiles[:i], km.Keyfiles[i+1:]...)
			break
		}
	}
}

// Clear removes all keyfiles
func (km *KeyfileManager) Clear() {
	km.Keyfiles = []Keyfile{}
}

// GetCombinedKey returns a combined key from all keyfiles and password
func (km *KeyfileManager) GetCombinedKey(password []byte) []byte {
	hasher := sha256.New()
	
	// Add password first
	hasher.Write(password)
	
	// Add keyfiles in order (order matters if RequireOrder is true)
	for i, kf := range km.Keyfiles {
		hasher.Write(kf.Hash[:])
		
		// If order is required, include position in hash
		if km.RequireOrder {
			hasher.Write([]byte{byte(i)})
		}
	}
	
	return hasher.Sum(nil)
}

// MoveKeyfile moves a keyfile to a different position
func (km *KeyfileManager) MoveKeyfile(fromIndex, toIndex int) {
	if fromIndex < 0 || fromIndex >= len(km.Keyfiles) ||
		toIndex < 0 || toIndex >= len(km.Keyfiles) {
		return
	}
	
	// Remove from old position
	keyfile := km.Keyfiles[fromIndex]
	km.Keyfiles = append(km.Keyfiles[:fromIndex], km.Keyfiles[fromIndex+1:]...)
	
	// Insert at new position
	if toIndex > fromIndex {
		toIndex-- // Adjust for removed element
	}
	
	km.Keyfiles = append(km.Keyfiles[:toIndex], 
		append([]Keyfile{keyfile}, km.Keyfiles[toIndex:]...)...)
}

// HasKeyfiles returns true if any keyfiles are loaded
func (km *KeyfileManager) HasKeyfiles() bool {
	return len(km.Keyfiles) > 0
}

// Count returns the number of keyfiles
func (km *KeyfileManager) Count() int {
	return len(km.Keyfiles)
}

// GetPaths returns all keyfile paths
func (km *KeyfileManager) GetPaths() []string {
	paths := make([]string, len(km.Keyfiles))
	for i, kf := range km.Keyfiles {
		paths[i] = kf.Path
	}
	return paths
}

// GenerateKeyfile creates a secure keyfile with random data
func GenerateKeyfile(outputPath string, sizeKB int) error {
	if sizeKB <= 0 {
		sizeKB = 1 // Default 1KB
	}
	
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	
	// Create the keyfile
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create keyfile: %w", err)
	}
	defer file.Close()
	
	// Generate random data
	buffer := make([]byte, 1024) // 1KB buffer
	for i := 0; i < sizeKB; i++ {
		if _, err := rand.Read(buffer); err != nil {
			return fmt.Errorf("generate random data: %w", err)
		}
		if _, err := file.Write(buffer); err != nil {
			return fmt.Errorf("write keyfile data: %w", err)
		}
	}
	
	return nil
}

// ValidateKeyfile checks if a file can be used as a keyfile
func ValidateKeyfile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat keyfile: %w", err)
	}
	
	if info.IsDir() {
		return fmt.Errorf("keyfile cannot be a directory")
	}
	
	if info.Size() == 0 {
		return fmt.Errorf("keyfile cannot be empty")
	}
	
	// Try to read the file to ensure it's accessible
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open keyfile: %w", err)
	}
	file.Close()
	
	return nil
}
