package gnupg

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// GnuPGCipher provides GnuPG encryption/decryption capabilities
type GnuPGCipher struct {
	gpgPath     string
	tempDir     string
	keyID       string
	passphrase  string
	initialized bool
}

// GnuPGOptions contains options for GnuPG operations
type GnuPGOptions struct {
	Cipher         string // AES256, AES192, AES128, TWOFISH, BLOWFISH, etc.
	Compression    string // ZIP, ZLIB, BZIP2, or none
	ArmorOutput    bool   // ASCII armored output
	UseSymmetric   bool   // Use symmetric encryption (password-based)
	KeyID          string // Key ID for asymmetric encryption
	TrustModel     string // pgp, classic, direct, always, auto
}

// DefaultGnuPGOptions returns sensible defaults
func DefaultGnuPGOptions() *GnuPGOptions {
	return &GnuPGOptions{
		Cipher:       "AES256",
		Compression:  "ZLIB",
		ArmorOutput:  false, // Binary output for HadesCrypt compatibility
		UseSymmetric: true,  // Password-based by default
		TrustModel:   "always",
	}
}

// NewGnuPGCipher creates a new GnuPG cipher instance
func NewGnuPGCipher() (*GnuPGCipher, error) {
	gpg := &GnuPGCipher{}
	
	// Find GPG executable
	var err error
	gpg.gpgPath, err = findGPGExecutable()
	if err != nil {
		return nil, fmt.Errorf("GPG not found: %w", err)
	}
	
	// Create temporary directory
	gpg.tempDir, err = os.MkdirTemp("", "hadescrypt_gpg_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	gpg.initialized = true
	return gpg, nil
}

// findGPGExecutable locates the GPG executable on the system
func findGPGExecutable() (string, error) {
	// Common GPG executable names
	candidates := []string{"gpg", "gpg2", "gpg.exe", "gpg2.exe"}
	
	// Try PATH first
	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	
	// Try common installation paths
	var commonPaths []string
	switch runtime.GOOS {
	case "windows":
		commonPaths = []string{
			`C:\Program Files\GnuPG\bin\gpg.exe`,
			`C:\Program Files (x86)\GnuPG\bin\gpg.exe`,
			`C:\Program Files\GNU\GnuPG\gpg.exe`,
			`C:\Program Files (x86)\GNU\GnuPG\gpg.exe`,
			`C:\msys64\usr\bin\gpg.exe`,
			`C:\Git\usr\bin\gpg.exe`,
		}
	case "darwin":
		commonPaths = []string{
			"/usr/local/bin/gpg",
			"/usr/local/bin/gpg2",
			"/opt/homebrew/bin/gpg",
			"/opt/homebrew/bin/gpg2",
			"/usr/bin/gpg",
		}
	case "linux":
		commonPaths = []string{
			"/usr/bin/gpg",
			"/usr/bin/gpg2",
			"/usr/local/bin/gpg",
			"/usr/local/bin/gpg2",
			"/snap/bin/gpg",
		}
	}
	
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("GPG executable not found")
}

// SetPassphrase sets the passphrase for symmetric encryption
func (g *GnuPGCipher) SetPassphrase(passphrase string) {
	g.passphrase = passphrase
}

// Cleanup removes temporary files and directories
func (g *GnuPGCipher) Cleanup() error {
	if g.tempDir != "" {
		return os.RemoveAll(g.tempDir)
	}
	return nil
}

// GetVersion returns the GPG version
func (g *GnuPGCipher) GetVersion() (string, error) {
	if !g.initialized {
		return "", fmt.Errorf("GnuPG cipher not initialized")
	}
	
	cmd := exec.Command(g.gpgPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GPG version: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	
	return "Unknown", nil
}

// ListCiphers returns available cipher algorithms
func (g *GnuPGCipher) ListCiphers() ([]string, error) {
	if !g.initialized {
		return nil, fmt.Errorf("GnuPG cipher not initialized")
	}
	
	cmd := exec.Command(g.gpgPath, "--with-colons", "--list-config", "cipher")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to common ciphers if listing fails
		return []string{"AES256", "AES192", "AES128", "TWOFISH", "BLOWFISH", "3DES"}, nil
	}
	
	var ciphers []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cfg:cipher:") {
			parts := strings.Split(line, ":")
			if len(parts) > 2 {
				ciphers = append(ciphers, strings.ToUpper(parts[2]))
			}
		}
	}
	
	if len(ciphers) == 0 {
		// Fallback
		return []string{"AES256", "AES192", "AES128", "TWOFISH", "BLOWFISH"}, nil
	}
	
	return ciphers, nil
}

// EncryptFile encrypts a file using GnuPG
func (g *GnuPGCipher) EncryptFile(inputPath, outputPath string, options *GnuPGOptions) error {
	if !g.initialized {
		return fmt.Errorf("GnuPG cipher not initialized")
	}
	
	if options == nil {
		options = DefaultGnuPGOptions()
	}
	
	// Build GPG command
	args := []string{
		"--batch",
		"--yes",
		"--quiet",
		"--cipher-algo", options.Cipher,
		"--compress-algo", options.Compression,
		"--trust-model", options.TrustModel,
	}
	
	if options.UseSymmetric {
		args = append(args, "--symmetric")
		if g.passphrase != "" {
			args = append(args, "--passphrase", g.passphrase)
		}
	} else if options.KeyID != "" {
		args = append(args, "--encrypt", "--recipient", options.KeyID)
	} else {
		return fmt.Errorf("either symmetric encryption or recipient key ID must be specified")
	}
	
	if options.ArmorOutput {
		args = append(args, "--armor")
	}
	
	args = append(args, "--output", outputPath, inputPath)
	
	// Execute GPG command
	cmd := exec.Command(g.gpgPath, args...)
	
	// Set environment for better security
	env := os.Environ()
	env = append(env, "GPG_TTY=")
	env = append(env, "DISPLAY=")
	cmd.Env = env
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("GPG encryption failed: %w, stderr: %s", err, stderr.String())
	}
	
	return nil
}

