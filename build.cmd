@echo off
setlocal enabledelayedexpansion


:: 设置变量
set PROJECT_NAME=excelparser
set OUTPUT_DIR=bin
set OUTPUT_EXE=%PROJECT_NAME%.exe
set OUPUT_CLI=%PROJECT_NAME%_cli.exe

:: 创建输出目录
echo [1/8] Creating output directory...
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

:: Check Go environment
echo [2/8] Checking Go environment...
go version
if %errorlevel% neq 0 (
    echo Error: Go environment is not properly configured
    pause
    exit /b 1
)

:: Download dependencies (if needed)
echo [3/8] Downloading dependencies...
go mod download
if %errorlevel% neq 0 (
    echo Warning: There may be issues downloading dependencies, but continuing compilation...
)

:: Check wails3 installation
echo [4/8] Checking wails3 installation...
wails3 version >nul 2>&1
if %errorlevel% neq 0 (
    echo Warning: wails3 is not installed or not in PATH, but continuing compilation...
    echo Tip: You can install wails3 from https://wails.io/ to build GUI applications
    exit /b 1
)

:: build gui application
echo [5/8] Building GUI application with wails3...
wails3 build >nul 2>&1
if %errorlevel% equ 0 (
    echo Build succeeded!
) else (
    echo Error: Build failed, please check the wails3 output for details
    pause
    exit /b 1
)

:: Build console application (if needed)
echo [6/8] Building console application...
go build -o %OUTPUT_DIR%\%OUPUT_CLI% main_cli.go >nul 2>&1
if %errorlevel% equ 0 (
    echo Console build succeeded!
) else (
    echo Error: Console build failed, please check the Go output for details
    pause
    exit /b 1
)

:: Use UPX to compress executable file (optional)
echo [7/8] Using UPX to compress executable file...
upx --best --lzma %OUTPUT_DIR%\%OUPUT_CLI% >nul 2>&1
if %errorlevel% equ 0 (
    echo UPX compression succeeded!
) else (
    echo UPX compression failed or UPX is not installed, skipping compression step
    echo Tip: You can download UPX from https://upx.github.io/ to reduce the executable size
)

:: Check build result
echo [8/8] Checking build result...
if not exist %OUTPUT_DIR%\%OUTPUT_EXE% (
    echo Error: Executable file not generated
    pause
    exit /b 1
)
if not exist %OUTPUT_DIR%\%OUPUT_CLI% (
    echo Warning: Console executable file not generated, but GUI build succeeded
    echo Tip: Check the Go build output for details on why the console build failed
)

:: Display file information
echo.
echo ========================================
echo Build succeeded!
echo ========================================
echo Executable file: %OUTPUT_DIR%\%OUTPUT_EXE%
echo File size:
for %%I in (%OUTPUT_DIR%\%OUTPUT_EXE%) do echo   %%~zI bytes
for %%I in (%OUTPUT_DIR%\%OUPUT_CLI%) do echo   %%~zI bytes
echo.

:: Run the application
echo Starting the application...
echo ========================================
start "" "%OUTPUT_DIR%\%OUTPUT_EXE%"
pause