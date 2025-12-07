#!/bin/bash
set -e

COMMAND="${1:-build}"

case "$COMMAND" in
    build)
        echo "Building Kader-Planung application..."
        go mod tidy

        # Build Linux binary
        go build -ldflags="-s -w" -o ./bin/kader-planung ./cmd/kader-planung
        echo "Build successful: ./bin/kader-planung"

        # Build Windows binary
        GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/kader-planung.exe ./cmd/kader-planung
        echo "Build successful: ./bin/kader-planung.exe"
        ;;

    clean)
        echo "Cleaning build artifacts..."
        rm -rf ./bin
        rm -f *.log
        rm -f *checkpoint*.json
        echo "Clean complete"
        ;;

    test)
        echo "Running tests..."
        go test -v ./...
        ;;

    run)
        echo "Running application..."
        shift
        ./bin/kader-planung "$@"
        ;;

    deps)
        echo "Installing dependencies..."
        go mod download
        go mod verify
        echo "Dependencies installed"
        ;;

    *)
        echo "Usage: ./build.sh [command]"
        echo "Commands:"
        echo "  build   - Build the application (default)"
        echo "  clean   - Clean build artifacts"
        echo "  test    - Run tests"
        echo "  run     - Run the built application"
        echo "  deps    - Install dependencies"
        exit 1
        ;;
esac
