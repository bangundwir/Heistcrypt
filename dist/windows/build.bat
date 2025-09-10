@echo off
REM HadesCrypt Windows Build Script
REM This script builds HadesCrypt for Windows with proper resources

echo ========================================
echo    HadesCrypt Windows Build Script
echo ========================================
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM Check if we're in the right directory
if not exist "main.go" (
    echo Error: main.go not found
    echo Please run this script from the HadesCrypt root directory
    pause
    exit /b 1
)

REM Get version from VERSION file
if exist "VERSION" (
    set /p VERSION=<VERSION
) else (
    set VERSION=2.0.0
)

echo Building HadesCrypt v%VERSION% for Windows...
echo.

REM Clean previous builds
if exist "HadesCrypt.exe" del "HadesCrypt.exe"
if exist "dist\windows\HadesCrypt.exe" del "dist\windows\HadesCrypt.exe"

REM Set build environment
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

echo Step 1: Installing dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo Error: Failed to install dependencies
    pause
    exit /b 1
)

echo Step 2: Building executable...
go build -ldflags "-s -w -X main.version=%VERSION%" -o "dist\windows\HadesCrypt.exe" .
if %errorlevel% neq 0 (
    echo Error: Build failed
    pause
    exit /b 1
)

REM Check if rsrc is available for embedding resources
where rsrc >nul 2>&1
if %errorlevel% equ 0 (
    echo Step 3: Embedding Windows resources...
    cd dist\windows
    rsrc -manifest manifest.xml -ico icon.ico -o rsrc.syso
    if exist "rsrc.syso" (
        echo Resources embedded successfully
        move rsrc.syso ..\..\
        cd ..\..
        echo Rebuilding with resources...
        go build -ldflags "-s -w -H windowsgui -X main.version=%VERSION%" -o "dist\windows\HadesCrypt.exe" .
        del rsrc.syso
    ) else (
        echo Warning: Failed to embed resources
        cd ..\..
    )
) else (
    echo Step 3: Skipping resource embedding (rsrc not found)
    echo To embed resources, install: go install github.com/akavel/rsrc@latest
)

REM Check if build was successful
if exist "dist\windows\HadesCrypt.exe" (
    echo.
    echo ========================================
    echo         BUILD SUCCESSFUL!
    echo ========================================
    echo.
    echo Executable: dist\windows\HadesCrypt.exe
    echo Version: %VERSION%
    echo Architecture: Windows x64
    echo GUI Mode: Yes (no console window)
    echo.
    
    REM Get file size
    for %%I in ("dist\windows\HadesCrypt.exe") do set SIZE=%%~zI
    echo File size: %SIZE% bytes
    echo.
    
    echo You can now distribute dist\windows\HadesCrypt.exe
    echo.
) else (
    echo.
    echo ========================================
    echo           BUILD FAILED!
    echo ========================================
    echo.
    echo Please check the error messages above
    pause
    exit /b 1
)

echo Build completed successfully!
pause
