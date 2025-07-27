@echo off
REM Portal64 API Build Script Wrapper for Windows
REM This batch file provides easy access to the PowerShell build script

setlocal enabledelayedexpansion

REM Check if PowerShell is available
powershell -Command "Get-Host" >nul 2>&1
if !errorlevel! neq 0 (
    echo ERROR: PowerShell is not available or not in PATH
    echo Please install PowerShell or add it to your PATH
    exit /b 1
)

REM Check if we have any arguments
if "%~1"=="" (
    set COMMAND=help
) else (
    set COMMAND=%~1
)

REM Run the PowerShell script with the command
powershell -ExecutionPolicy Bypass -File "build.ps1" "%COMMAND%"

REM Preserve exit code
exit /b !errorlevel!
