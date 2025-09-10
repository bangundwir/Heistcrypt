# HadesCrypt Windows Build Script (PowerShell)
# This script builds HadesCrypt for Windows with proper resources

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "    HadesCrypt Windows Build Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
try {
    $goVersion = go version
    Write-Host "Found Go: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "Error: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from https://golang.org/dl/" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

# Check if we're in the right directory
if (-not (Test-Path "main.go")) {
    Write-Host "Error: main.go not found" -ForegroundColor Red
    Write-Host "Please run this script from the HadesCrypt root directory" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

# Get version from VERSION file
if (Test-Path "VERSION") {
    $version = Get-Content "VERSION" -Raw
    $version = $version.Trim()
} else {
    $version = "2.0.0"
}

Write-Host "Building HadesCrypt v$version for Windows..." -ForegroundColor Yellow
Write-Host ""

# Clean previous builds
if (Test-Path "HadesCrypt.exe") { Remove-Item "HadesCrypt.exe" -Force }
if (Test-Path "dist\windows\HadesCrypt.exe") { Remove-Item "dist\windows\HadesCrypt.exe" -Force }

# Set build environment
$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "Step 1: Installing dependencies..." -ForegroundColor Blue
try {
    go mod tidy
    Write-Host "Dependencies installed successfully" -ForegroundColor Green
} catch {
    Write-Host "Error: Failed to install dependencies" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "Step 2: Building executable..." -ForegroundColor Blue
try {
    $buildFlags = "-s -w -H windowsgui -X main.version=$version"
    go build -ldflags $buildFlags -o "dist\windows\HadesCrypt.exe" .
    Write-Host "Executable built successfully" -ForegroundColor Green
} catch {
    Write-Host "Error: Build failed" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

# Check if rsrc is available for embedding resources
$rsrcAvailable = $false
try {
    rsrc -h | Out-Null
    $rsrcAvailable = $true
} catch {
    # rsrc not available
}

if ($rsrcAvailable) {
    Write-Host "Step 3: Embedding Windows resources..." -ForegroundColor Blue
    Push-Location "dist\windows"
    try {
        if (Test-Path "icon.ico") {
            rsrc -manifest manifest.xml -ico icon.ico -o rsrc.syso
        } else {
            rsrc -manifest manifest.xml -o rsrc.syso
        }
        
        if (Test-Path "rsrc.syso") {
            Write-Host "Resources embedded successfully" -ForegroundColor Green
            Move-Item "rsrc.syso" "..\..\rsrc.syso" -Force
            Pop-Location
            Write-Host "Rebuilding with resources..." -ForegroundColor Blue
            go build -ldflags $buildFlags -o "dist\windows\HadesCrypt.exe" .
            Remove-Item "rsrc.syso" -Force -ErrorAction SilentlyContinue
        } else {
            Write-Host "Warning: Failed to embed resources" -ForegroundColor Yellow
            Pop-Location
        }
    } catch {
        Write-Host "Warning: Resource embedding failed" -ForegroundColor Yellow
        Pop-Location
    }
} else {
    Write-Host "Step 3: Skipping resource embedding (rsrc not found)" -ForegroundColor Yellow
    Write-Host "To embed resources, install: go install github.com/akavel/rsrc@latest" -ForegroundColor Cyan
}

# Check if build was successful
if (Test-Path "dist\windows\HadesCrypt.exe") {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "         BUILD SUCCESSFUL!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Executable: dist\windows\HadesCrypt.exe" -ForegroundColor Cyan
    Write-Host "Version: $version" -ForegroundColor Cyan
    Write-Host "Architecture: Windows x64" -ForegroundColor Cyan
    Write-Host "GUI Mode: Yes (no console window)" -ForegroundColor Cyan
    Write-Host ""
    
    # Get file size
    $fileSize = (Get-Item "dist\windows\HadesCrypt.exe").Length
    $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
    Write-Host "File size: $fileSize bytes ($fileSizeMB MB)" -ForegroundColor Cyan
    Write-Host ""
    
    # Get file hash for verification
    $hash = Get-FileHash "dist\windows\HadesCrypt.exe" -Algorithm SHA256
    Write-Host "SHA256: $($hash.Hash)" -ForegroundColor Gray
    Write-Host ""
    
    Write-Host "You can now distribute dist\windows\HadesCrypt.exe" -ForegroundColor Green
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "           BUILD FAILED!" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please check the error messages above" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "Build completed successfully!" -ForegroundColor Green
Read-Host "Press Enter to exit"
