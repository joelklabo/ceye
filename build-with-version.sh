#!/bin/bash
set -e

cd /Users/honk/code/ceye

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")  
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

echo "🔨 Building ceye..."
echo "  Version:    $VERSION"
echo "  Commit:     $COMMIT"
echo "  Build Time: $BUILD_TIME"
echo

go build \
  -ldflags "-X 'main.Version=$VERSION' -X 'main.GitCommit=$COMMIT' -X 'main.BuildTime=$BUILD_TIME'" \
  -o bin/ceye \
  ./cmd/ceye

echo
echo "✅ Built successfully!"
echo
./bin/ceye --version
echo
echo "📍 Binary location: $(pwd)/bin/ceye"
echo
echo "To install system-wide:"
echo "  sudo cp bin/ceye /usr/local/bin/ceye"
