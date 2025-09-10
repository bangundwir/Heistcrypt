@echo off
REM HadesCrypt Release Script for Windows
REM Usage: scripts\create-release.bat <version>
REM Example: scripts\create-release.bat 2.0.1

setlocal enabledelayedexpansion

set VERSION=%1
if "%VERSION%"=="" (
    echo Usage: %0 ^<version^>
    echo Example: %0 2.0.1
    exit /b 1
)

echo 🔱 Creating release for HadesCrypt v%VERSION%

REM Update VERSION file
echo %VERSION% > VERSION
echo ✅ Updated VERSION file to %VERSION%

REM Update CHANGELOG.md if it exists
if exist "CHANGELOG.md" (
    echo # Changelog > temp_changelog.md
    echo. >> temp_changelog.md
    echo ## [v%VERSION%] - %date% >> temp_changelog.md
    echo. >> temp_changelog.md
    echo ### Added >> temp_changelog.md
    echo - >> temp_changelog.md
    echo. >> temp_changelog.md
    echo ### Changed >> temp_changelog.md
    echo - >> temp_changelog.md
    echo. >> temp_changelog.md
    echo ### Fixed >> temp_changelog.md
    echo - >> temp_changelog.md
    echo. >> temp_changelog.md
    echo. >> temp_changelog.md
    type CHANGELOG.md >> temp_changelog.md
    move temp_changelog.md CHANGELOG.md
    echo ✅ Updated CHANGELOG.md
)

REM Create git tag
git add VERSION CHANGELOG.md
git commit -m "Release v%VERSION%"
git tag -a "v%VERSION%" -m "Release v%VERSION%"

echo ✅ Created git tag v%VERSION%

REM Push to remote
echo 🚀 Pushing to remote repository...
git push origin main
git push origin "v%VERSION%"

echo.
echo 🎉 Release v%VERSION% created successfully!
echo.
echo Next steps:
echo 1. GitHub Actions will automatically build and create a release
echo 2. Check the Actions tab in your GitHub repository
echo 3. The release will be available in your repository releases
echo.
echo To create a manual build, you can also run:
echo   go build -ldflags "-s -w -H windowsgui -X main.version=%VERSION%" -o dist\windows\HadesCrypt-v%VERSION%-Windows-x64.exe .
