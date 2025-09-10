@echo off
REM HadesCrypt Windows Distribution Package Creator

echo ========================================
echo   HadesCrypt Distribution Packager
echo ========================================
echo.

REM Get version
if exist "..\..\VERSION" (
    set /p VERSION=<..\..\VERSION
) else (
    set VERSION=2.0.0
)

set PACKAGE_NAME=HadesCrypt-v%VERSION%-Windows-x64
set PACKAGE_DIR=dist\%PACKAGE_NAME%

echo Creating distribution package: %PACKAGE_NAME%
echo.

REM Clean previous package
if exist "%PACKAGE_DIR%" rmdir /s /q "%PACKAGE_DIR%"

REM Create package directory
mkdir "%PACKAGE_DIR%"
mkdir "%PACKAGE_DIR%\docs"

REM Check if executable exists
if not exist "HadesCrypt.exe" (
    echo Error: HadesCrypt.exe not found
    echo Please build the executable first using build.bat
    pause
    exit /b 1
)

echo Step 1: Copying executable...
copy "HadesCrypt.exe" "%PACKAGE_DIR%\HadesCrypt.exe" >nul

echo Step 2: Copying documentation...
if exist "..\..\README.md" copy "..\..\README.md" "%PACKAGE_DIR%\docs\README.md" >nul
if exist "..\..\CHANGELOG.md" copy "..\..\CHANGELOG.md" "%PACKAGE_DIR%\docs\CHANGELOG.md" >nul
if exist "..\..\COMMENT_DETECTION.md" copy "..\..\COMMENT_DETECTION.md" "%PACKAGE_DIR%\docs\COMMENT_DETECTION.md" >nul
if exist "..\..\LICENSE" copy "..\..\LICENSE" "%PACKAGE_DIR%\docs\LICENSE" >nul
copy "README.md" "%PACKAGE_DIR%\docs\BUILD_INSTRUCTIONS.md" >nul

echo Step 3: Creating installation files...
copy "install.bat" "%PACKAGE_DIR%\install.bat" >nul

echo Step 4: Creating quick start guide...
(
echo # HadesCrypt v%VERSION% - Quick Start Guide
echo.
echo ## Installation
echo 1. Extract all files to a folder
echo 2. Run install.bat as Administrator ^(optional^)
echo 3. Or simply run HadesCrypt.exe directly
echo.
echo ## Basic Usage
echo 1. Launch HadesCrypt.exe
echo 2. Drag and drop your file into the application
echo 3. Enter a strong password
echo 4. Click "Encrypt" to encrypt or "Decrypt" to decrypt
echo.
echo ## Features
echo - Multiple encryption modes ^(AES-256, ChaCha20, Post-Quantum, GnuPG^)
echo - Password strength meter and generator
echo - Comments and keyfiles support
echo - Drag and drop interface
echo - Progress tracking
echo.
echo ## Support
echo For help and documentation, see the docs folder.
echo.
echo ## Version: %VERSION%
echo Build Date: %DATE%
echo.
echo Lock your secrets, rule your data. 🔱
) > "%PACKAGE_DIR%\QUICK_START.txt"

echo Step 5: Creating file list...
(
echo HadesCrypt v%VERSION% Distribution Contents:
echo.
echo HadesCrypt.exe          - Main application executable
echo install.bat             - Optional system installer
echo QUICK_START.txt         - Quick start guide
echo docs\                   - Documentation folder
echo   README.md             - Main documentation
echo   CHANGELOG.md          - Version history
echo   COMMENT_DETECTION.md  - Comment feature documentation
echo   BUILD_INSTRUCTIONS.md - Build instructions
echo   LICENSE               - License information
echo.
echo File Verification:
) > "%PACKAGE_DIR%\FILES.txt"

REM Add file hashes for verification
echo Calculating file hashes...
powershell -Command "Get-FileHash '%PACKAGE_DIR%\HadesCrypt.exe' -Algorithm SHA256 | ForEach-Object { 'HadesCrypt.exe SHA256: ' + $_.Hash }" >> "%PACKAGE_DIR%\FILES.txt"

echo.
echo ========================================
echo     PACKAGE CREATED SUCCESSFULLY!
echo ========================================
echo.
echo Package: %PACKAGE_DIR%
echo Version: %VERSION%
echo Contents:
dir "%PACKAGE_DIR%" /b
echo.
echo Package size:
powershell -Command "$size = (Get-ChildItem '%PACKAGE_DIR%' -Recurse | Measure-Object -Property Length -Sum).Sum; Write-Host ([math]::Round($size / 1MB, 2)) 'MB'"
echo.
echo To create ZIP archive, use:
echo powershell Compress-Archive -Path "%PACKAGE_DIR%\*" -DestinationPath "%PACKAGE_NAME%.zip"
echo.
pause
