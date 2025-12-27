@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo   Windows GUI应用 无窗口编译和运行脚本
echo ========================================
echo.

:: 设置变量
set PROJECT_NAME=excelparser
set OUTPUT_DIR=.
set OUTPUT_EXE=%PROJECT_NAME%.exe

:: 清理之前的构建
echo [1/7] Cleaning previous build files...
if exist %OUTPUT_DIR%\%OUTPUT_EXE% del /q %OUTPUT_DIR%\%OUTPUT_EXE%
if exist %OUTPUT_EXE% del /q %OUTPUT_EXE%
if exist *.syso del /f /q *.syso

:: 创建输出目录
echo [2/7] Creating output directory...
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

:: 检查windres工具并创建资源文件
echo [3/7] Checking windres tool and creating resource file...
where windres >nul 2>&1
if %errorlevel% equ 0 (
    echo Creating resource file...
    windres.exe -i app.rc -o defaultRes_windows_amd64.syso -F pe-x86-64
    if %errorlevel% neq 0 (
        echo Warning: Resource file creation failed, continuing compilation without custom icons
    )
) else (
    echo Warning: windres not found, skipping custom resource creation
    echo Tip: Install MinGW-w64 to enable custom icons and version information
)

:: Check Go environment
echo [4/7] Checking Go environment...
go version
if %errorlevel% neq 0 (
    echo Error: Go environment is not properly configured
    pause
    exit /b 1
)

:: Download dependencies (if needed)
echo [5/7] Downloading dependencies...
go mod download
if %errorlevel% neq 0 (
    echo Warning: There may be issues downloading dependencies, but continuing compilation...
)

:: Compile as Windows GUI application (no console window)
echo [6/7] Compiling as Windows GUI application (no console window)...
go build -tags gui,tempdll -ldflags "-w -s -H=windowsgui" -o %OUTPUT_DIR%\%OUTPUT_EXE%
if %errorlevel% neq 0 (
    echo Compilation failed!
    pause
    exit /b 1
)

:: 使用UPX压缩可执行文件
echo [6.5/7] Using UPX to compress executable file...
upx --best --lzma %OUTPUT_DIR%\%OUTPUT_EXE% >nul 2>&1
if %errorlevel% equ 0 (
    echo UPX compression succeeded!
) else (
    echo UPX compression failed or UPX is not installed, skipping compression step
    echo Tip: You can download UPX from https://upx.github.io/ to reduce the executable size
)

:: Check build result
echo [7/7] Checking build result...
if not exist %OUTPUT_DIR%\%OUTPUT_EXE% (
    echo Error: Executable file not generated
    pause
    exit /b 1
)

:: 显示文件信息
echo.
echo ========================================
echo 编译成功!
echo ========================================
echo 可执行文件: %OUTPUT_DIR%\%OUTPUT_EXE%
echo 文件大小:
for %%I in (%OUTPUT_DIR%\%OUTPUT_EXE%) do echo   %%~zI 字节
echo.

:: 运行程序
echo 正在启动应用程序...
echo ========================================
start "" "%OUTPUT_DIR%\%OUTPUT_EXE%"
pause