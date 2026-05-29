@echo off
setlocal EnableDelayedExpansion
chcp 65001 >nul 2>&1

:: ==============================================================================
:: GCAM 一键安装脚本 (Windows 版本)
:: 使用方式: curl -sSL https://gcam.dong4j.site/install.bat -o install.bat && install.bat
:: 或下载后直接运行: install.bat
:: ==============================================================================

set "REPO=dong4j/gemini-cli-account-manager"
set "BINARY_NAME=gcam"

echo.
echo ==========================================
echo   GCAM 一键安装脚本 (Windows)
echo   https://github.com/%REPO%
echo ==========================================
echo.

:: 检测系统架构
set "OS=windows"
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set "ARCH=amd64"
) else if "%PROCESSOR_ARCHITECTURE%"=="ARM64" (
    set "ARCH=arm64"
) else (
    echo [ERROR] 不支持的架构: %PROCESSOR_ARCHITECTURE%
    echo 仅支持 x64 和 ARM64
    exit /b 1
)

echo [INFO] 检测到平台: %OS%-%ARCH%

:: 获取版本号（从参数或 GitHub API）
set "VERSION=%~1"
if not defined VERSION (
    echo [INFO] 获取最新版本号...
    for /f "delims=" %%i in ('powershell -NoProfile -Command "(Invoke-WebRequest -Uri 'https://api.github.com/repos/%REPO%/releases/latest' -UseBasicParsing).Content | ConvertFrom-Json | Select-Object -ExpandProperty tag_name" 2^>nul') do set "VERSION=%%i"
    if not defined VERSION (
        echo [ERROR] 无法获取最新版本，请检查网络连接
        echo 或指定版本: install.bat v1.1.0
        exit /b 1
    )
)
echo [INFO] 版本: %VERSION%

:: 设置安装目录
set "INSTALL_DIR=%LOCALAPPDATA%\Programs\gcam"
if not defined LOCALAPPDATA (
    set "INSTALL_DIR=%USERPROFILE%\AppData\Local\Programs\gcam"
)

:: 创建安装目录
if not exist "%INSTALL_DIR%" (
    echo [INFO] 创建安装目录: %INSTALL_DIR%
    mkdir "%INSTALL_DIR%"
)

:: 设置下载文件名
set "FILENAME=%BINARY_NAME%-%OS%-%ARCH%.exe"
set "DOWNLOAD_URL=https://github.com/%REPO%/releases/download/%VERSION%/%FILENAME%"
set "TARGET=%INSTALL_DIR%\%BINARY_NAME%.exe"

echo [INFO] 下载地址: %DOWNLOAD_URL%
echo [INFO] 正在下载...

:: 下载文件
powershell -NoProfile -Command "Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%TARGET%' -UseBasicParsing"
if errorlevel 1 (
    echo [ERROR] 下载失败，请检查版本号是否正确: %VERSION%
    exit /b 1
)

:: 验证文件
if not exist "%TARGET%" (
    echo [ERROR] 下载文件不存在
    exit /b 1
)

for %%A in ("%TARGET%") do set "SIZE=%%~zA"
if %SIZE% LSS 1000 (
    echo [ERROR] 下载文件过小，可能下载失败
    exit /b 1
)

echo [SUCCESS] 安装完成!
echo.
echo ==========================================
echo   GCAM 已成功安装到:
echo   %TARGET%
echo ==========================================
echo.

:: 检查 PATH
echo [INFO] 检查 PATH 配置...
echo %PATH% | findstr /i /C:"%INSTALL_DIR%" >nul
if errorlevel 1 (
    echo [WARN] 需要将 %INSTALL_DIR% 添加到系统 PATH
    echo.
    echo 请手动添加到系统环境变量:
    echo   1. 按 Win + R，输入 sysdm.cpl 回车
    echo   2. 点击 "高级" ^> "环境变量"
    echo   3. 在 "用户变量" 中选择 "Path"，点击 "编辑"
    echo   4. 点击 "新建"，添加以下路径:
    echo   %INSTALL_DIR%
    echo.
    echo 添加后，重新打开终端使 PATH 生效。
) else (
    echo [INFO] PATH 配置正确
)

echo.
echo ==========================================
echo   常用命令
echo ==========================================
echo.
echo   gcam               查看账号列表
echo   gcam 1             切换到 1 号账号
echo   gcam next          切换到下一个账号
echo   gcam quota         查看配额使用情况
echo   gcam pool login    添加新账号
echo   gcam menu          打开交互式菜单
echo.
echo   安装钩子（启用 /gcam 命令）:
echo   gcam install
echo.
echo   查看帮助:
echo   gcam --help
echo.
echo ==========================================
echo.
echo   文档: https://gcam.dong4j.site
echo   GitHub: https://github.com/%REPO%
echo.
endlocal
