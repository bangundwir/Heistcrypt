package cryptoengine

import (
	"fmt"
	"os"
	
	"github.com/bangundwir/HadesCrypt/internal/gnupg"
)

// EncryptFileWithGnuPG encrypts a file using GnuPG
func EncryptFileWithGnuPG(inputPath, outputPath string, password []byte, onProgress ProgressCallback) error {
    // Initialize GnuPG cipher
    gpgCipher, err := gnupg.NewGnuPGCipher()
    if err != nil {
        return fmt.Errorf("failed to initialize GnuPG: %w", err)
    }
    defer gpgCipher.Cleanup()
    
    // Set passphrase
    gpgCipher.SetPassphrase(string(password))
    
    // Get file size for progress reporting
    fileInfo, err := os.Stat(inputPath)
    if err != nil {
        return fmt.Errorf("failed to get file info: %w", err)
    }
    totalSize := fileInfo.Size()
    
    // Configure GnuPG options
    options := gnupg.DefaultGnuPGOptions()
    options.Cipher = "AES256"
    options.Compression = "ZLIB"
    options.UseSymmetric = true
    options.ArmorOutput = false // Binary output
    
    // Report initial progress
    if onProgress != nil {
        onProgress(0, totalSize)
    }
    
    // Encrypt file
    err = gpgCipher.EncryptFile(inputPath, outputPath, options)
    if err != nil {
        return fmt.Errorf("GnuPG encryption failed: %w", err)
    }
    
    // Report completion
    if onProgress != nil {
        onProgress(totalSize, totalSize)
    }
    
    return nil
}

// DecryptFileWithGnuPG decrypts a file using GnuPG
func DecryptFileWithGnuPG(inputPath, outputPath string, password []byte, onProgress ProgressCallback) error {
    // Initialize GnuPG cipher
    gpgCipher, err := gnupg.NewGnuPGCipher()
    if err != nil {
        return fmt.Errorf("failed to initialize GnuPG: %w", err)
    }
    defer gpgCipher.Cleanup()
    
    // Set passphrase
    gpgCipher.SetPassphrase(string(password))
    
    // Get file size for progress reporting
    fileInfo, err := os.Stat(inputPath)
    if err != nil {
        return fmt.Errorf("failed to get file info: %w", err)
    }
    totalSize := fileInfo.Size()
    
    // Configure GnuPG options
    options := gnupg.DefaultGnuPGOptions()
    options.UseSymmetric = true
    
    // Report initial progress
    if onProgress != nil {
        onProgress(0, totalSize)
    }
    
    // Decrypt file
    err = gpgCipher.DecryptFile(inputPath, outputPath, options)
    if err != nil {
        return fmt.Errorf("GnuPG decryption failed: %w", err)
    }
    
    // Report completion
    if onProgress != nil {
        onProgress(totalSize, totalSize)
    }
    
    return nil
}
