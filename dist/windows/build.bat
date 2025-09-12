@echo off
setlocal ENABLEDELAYEDEXPANSION
REM ------------------------------------------------------------
REM  HadesCrypt Windows Build Script (Enhanced)
REM  Usage:
REM    build.bat [version] [archs] [zip]
REM      version : optional override (e.g. 2.0.1). Falls back to VERSION file.
REM      archs   : list separated by commas (default: amd64). Supported: amd64,arm64,386
REM      zip     : if set to "zip" creates .zip + .sha256 files
REM  Examples:
REM    build.bat
REM    build.bat 2.0.1
REM    build.bat 2.0.1 amd64,arm64 zip
REM ------------------------------------------------------------

echo ==========================================================
echo                 HadesCrypt Windows Builder
echo ==========================================================

REM Determine script directory (this file lives in dist\windows)
set SCRIPT_DIR=%~dp0
for %%I in ("%SCRIPT_DIR%..\..") do set ROOT_DIR=%%~fI
cd /d "%ROOT_DIR%"

REM Validate Go
go version >nul 2>&1 || (
    echo [ERROR] Go not found in PATH. Install from https://go.dev/dl
    goto :fail
)

REM Resolve version
set INPUT_VERSION=%~1
if NOT "%INPUT_VERSION%"=="" set VERSION=%INPUT_VERSION%
if "%VERSION%"=="" if exist VERSION set /p VERSION=<VERSION
if "%VERSION%"=="" set VERSION=dev

REM Resolve architectures
set ARCHS=%~2
if "%ARCHS%"=="" set ARCHS=amd64

REM Normalize comma list to space
set ARCHS=%ARCHS:,= %

REM Check for zip flag
set MAKE_ZIP=%~3

echo Version    : %VERSION%
echo Architectures: %ARCHS%
if /I "%MAKE_ZIP%"=="zip" echo Packaging : ZIP + SHA256
echo Root Dir   : %ROOT_DIR%
echo.

REM Prepare output dir
if not exist dist\windows mkdir dist\windows

echo [1/6] Tidying modules...
go mod tidy || goto :fail

REM Optional resource embedding (manifest/icon) only once per run
set RSRCSUPPORTED=0
where rsrc >nul 2>&1 && set RSRCSUPPORTED=1
if %RSRCSUPPORTED%==1 (
    if exist dist\windows\manifest.xml if exist dist\windows\icon.ico (
        echo [INFO] Resource files detected (manifest.xml, icon.ico)
    ) else (
        echo [WARN] rsrc found but manifest.xml/icon.ico missing in dist\windows. Skipping embedding.
        set RSRCSUPPORTED=0
    )
) else (
    echo [INFO] rsrc not installed (go install github.com/akavel/rsrc@latest) - skipping resources.
)

for %%A in (%ARCHS%) do call :build_arch %%A || goto :fail

echo.
echo ==========================================================
echo BUILD COMPLETE for version %VERSION%
echo Output dir: dist\windows
echo ==========================================================
goto :eof

REM ------------------------------------------------------------
:build_arch
set ARCH=%1
echo.
echo [2/6] Building arch: %ARCH%
set GOOS=windows
set GOARCH=%ARCH%
set CGO_ENABLED=1

set BIN_NAME=HadesCrypt-%VERSION%-windows-%ARCH%.exe
set BIN_PATH=dist\windows\%BIN_NAME%
if exist "%BIN_PATH%" del "%BIN_PATH%"

if %RSRCSUPPORTED%==1 (
    pushd dist\windows
    rsrc -manifest manifest.xml -ico icon.ico -o rsrc.syso >nul 2>&1
    if exist rsrc.syso (
        echo [3/6] Embedded resources for %ARCH%
        move rsrc.syso ..\.. >nul
        popd
        go build -ldflags "-s -w -H windowsgui -X main.version=%VERSION%" -o "%BIN_PATH%" . || goto :build_fail
        del rsrc.syso >nul 2>&1
    ) else (
        popd
        echo [WARN] Resource embedding failed; continuing without.
        go build -ldflags "-s -w -H windowsgui -X main.version=%VERSION%" -o "%BIN_PATH%" . || goto :build_fail
    )
) else (
    go build -ldflags "-s -w -H windowsgui -X main.version=%VERSION%" -o "%BIN_PATH%" . || goto :build_fail
)

for %%I in ("%BIN_PATH%") do set BINSIZE=%%~zI
echo [4/6] Built %BIN_NAME% (%BINSIZE% bytes)

REM Compute SHA256 checksum (certutil fallback)
echo [5/6] Generating SHA256 checksum...
certutil -hashfile "%BIN_PATH%" SHA256 > "%BIN_PATH%.sha256.tmp" 2>nul
if exist "%BIN_PATH%.sha256.tmp" (
    (for /f "usebackq tokens=*" %%L in ("%BIN_PATH%.sha256.tmp") do @echo %%L) > "%BIN_PATH%.sha256.full"
    for /f "tokens=1" %%H in ('findstr /R /I "^[0-9A-F][0-9A-F]*$" "%BIN_PATH%.sha256.full"') do echo %%H *%BIN_NAME%>"%BIN_PATH%.sha256"
    del "%BIN_PATH%.sha256.tmp" "%BIN_PATH%.sha256.full" >nul 2>&1
    if exist "%BIN_PATH%.sha256" echo [INFO] SHA256: created %BIN_NAME%.sha256
) else (
    echo [WARN] Could not generate SHA256 (certutil unavailable)
)

if /I "%MAKE_ZIP%"=="zip" (
    echo [6/6] Creating ZIP package...
    set ZIP_NAME=%BIN_NAME:.exe=.zip%
    powershell -NoLogo -NoProfile -Command "Compress-Archive -Path '%BIN_PATH%' -DestinationPath 'dist\\windows\\%ZIP_NAME%' -Force" >nul 2>&1 && echo [INFO] Created %ZIP_NAME% || echo [WARN] ZIP creation failed (PowerShell Compress-Archive missing)
)

echo [DONE] %ARCH% build finished.
exit /b 0

:build_fail
echo [ERROR] Build failed for arch %ARCH%
exit /b 1

REM ------------------------------------------------------------
:fail
echo.
echo ================= BUILD FAILED =================
echo Check messages above.
echo =================================================
exit /b 1
