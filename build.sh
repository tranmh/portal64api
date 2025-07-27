#!/bin/bash
# Cross-platform build script wrapper
# This script automatically detects the platform and runs the appropriate build tool

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    printf "${1}${2}${NC}\n"
}

# Detect platform
detect_platform() {
    case "$(uname -s)" in
        Linux*)     PLATFORM=Linux;;
        Darwin*)    PLATFORM=Mac;;
        CYGWIN*)    PLATFORM=Cygwin;;
        MINGW*)     PLATFORM=MinGw;;
        MSYS*)      PLATFORM=Msys;;
        *)          PLATFORM="UNKNOWN:$(uname -s)"
    esac
}

# Main execution
detect_platform

if [[ "$PLATFORM" == "Linux" || "$PLATFORM" == "Mac" ]]; then
    # Unix-like systems - use Makefile
    if command -v make >/dev/null 2>&1; then
        if [ -f "Makefile" ]; then
            print_color $GREEN "Detected Unix-like system, using Makefile..."
            make "$@"
        else
            print_color $RED "Makefile not found!"
            exit 1
        fi
    else
        print_color $RED "Make command not found. Please install make."
        exit 1
    fi
elif [[ "$PLATFORM" =~ (Cygwin|MinGw|Msys) ]]; then
    # Windows with Unix-like environment
    print_color $YELLOW "Detected Windows with Unix-like environment..."
    if command -v make >/dev/null 2>&1 && [ -f "Makefile" ]; then
        print_color $GREEN "Using Makefile..."
        make "$@"
    elif command -v powershell.exe >/dev/null 2>&1 && [ -f "build.ps1" ]; then
        print_color $GREEN "Using PowerShell build script..."
        powershell.exe -ExecutionPolicy Bypass -File build.ps1 "$@"
    else
        print_color $RED "Neither make nor PowerShell available!"
        exit 1
    fi
else
    # Unknown platform
    print_color $YELLOW "Unknown platform: $PLATFORM"
    print_color $YELLOW "Please use the appropriate build tool for your system:"
    print_color $GREEN "  Linux/Mac: make <command>"
    print_color $GREEN "  Windows:   .\\build.ps1 <command>"
fi
