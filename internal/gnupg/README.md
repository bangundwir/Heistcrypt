# GnuPG Module for HadesCrypt

This module provides GnuPG/OpenPGP encryption and decryption capabilities for HadesCrypt.

## Features

### 🔐 **OpenPGP Standard Compliance**
- Full compatibility with GnuPG/OpenPGP standard
- Industry-standard encryption algorithms
- Cross-platform compatibility with other PGP tools

### 🛠️ **Supported Algorithms**
- **Ciphers**: AES256, AES192, AES128, TWOFISH, BLOWFISH, 3DES
- **Compression**: ZIP, ZLIB, BZIP2, or none
- **Hash**: SHA-256, SHA-512, SHA-1
- **Key Derivation**: PBKDF2 with configurable iterations

### ⚙️ **Configuration Options**
- **Symmetric Encryption**: Password-based encryption (default)
- **Asymmetric Encryption**: Public key encryption (with recipient key ID)
- **Output Format**: Binary or ASCII armored
- **Compression**: Configurable compression algorithms
- **Trust Model**: Flexible trust model settings

## Installation Requirements

### Windows
- **GnuPG for Windows**: Download from https://www.gnupg.org/download/
- **Common paths**:
  - `C:\Program Files\GnuPG\bin\gpg.exe`
  - `C:\Program Files (x86)\GnuPG\bin\gpg.exe`
  - Via Git Bash: `C:\Git\usr\bin\gpg.exe`
  - Via MSYS2: `C:\msys64\usr\bin\gpg.exe`

### macOS
- **Homebrew**: `brew install gnupg`
- **MacPorts**: `port install gnupg2`
- **Common paths**:
  - `/usr/local/bin/gpg`
  - `/opt/homebrew/bin/gpg`

### Linux
- **Ubuntu/Debian**: `sudo apt install gnupg`
- **RHEL/CentOS**: `sudo yum install gnupg2`
- **Arch Linux**: `sudo pacman -S gnupg`
- **Common paths**:
  - `/usr/bin/gpg`
  - `/usr/local/bin/gpg`

## Usage

### Basic Encryption
```go
import "github.com/bangundwir/HadesCrypt/internal/gnupg"

// Create cipher
gpgCipher, err := gnupg.NewGnuPGCipher()
if err != nil {
    log.Fatal(err)
}
defer gpgCipher.Cleanup()

// Set password
gpgCipher.SetPassphrase("your-password")

// Configure options
options := gnupg.DefaultGnuPGOptions()
options.Cipher = "AES256"
options.Compression = "ZLIB"

// Encrypt file
err = gpgCipher.EncryptFile("input.txt", "output.gpg", options)
```

### Basic Decryption
```go
// Decrypt file
err = gpgCipher.DecryptFile("output.gpg", "decrypted.txt", options)
```

### Stream Operations
```go
// Encrypt stream
err = gpgCipher.EncryptStream(inputReader, outputWriter, options)

// Decrypt stream
err = gpgCipher.DecryptStream(inputReader, outputWriter, options)
```

## Security Features

### 🔒 **Encryption Security**
- **AES-256**: Default cipher for maximum security
- **PBKDF2**: Strong key derivation from passwords
- **Authenticated Encryption**: Built-in integrity protection
- **Secure Random**: Cryptographically secure random number generation

### 🛡️ **Implementation Security**
- **Memory Safety**: Secure cleanup of temporary files
- **Environment Isolation**: Clean environment for GPG execution
- **Error Handling**: Comprehensive error reporting
- **Process Security**: Secure process execution with controlled environment

### 🔐 **Key Management**
- **Symmetric Mode**: Password-based encryption (default)
- **Asymmetric Mode**: Public key encryption support
- **Key Ring Integration**: Uses system GPG keyring when available
- **Trust Models**: Flexible trust model configuration

## Integration with HadesCrypt

### UI Integration
The GnuPG module is integrated into HadesCrypt's main UI:
- **Encryption Mode**: "🔐 GnuPG/OpenPGP (Standard)"
- **Progress Reporting**: Real-time progress updates
- **Error Handling**: User-friendly error messages

### File Format
- **Standard OpenPGP**: Produces standard .gpg files
- **Cross-compatibility**: Compatible with other PGP tools
- **Metadata**: Preserves original file metadata when possible

### Performance
- **Large Files**: Efficient handling of large files
- **Streaming**: Memory-efficient streaming operations
- **Progress Tracking**: Real-time progress reporting
- **Background Operations**: Non-blocking UI operations

## Error Handling

### Common Issues
1. **GPG Not Found**: Install GnuPG or add to PATH
2. **Permission Denied**: Check file permissions
3. **Wrong Passphrase**: Verify password is correct
4. **Corrupted File**: File may be damaged or not a GPG file

### Troubleshooting
```go
// Check if GnuPG is available
if !gnupg.IsAvailable() {
    fmt.Println("GnuPG is not installed or not in PATH")
}

// Get GPG information
info, err := gnupg.GetGPGInfo()
if err == nil {
    fmt.Printf("GPG Version: %s\n", info["version"])
    fmt.Printf("GPG Path: %s\n", info["path"])
    fmt.Printf("Available Ciphers: %s\n", info["ciphers"])
}
```

## Advantages of GnuPG Mode

### 🌍 **Universal Compatibility**
- **Standard Format**: OpenPGP is an open standard (RFC 4880)
- **Cross-platform**: Works across all operating systems
- **Tool Interoperability**: Compatible with GPG, Kleopatra, Thunderbird, etc.
- **Long-term Support**: Mature, well-tested implementation

### 🔧 **Flexibility**
- **Multiple Algorithms**: Wide range of supported ciphers
- **Configurable Security**: Adjustable security parameters
- **Key Management**: Full key management capabilities
- **Trust Networks**: Web of trust support

### 🏢 **Enterprise Ready**
- **Compliance**: Meets various compliance requirements
- **Audit Trail**: Comprehensive logging capabilities
- **Integration**: Easy integration with existing workflows
- **Support**: Commercial support available

## Best Practices

### Security
1. **Use AES-256**: Stick with AES-256 for maximum security
2. **Strong Passwords**: Use complex passwords or key files
3. **Regular Updates**: Keep GnuPG updated to latest version
4. **Secure Storage**: Store encrypted files securely

### Performance
1. **Compression**: Enable compression for large text files
2. **Batch Operations**: Process multiple files efficiently
3. **Memory Management**: Clean up resources promptly
4. **Progress Monitoring**: Use progress callbacks for user feedback

### Compatibility
1. **Standard Options**: Use default options for maximum compatibility
2. **Binary Format**: Use binary format for smaller file sizes
3. **Version Check**: Verify GnuPG version compatibility
4. **Testing**: Test with target systems before deployment

## Future Enhancements

### Planned Features
- **Public Key Support**: Full asymmetric encryption support
- **Key Generation**: Built-in key pair generation
- **Digital Signatures**: File signing and verification
- **Key Management**: Advanced key management features
- **Hardware Tokens**: Smart card and hardware token support

### Integration Improvements
- **Key Import/Export**: GUI for key management
- **Recipient Selection**: UI for selecting recipients
- **Trust Management**: Visual trust network management
- **Batch Operations**: Multi-file encryption support

---

*GnuPG Module - Bringing industry-standard OpenPGP encryption to HadesCrypt* 🔐
