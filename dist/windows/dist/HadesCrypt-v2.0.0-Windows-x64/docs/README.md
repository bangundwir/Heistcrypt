# HadesCrypt 🔱

**Lock your secrets, rule your data.**

HadesCrypt is a modern, user-friendly file and folder encryption application built with Go and Fyne GUI framework. It provides military-grade encryption with an intuitive drag-and-drop interface.

## Features

### Core Features
- **Drag & Drop Interface**: Simply drag files or folders to encrypt/decrypt them
- **AES-256-GCM Encryption**: Military-grade encryption with authenticated encryption
- **Argon2id Key Derivation**: State-of-the-art password-based key derivation
- **Password Generator**: Built-in secure password generator with strength meter
- **Progress Tracking**: Real-time progress reporting with ETA

### Advanced Features
- **Folder Encryption**: Recursive encryption of entire directories
- **Multiple Encryption Modes**: 
  - Normal Mode (AES-256-GCM)
  - Paranoid Mode (planned: XChaCha20 + Serpent cascade)
- **Configuration System**: Persistent settings stored at `~/.hadescrypt/config.json`
- **Operation History**: Track all encryption/decryption operations
- **Profiles**: Save and load encryption presets
- **Force Decrypt**: Attempt recovery of corrupted files

### Security Features
- **Secure Memory Handling**: Sensitive data is cleared from memory when possible
- **Cryptographically Secure RNG**: Uses `crypto/rand` for all random number generation
- **Unique Nonces**: Each encryption operation uses a unique nonce
- **Integrity Protection**: Built-in authentication prevents tampering

## Installation

### Prerequisites
- Go 1.21 or later
- Git

### Build from Source
```bash
git clone https://github.com/bangundwir/HadesCrypt.git
cd HadesCrypt
go build -o HadesCrypt.exe
```

## Usage

### Basic Usage
1. **Launch HadesCrypt**
2. **Select a file or folder**:
   - Drag and drop files/folders onto the interface, or
   - Click "Select File" button
3. **Enter a password** or click "Generate" for a secure password
4. **Choose operation**:
   - Click "🔒 Encrypt" to encrypt the selected item
   - Click "🔓 Decrypt" to decrypt the selected item

### Advanced Options
Click "Advanced Options ▼" to access additional features:
- **Use Keyfiles**: Add keyfile-based authentication (planned)
- **Paranoid Mode**: Use multiple encryption algorithms (planned)
- **Reed-Solomon ECC**: Add error correction for archival (planned)
- **Force Decrypt**: Attempt to decrypt corrupted files
- **Split into Chunks**: Split large files into smaller pieces (planned)
- **Compress Files**: Compress before encryption (planned)
- **Deniability Mode**: Make encrypted data indistinguishable from random (planned)
- **Recursive Mode**: Enable folder encryption/decryption

## File Formats

### Encrypted Files
- **Extension**: `.hadescrypt`
- **Format**: Custom binary format with header containing metadata
- **Structure**:
  ```
  [4 bytes] Magic: "HAD1"
  [1 byte]  Version
  [16 bytes] Salt for Argon2id
  [8 bytes]  Nonce prefix
  [4 bytes]  Chunk size
  [8 bytes]  Original file size
  [remaining] Encrypted data chunks
  ```

### Encrypted Folders
- Folders are compressed into tar.gz archives before encryption
- The archive is then encrypted using the same format as files
- Decryption automatically detects and extracts archives

## Configuration

Configuration is stored at `~/.hadescrypt/config.json` and includes:
- Window size and theme preferences
- Argon2id parameters (memory, iterations, parallelism)
- Operation history
- Saved profiles
- Last used settings

## Security Considerations

### Cryptographic Algorithms
- **AES-256-GCM**: Provides both confidentiality and authenticity
- **Argon2id**: Memory-hard key derivation function resistant to GPU attacks
- **Default Parameters**: Balanced for desktop security (64 MiB memory, 1 iteration, 4 threads)

### Best Practices
- Use strong, unique passwords
- Keep encrypted files and passwords in separate locations
- Regularly backup encrypted files
- Verify decryption immediately after encryption

### Limitations
- Password recovery is impossible - keep passwords safe
- File metadata (size, timestamps) may leak information
- Large files require sufficient available disk space for temporary files

## Development

### Project Structure
```
HadesCrypt/
├── main.go                 # Main application and GUI
├── internal/
│   ├── archiver/          # Folder archiving functionality
│   ├── config/            # Configuration management
│   ├── cryptoengine/      # Core encryption/decryption
│   ├── password/          # Password generation and strength
│   └── ui/                # UI utilities
├── go.mod                 # Go module definition
└── README.md              # This file
```

### Building
```bash
# Build for current platform
go build -o HadesCrypt.exe

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o HadesCrypt-linux
GOOS=darwin GOARCH=amd64 go build -o HadesCrypt-macos
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues.

## Acknowledgments

- Built with [Fyne](https://fyne.io/) for cross-platform GUI
- Uses [Argon2](https://github.com/P-H-C/phc-winner-argon2) for key derivation
- Inspired by modern encryption tools and security best practices
