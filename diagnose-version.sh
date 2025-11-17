#!/bin/bash

echo "🔍 ceye Version Diagnostic"
echo "=========================="
echo

echo "1. Checking installed versions:"
echo "-------------------------------"
if [ -f "/usr/local/bin/ceye" ]; then
    echo "✅ Found: /usr/local/bin/ceye"
    /usr/local/bin/ceye --version 2>&1 || echo "  ⚠️  Failed to run"
    echo
else
    echo "❌ Not found: /usr/local/bin/ceye"
    echo
fi

if [ -f "/Users/honk/code/ceye/bin/ceye" ]; then
    echo "✅ Found: /Users/honk/code/ceye/bin/ceye"
    /Users/honk/code/ceye/bin/ceye --version 2>&1 || echo "  ⚠️  Failed to run"
    echo
else
    echo "❌ Not found: /Users/honk/code/ceye/bin/ceye"
    echo
fi

echo "2. Which ceye is in PATH:"
echo "-------------------------"
which ceye 2>&1 || echo "❌ ceye not in PATH"
echo

echo "3. Testing basic execution:"
echo "---------------------------"
echo "Testing system ceye (if exists):"
if command -v ceye >/dev/null 2>&1; then
    timeout 2 ceye --demo --demo-duration 1s 2>&1 | head -5 || echo "Failed or timed out"
else
    echo "ceye command not found"
fi
echo

echo "Testing local ceye:"
if [ -f "/Users/honk/code/ceye/bin/ceye" ]; then
    timeout 2 /Users/honk/code/ceye/bin/ceye --demo --demo-duration 1s 2>&1 | head -5 || echo "Failed or timed out"
else
    echo "Local binary not found"
fi
echo

echo "4. Recommendation:"
echo "------------------"
echo "Run one of these commands to rebuild and test:"
echo
echo "  cd /Users/honk/code/ceye"
echo "  go build -o bin/ceye ./cmd/ceye"
echo "  ./bin/ceye --version"
echo "  ./bin/ceye --demo --demo-duration 3s"
echo
echo "If that works, install it:"
echo "  sudo cp bin/ceye /usr/local/bin/ceye"
echo

echo "✅ Diagnostic complete!"
