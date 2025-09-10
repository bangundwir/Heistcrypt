# HadesCrypt Windows Distribution

This folder contains Windows-specific build files and distribution assets for HadesCrypt.

## 🏗️ Building for Windows

### Prerequisites
- **Go 1.21+**: Download from [golang.org](https://golang.org/dl/)
- **Git**: For version control and dependency management
- **Optional**: `rsrc` tool for embedding Windows resources

### Quick Build

#### Option 1: Batch Script (Recommended)
```cmd
# From HadesCrypt root directory
dist\windows\build.bat
```

#### Option 2: PowerShell Script
```powershell
# From HadesCrypt root directory
powershell -ExecutionPolicy Bypass -File dist\windows\build.ps1
```

#### Option 3: Manual Build
```cmd
# From HadesCrypt root directory
go mod tidy
go build -ldflags "-s -w -H windowsgui" -o dist\windows\HadesCrypt.exe .
```

### Build Features

#### 🎯 **Optimized Build**
- **Small Binary**: `-s -w` flags strip debug info and symbols
- **GUI Mode**: `-H windowsgui` removes console window
- **Version Info**: Embeds version from `VERSION` file
- **Dependencies**: Automatically resolves and includes all Go modules

#### 📦 **Windows Integration**
- **Manifest**: Windows compatibility and privilege settings
- **Version Info**: Proper Windows file properties
- **Resource Embedding**: Icons and metadata (if `rsrc` available)
- **Digital Signature Ready**: Prepared for code signing

## 📁 Files Description

### Build Scripts
- **`build.bat`**: Windows Batch build script (works everywhere)
- **`build.ps1`**: PowerShell build script (advanced features)

### Windows Resources
- **`manifest.xml`**: Windows application manifest
  - Windows 10/11 compatibility
  - UAC privilege settings
  - High DPI awareness
- **`versioninfo.rc`**: Windows version information
  - File version and product version
  - Company and copyright info
  - Description and comments

### Documentation
- **`README.md`**: This file with build instructions
- **`INSTALL.md`**: Installation instructions for end users

## 🚀 Distribution

### Single Executable
HadesCrypt builds as a single, standalone executable:
- **No Installation Required**: Just run `HadesCrypt.exe`
- **No Dependencies**: All libraries statically linked
- **Portable**: Can run from any location
- **Small Size**: Optimized build (~15-25 MB)

### System Requirements
- **OS**: Windows 10, 11 (x64)
- **Memory**: 100 MB RAM minimum
- **Disk**: 50 MB free space
- **Optional**: GnuPG for OpenPGP encryption mode

### Antivirus Considerations
Some antivirus software may flag the executable due to:
- **Encryption Capabilities**: Legitimate encryption tools sometimes trigger false positives
- **Unsigned Binary**: Without code signing certificate
- **Packed Executable**: Go binaries are compressed

**Solutions**:
1. **Whitelist**: Add to antivirus whitelist
2. **Code Signing**: Sign binary with certificate (for distribution)
3. **VirusTotal**: Check hash on VirusTotal for verification

## 🔧 Advanced Build Options

### Resource Embedding
To embed Windows resources (icons, version info):

```cmd
# Install rsrc tool
go install github.com/akavel/rsrc@latest

# Build with resources (done automatically by build scripts)
rsrc -manifest manifest.xml -ico icon.ico -o rsrc.syso
go build -ldflags "-s -w -H windowsgui" -o HadesCrypt.exe .
```

### Cross-Compilation
Build from Linux/macOS for Windows:

```bash
# Set environment
export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64

# Build
go build -ldflags "-s -w" -o HadesCrypt.exe .
```

### Debug Build
For debugging with console output:

```cmd
go build -o HadesCrypt-debug.exe .
```

## 📋 Build Verification

### File Properties
After successful build, verify:
- **File Size**: ~15-25 MB (depending on features)
- **Version Info**: Right-click → Properties → Details
- **Digital Signature**: (if signed)

### Functionality Test
1. **Launch**: Double-click `HadesCrypt.exe`
2. **GUI**: Should open without console window
3. **Drag & Drop**: Test file selection
4. **Encryption**: Test with small file
5. **Decryption**: Verify round-trip works

### Hash Verification
```powershell
# Generate SHA256 hash
Get-FileHash HadesCrypt.exe -Algorithm SHA256
```

## 🛠️ Troubleshooting

### Build Errors

#### "Go not found"
- Install Go from [golang.org](https://golang.org/dl/)
- Add Go to system PATH
- Restart command prompt

#### "main.go not found"
- Run build script from HadesCrypt root directory
- Ensure you're in correct folder

#### "Build failed"
- Check Go version: `go version`
- Update dependencies: `go mod tidy`
- Clear module cache: `go clean -modcache`

### Runtime Issues

#### "Application failed to start"
- Check Windows version (requires Windows 10+)
- Install Visual C++ Redistributable if needed
- Run as administrator if permission issues

#### "Antivirus blocking"
- Add to antivirus whitelist
- Download from official source only
- Verify file hash against official releases

## 🎯 Distribution Checklist

Before distributing HadesCrypt.exe:

- [ ] Build successful without errors
- [ ] Version info embedded correctly
- [ ] File size reasonable (~15-25 MB)
- [ ] GUI launches without console
- [ ] Basic encryption/decryption works
- [ ] Antivirus scan clean
- [ ] Hash documented for verification
- [ ] README and installation instructions included

## 📞 Support

For build issues or questions:
- **GitHub Issues**: Report build problems
- **Documentation**: Check main README.md
- **Community**: Join discussions for help

---

*HadesCrypt Windows Distribution - Professional encryption for everyone* 🔱
