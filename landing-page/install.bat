@echo off
setlocal EnableDelayedExpansion
chcp 65001 >nul 2>&1

:: ==============================================================================
:: GCAM One-Click Installer (Windows)
:: Usage: curl -sSL https://gcam.dong4j.site/install.bat -o install.bat && install.bat
:: Or run directly: install.bat
:: ==============================================================================

set "REPO=dong4j/gemini-cli-account-manager"
set "BINARY_NAME=gcam"

echo.
echo ==========================================
echo   GCAM One-Click Installer (Windows)
echo   https://github.com/%REPO%
echo ==========================================
echo.

:: Detect system architecture
set "OS=windows"
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set "ARCH=amd64"
) else if "%PROCESSOR_ARCHITECTURE%"=="ARM64" (
    set "ARCH=arm64"
) else (
    echo [ERROR] Unsupported architecture: %PROCESSOR_ARCHITECTURE%
    echo Only x64 and ARM64 are supported.
    exit /b 1
)

echo [INFO] Detected platform: %OS%-%ARCH%

:: Get version (from parameter or GitHub API)
set "VERSION=%~1"
if not defined VERSION (
    echo [INFO] Fetching latest version...
    for /f "delims=" %%i in ('powershell -NoProfile -Command "(Invoke-WebRequest -Uri 'https://api.github.com/repos/%REPO%/releases/latest' -UseBasicParsing).Content | ConvertFrom-Json | Select-Object -ExpandProperty tag_name" 2^>nul') do set "VERSION=%%i"
    if not defined VERSION (
        echo [ERROR] Failed to fetch latest version. Please check your network or specify a version.
        echo Usage: install.bat v1.1.0
        exit /b 1
    )
)
echo [INFO] Version: %VERSION%

:: Set install directory
set "INSTALL_DIR=%LOCALAPPDATA%\Programs\gcam"
if not defined LOCALAPPDATA (
    set "INSTALL_DIR=%USERPROFILE%\AppData\Local\Programs\gcam"
)

:: Create install directory
if not exist "%INSTALL_DIR%" (
    echo [INFO] Creating install directory: %INSTALL_DIR%
    mkdir "%INSTALL_DIR%"
)

:: Set download filename
set "FILENAME=%BINARY_NAME%-%OS%-%ARCH%.exe"
set "DOWNLOAD_URL=https://github.com/%REPO%/releases/download/%VERSION%/%FILENAME%"
set "TARGET=%INSTALL_DIR%\%BINARY_NAME%.exe"

echo [INFO] Download URL: %DOWNLOAD_URL%
echo [INFO] Downloading...

:: Download file
powershell -NoProfile -Command "Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%TARGET%' -UseBasicParsing"
if errorlevel 1 (
    echo [ERROR] Download failed. Please verify the version: %VERSION%
    exit /b 1
)

:: Verify file
if not exist "%TARGET%" (
    echo [ERROR] Downloaded file does not exist
    exit /b 1
)

for %%A in ("%TARGET%") do set "SIZE=%%~zA"
if %SIZE% LSS 1000 (
    echo [ERROR] Downloaded file is too small, may have failed
    exit /b 1
)

echo [SUCCESS] Installation complete!
echo.
echo ==========================================
echo   GCAM installed to:
echo   %TARGET%
echo ==========================================
echo.

:: Check PATH
echo [INFO] Checking PATH configuration...
echo %PATH% | findstr /i /C:"%INSTALL_DIR%" >nul
if errorlevel 1 (
    echo [WARN] Need to add %INSTALL_DIR% to system PATH
    echo.
    echo Please add to system environment variables:
    echo   1. Press Win + R, type sysdm.ccl and press Enter
    echo   2. Click "Advanced" ^> "Environment Variables"
    echo   3. In "User variables", select "Path" and click "Edit"
    echo   4. Click "New" and add the following path:
    echo   %INSTALL_DIR%
    echo.
    echo Restart your terminal after adding.
) else (
    echo [INFO] PATH is correctly configured
)

echo.
echo ==========================================
echo   Common Commands
echo ==========================================
echo.
echo   gcam               List accounts
echo   gcam 1             Switch to account #1
echo   gcam next          Switch to next account
echo   gcam quota         Check quota usage
echo   gcam pool login    Add new account
echo   gcam menu          Open interactive menu
echo.
echo   Install hooks (enable /gcam command):
echo   gcam install
echo.
echo   View help:
echo   gcam --help
echo.
echo ==========================================
echo.
echo   Docs: https://gcam.dong4j.site
echo   GitHub: https://github.com/%REPO%
echo.
endlocal
