@echo off
REM HadesCrypt Windows Installer
REM Simple installer that copies HadesCrypt.exe to a system location

echo ========================================
echo       HadesCrypt Windows Installer
echo ========================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo This installer requires administrator privileges.
    echo Please run as administrator.
    echo.
    pause
    exit /b 1
)

REM Check if HadesCrypt.exe exists
if not exist "HadesCrypt.exe" (
    echo Error: HadesCrypt.exe not found in current directory
    echo Please ensure HadesCrypt.exe is in the same folder as this installer
    echo.
    pause
    exit /b 1
)

REM Set installation directory
set INSTALL_DIR=C:\Program Files\HadesCrypt

echo Installing HadesCrypt to: %INSTALL_DIR%
echo.

REM Create installation directory
if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
    if %errorlevel% neq 0 (
        echo Error: Failed to create installation directory
        pause
        exit /b 1
    )
)

REM Copy executable
copy "HadesCrypt.exe" "%INSTALL_DIR%\HadesCrypt.exe" >nul
if %errorlevel% neq 0 (
    echo Error: Failed to copy HadesCrypt.exe
    pause
    exit /b 1
)

REM Create desktop shortcut
set DESKTOP=%USERPROFILE%\Desktop
if exist "%DESKTOP%" (
    echo Creating desktop shortcut...
    powershell -Command "$WshShell = New-Object -comObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%DESKTOP%\HadesCrypt.lnk'); $Shortcut.TargetPath = '%INSTALL_DIR%\HadesCrypt.exe'; $Shortcut.WorkingDirectory = '%INSTALL_DIR%'; $Shortcut.Description = 'HadesCrypt - Advanced File Encryption'; $Shortcut.Save()"
)

REM Create Start Menu shortcut
set START_MENU=%APPDATA%\Microsoft\Windows\Start Menu\Programs
if exist "%START_MENU%" (
    echo Creating Start Menu shortcut...
    powershell -Command "$WshShell = New-Object -comObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%START_MENU%\HadesCrypt.lnk'); $Shortcut.TargetPath = '%INSTALL_DIR%\HadesCrypt.exe'; $Shortcut.WorkingDirectory = '%INSTALL_DIR%'; $Shortcut.Description = 'HadesCrypt - Advanced File Encryption'; $Shortcut.Save()"
)

REM Add to PATH (optional)
echo.
set /p ADD_TO_PATH="Add HadesCrypt to system PATH? (y/n): "
if /i "%ADD_TO_PATH%"=="y" (
    echo Adding to system PATH...
    setx PATH "%PATH%;%INSTALL_DIR%" /M >nul 2>&1
    if %errorlevel% equ 0 (
        echo Added to PATH successfully
    ) else (
        echo Warning: Failed to add to PATH
    )
)

REM Create uninstaller
echo Creating uninstaller...
(
echo @echo off
echo REM HadesCrypt Uninstaller
echo echo Uninstalling HadesCrypt...
echo.
echo REM Remove desktop shortcut
echo if exist "%DESKTOP%\HadesCrypt.lnk" del "%DESKTOP%\HadesCrypt.lnk"
echo.
echo REM Remove Start Menu shortcut  
echo if exist "%START_MENU%\HadesCrypt.lnk" del "%START_MENU%\HadesCrypt.lnk"
echo.
echo REM Remove installation directory
echo if exist "%INSTALL_DIR%" rmdir /s /q "%INSTALL_DIR%"
echo.
echo echo HadesCrypt has been uninstalled successfully
echo pause
) > "%INSTALL_DIR%\uninstall.bat"

echo.
echo ========================================
echo      INSTALLATION COMPLETED!
echo ========================================
echo.
echo HadesCrypt has been installed successfully!
echo.
echo Installation location: %INSTALL_DIR%
echo Desktop shortcut: Created
echo Start Menu shortcut: Created
echo.
echo You can now:
echo - Double-click the desktop shortcut to launch HadesCrypt
echo - Find HadesCrypt in the Start Menu
echo - Run 'HadesCrypt' from command line (if added to PATH)
echo.
echo To uninstall, run: %INSTALL_DIR%\uninstall.bat
echo.
pause