// DecryptFile decrypts a file using GnuPG
func (g *GnuPGCipher) DecryptFile(inputPath, outputPath string, options *GnuPGOptions) error {
	if !g.initialized {
		return fmt.Errorf("GnuPG cipher not initialized")
	}
	
	if options == nil {
		options = DefaultGnuPGOptions()
	}
	
	// Build GPG command
	args := []string{
		"--batch",
		"--yes",
		"--quiet",
		"--trust-model", options.TrustModel,
		"--decrypt",
	}
	
	if g.passphrase != "" {
		args = append(args, "--passphrase", g.passphrase)
	}
	
	args = append(args, "--output", outputPath, inputPath)
	
	// Execute GPG command
	cmd := exec.Command(g.gpgPath, args...)
	
	// Set environment
	env := os.Environ()
	env = append(env, "GPG_TTY=")
	env = append(env, "DISPLAY=")
	cmd.Env = env
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("GPG decryption failed: %w, stderr: %s", err, stderr.String())
	}
	
	return nil
}

// EncryptStream encrypts data from reader to writer
func (g *GnuPGCipher) EncryptStream(input io.Reader, output io.Writer, options *GnuPGOptions) error {
	if !g.initialized {
		return fmt.Errorf("GnuPG cipher not initialized")
	}
	
	if options == nil {
		options = DefaultGnuPGOptions()
	}
	
	// Create temporary files
	tempInput := filepath.Join(g.tempDir, fmt.Sprintf("input_%d", time.Now().UnixNano()))
	tempOutput := filepath.Join(g.tempDir, fmt.Sprintf("output_%d", time.Now().UnixNano()))
	
	defer os.Remove(tempInput)
	defer os.Remove(tempOutput)
	
	// Write input to temp file
	inputFile, err := os.Create(tempInput)
	if err != nil {
		return fmt.Errorf("failed to create temp input file: %w", err)
	}
	
	_, err = io.Copy(inputFile, input)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("failed to write input data: %w", err)
	}
	
	// Encrypt
	err = g.EncryptFile(tempInput, tempOutput, options)
	if err != nil {
		return err
	}
	
	// Read output
	outputFile, err := os.Open(tempOutput)
	if err != nil {
		return fmt.Errorf("failed to open encrypted output: %w", err)
	}
	defer outputFile.Close()
	
	_, err = io.Copy(output, outputFile)
	if err != nil {
		return fmt.Errorf("failed to write encrypted data: %w", err)
	}
	
	return nil
}

// DecryptStream decrypts data from reader to writer
func (g *GnuPGCipher) DecryptStream(input io.Reader, output io.Writer, options *GnuPGOptions) error {
	if !g.initialized {
		return fmt.Errorf("GnuPG cipher not initialized")
	}
	
	if options == nil {
		options = DefaultGnuPGOptions()
	}
	
	// Create temporary files
	tempInput := filepath.Join(g.tempDir, fmt.Sprintf("input_%d", time.Now().UnixNano()))
	tempOutput := filepath.Join(g.tempDir, fmt.Sprintf("output_%d", time.Now().UnixNano()))
	
	defer os.Remove(tempInput)
	defer os.Remove(tempOutput)
	
	// Write input to temp file
	inputFile, err := os.Create(tempInput)
	if err != nil {
		return fmt.Errorf("failed to create temp input file: %w", err)
	}
	
	_, err = io.Copy(inputFile, input)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("failed to write input data: %w", err)
	}
	
	// Decrypt
	err = g.DecryptFile(tempInput, tempOutput, options)
	if err != nil {
		return err
	}
	
	// Read output
	outputFile, err := os.Open(tempOutput)
	if err != nil {
		return fmt.Errorf("failed to open decrypted output: %w", err)
	}
	defer outputFile.Close()
	
	_, err = io.Copy(output, outputFile)
	if err != nil {
		return fmt.Errorf("failed to write decrypted data: %w", err)
	}
	
	return nil
}

// GenerateRandomData generates cryptographically secure random data
func (g *GnuPGCipher) GenerateRandomData(size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := rand.Read(data)
	return data, err
}

// IsAvailable checks if GnuPG is available on the system
func IsAvailable() bool {
	_, err := findGPGExecutable()
	return err == nil
}

// GetGPGInfo returns information about the available GPG installation
func GetGPGInfo() (map[string]string, error) {
	gpg, err := NewGnuPGCipher()
	if err != nil {
		return nil, err
	}
	defer gpg.Cleanup()
	
	info := make(map[string]string)
	
	// Get version
	version, err := gpg.GetVersion()
	if err != nil {
		info["version"] = "Unknown"
	} else {
		info["version"] = version
	}
	
	// Get path
	info["path"] = gpg.gpgPath
	
	// Get available ciphers
	ciphers, err := gpg.ListCiphers()
	if err != nil {
		info["ciphers"] = "Unknown"
	} else {
		info["ciphers"] = strings.Join(ciphers, ", ")
	}
	
	return info, nil
}
