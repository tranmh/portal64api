@echo off
echo Building Somatogramm...

REM Create bin directory if it doesn't exist
if not exist "bin" mkdir bin

REM Build for Windows
echo Building for Windows...
go build -o bin\somatogramm.exe .\cmd\somatogramm\
if %errorlevel% neq 0 (
    echo Build failed for Windows
    pause
    exit /b %errorlevel%
)

REM Build for Linux (cross-compilation)
echo Building for Linux...
set GOOS=linux
set GOARCH=amd64
go build -o bin\somatogramm .\cmd\somatogramm\
if %errorlevel% neq 0 (
    echo Build failed for Linux
    pause
    exit /b %errorlevel%
)

REM Reset environment variables
set GOOS=
set GOARCH=

echo Build completed successfully!
echo Windows executable: bin\somatogramm.exe
echo Linux executable: bin\somatogramm

pause