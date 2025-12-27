@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo   Windows GUI Application Build and UPX Compression Script
echo ========================================
echo.

:: Set variables
set PROJECT_NAME=excelparser
set OUTPUT_DIR=.
set OUTPUT_EXE=%PROJECT_NAME%.exe

:: Check Go environment
echo [1/3] Checking Go environment...
go version
if %errorlevel% neq 0 (
    echo Error: Go environment not properly configured
    pause
    exit /b 1
)

:: Download dependencies (if needed)
echo [2/3] Downloading dependency packages...
go mod download
if %errorlevel% neq 0 (
    echo Warning: Dependency download may have issues, but continuing compilation...
)

:: Compile to Windows GUI application (no console window)
echo [3/3] Compiling to Windows GUI application (no console window)...
go build -tags gui,tempdll -ldflags "-w -s -H=windowsgui" -o %OUTPUT_DIR%\%OUTPUT_EXE%
if %errorlevel% neq 0 (
    echo Compilation failed!
    pause
    exit /b 1
)

:: Check compilation result
echo Checking compilation result...
if not exist %OUTPUT_DIR%\%OUTPUT_EXE% (
    echo Error: Executable file not generated
    pause
    exit /b 1
)

:: Display file information
echo.
echo ========================================
echo Compilation successful!
echo ========================================
echo Executable file: %OUTPUT_DIR%\%OUTPUT_EXE%
echo File size:
for %%I in (%OUTPUT_DIR%\%OUTPUT_EXE%) do echo   %%~zI bytes
echo.

:: Use UPX to compress executable file
echo ========================================
echo Using UPX to compress executable file...
upx --best --lzma %OUTPUT_DIR%\%OUTPUT_EXE% >nul 2>&1
if %errorlevel% equ 0 (
    echo UPX compression successful!
    
    :: Display compressed file size
    echo Compressed file size:
    for %%I in (%OUTPUT_DIR%\%OUTPUT_EXE%) do echo   %%~zI bytes
    
    echo.
    echo ========================================
    echo Compression complete!
    echo ========================================
) else (
    echo UPX compression failed or UPX not installed, skipping compression step
    echo Tip: You can download UPX from https://upx.github.io/ to reduce executable file size
)

echo.
echo Program is ready!
echo Executable file located at: %OUTPUT_DIR%\%OUTPUT_EXE%
echo.
pause