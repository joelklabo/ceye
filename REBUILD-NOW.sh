#!/bin/bash
# Build and optionally install ceye

set -e

cd /Users/honk/code/ceye

echo "🔨 Building ceye..."
go build -o bin/ceye ./cmd/ceye
echo "✅ Build complete!"
echo

echo "🧪 Testing..."
./bin/ceye --version
echo

# Ask if user wants to install
read -p "Install to /usr/local/bin? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "📦 Installing..."
    sudo install -m 755 bin/ceye /usr/local/bin/ceye
    sudo xattr -c /usr/local/bin/ceye 2>/dev/null || true
    echo "✅ Installed to /usr/local/bin/ceye"
    echo
    echo "Testing installed version..."
    ceye --version
    echo
    echo "🎉 All done! Run: ceye --demo --webhooks"
else
    echo "Skipped install."
    echo "To test: ./bin/ceye --demo --webhooks"
fi
