@echo off
REM Generate Swagger documentation for Portal64 API
echo Generating Swagger documentation...

REM Navigate to project directory if not already there
cd /d "%~dp0"

REM Run swagger init with correct parameters
swag init -g cmd/server/main.go -o docs/generated --parseInternal

if %ERRORLEVEL% EQU 0 (
    echo Swagger documentation generated successfully!
    echo Documentation available at: docs/generated/
    echo.
    echo Files generated:
    echo - docs/generated/docs.go
    echo - docs/generated/swagger.json 
    echo - docs/generated/swagger.yaml
    echo.
    echo Note: You can ignore the warning about "no Go files in root directory"
    echo The generation process works correctly despite this warning.
) else (
    echo Error generating Swagger documentation!
    echo Error code: %ERRORLEVEL%
    pause
)
