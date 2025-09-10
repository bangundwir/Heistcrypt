# GitHub Workflows Documentation

This document explains the GitHub Actions workflows set up for HadesCrypt.

## 🔄 Workflows Overview

### 1. Build and Release (`build-and-release.yml`)

**Triggers:**
- Push to version tags (e.g., `v2.0.0`)
- Manual workflow dispatch

**Jobs:**
- **Build**: Creates Windows executable with version info
- **Release**: Creates GitHub release with artifacts
- **Test Build**: Manual build testing

**Features:**
- ✅ Automatic version detection from git tags
- ✅ Windows x64 build with GUI mode
- ✅ Portable package creation with documentation
- ✅ GitHub release automation
- ✅ Build artifact upload

### 2. CI/CD (`ci.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main`

**Jobs:**
- **Test**: Runs Go tests and linting
- **Build Check**: Multi-platform build verification
- **Security Scan**: Gosec security analysis

**Features:**
- ✅ Automated testing on every commit
- ✅ Multi-platform build verification (Windows, Linux, macOS)
- ✅ Security vulnerability scanning
- ✅ Code quality checks

## 🚀 How to Create a Release

### Method 1: Using Release Scripts

**Windows:**
```cmd
scripts\create-release.bat 2.0.1
```

**Linux/macOS:**
```bash
chmod +x scripts/create-release.sh
./scripts/create-release.sh 2.0.1
```

### Method 2: Manual Git Tag

```bash
# Update version
echo "2.0.1" > VERSION

# Commit changes
git add VERSION
git commit -m "Release v2.0.1"

# Create and push tag
git tag -a "v2.0.1" -m "Release v2.0.1"
git push origin main
git push origin "v2.0.1"
```

### Method 3: Manual Workflow Dispatch

1. Go to GitHub Actions tab
2. Select "Build and Release HadesCrypt"
3. Click "Run workflow"
4. Enter version number (e.g., `2.0.1`)
5. Click "Run workflow"

## 📦 Release Artifacts

Each release includes:

- **Executable**: `HadesCrypt-v2.0.1-Windows-x64.exe`
- **Portable Package**: Complete folder with:
  - Executable
  - README.md
  - CHANGELOG.md
  - QUICK_START.txt
  - FILES.txt (file listing)

## 🔧 Build Configuration

### Build Flags Used:
```bash
go build -ldflags "-s -w -H windowsgui -X main.version=2.0.1" -o output.exe .
```

- `-s -w`: Strip debug info and symbol table (smaller binary)
- `-H windowsgui`: Hide console window (GUI mode)
- `-X main.version=2.0.1`: Set version at build time

### Environment Variables:
- `GO_VERSION`: Go version to use (currently 1.21)
- `APP_NAME`: Application name (HadesCrypt)
- `BUILD_DIR`: Build output directory (dist)

## 🛡️ Security Features

- **Gosec Scanning**: Automated security vulnerability detection
- **SARIF Upload**: Security results uploaded to GitHub Security tab
- **Dependency Caching**: Go modules cached for faster builds
- **Artifact Retention**: Build artifacts kept for 30 days

## 📋 Workflow Status

Check workflow status at:
- **Actions Tab**: https://github.com/your-repo/actions
- **Releases**: https://github.com/your-repo/releases

## 🔍 Troubleshooting

### Common Issues:

1. **Build Fails**: Check Go version compatibility
2. **Release Not Created**: Verify tag format (must start with 'v')
3. **Artifacts Missing**: Check upload-artifact step logs
4. **Security Scan Fails**: Review Gosec output for issues

### Manual Build Commands:

```bash
# Install dependencies
go mod tidy

# Build for Windows
go build -ldflags "-s -w -H windowsgui" -o HadesCrypt.exe .

# Build with version
go build -ldflags "-s -w -H windowsgui -X main.version=2.0.1" -o HadesCrypt-v2.0.1.exe .
```

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Build Documentation](https://pkg.go.dev/cmd/go#hdr-Build_modes)
- [Fyne GUI Framework](https://fyne.io/)
