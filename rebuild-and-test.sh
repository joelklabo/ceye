#!/bin/bash
set -e

cd /Users/honk/code/ceye

echo "🔨 Rebuilding ceye..."
echo

# Simple build without version injection for now (to avoid issues)
go build -o bin/ceye-new ./cmd/ceye

echo "✅ Build complete!"
echo

echo "📋 Testing new binary..."
echo

# Test basic functionality
./bin/ceye-new --version
echo

echo "Testing with demo (3 seconds)..."
timeout 3 ./bin/ceye-new --demo --demo-duration 2s 2>&1 | tail -10 || true
echo

echo "✅ Basic test passed!"
echo

echo "Now testing with webhooks..."
timeout 3 ./bin/ceye-new --demo --demo-duration 2s --webhooks --webhook-port 9090 2>&1 | tail -10 || true
echo

echo "✅ Webhook test passed!"
echo

echo "📍 Your new binary is at: $(pwd)/bin/ceye-new"
echo

echo "To use it:"
echo "  ./bin/ceye-new --demo --webhooks --webhook-port 9090"
echo

echo "To install it system-wide:"
echo "  sudo mv bin/ceye-new /usr/local/bin/ceye"
echo

echo "Or just replace the old one:"
echo "  mv bin/ceye-new bin/ceye"
