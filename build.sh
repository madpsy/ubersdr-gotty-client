#!/bin/bash

# Build script for uberterm (gotty-client with session support)

set -e

echo "Building uberterm..."

# Get version from git or use default
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | awk '{print $3}')

# Build flags
LDFLAGS="-X main.VERSION=${VERSION}"

# Ensure dependencies are available
echo "Downloading dependencies..."
go mod download

# Build the binary
echo "Compiling..."
go build -ldflags "${LDFLAGS}" -o uberterm ./cmd/gotty-client

echo ""
echo "✓ Build successful!"
echo "  Binary: ./uberterm"
echo "  Version: ${VERSION}"
echo "  Go: ${GO_VERSION}"
echo ""
echo "Usage examples:"
echo "  ./uberterm http://localhost:8080"
echo "  ./uberterm sessions http://localhost:8080"
echo "  ./uberterm destroy http://localhost:8080 session-name"
echo "  ./uberterm --admin-password secret http://localhost:8080"
echo ""
