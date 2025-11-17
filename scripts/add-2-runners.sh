#!/bin/bash
# Quick script to add 2 more self-hosted runners (total of 3)

set -e

echo "========================================="
echo "Adding 2 More Self-Hosted Runners"
echo "========================================="
echo

# Check if we already have a runner
EXISTING=$(ps aux | grep "Runner.Listener" | grep -v grep | wc -l)
echo "Current runners: $EXISTING"
echo "Target runners: 3"
echo "Need to add: 2"
echo

# Get repo details
REPO_OWNER="joelklabo"
REPO_NAME="ceye"

echo "📦 Repository: $REPO_OWNER/$REPO_NAME"
echo

# Check gh CLI
if ! command -v gh &> /dev/null; then
    echo "❌ gh CLI not found. Install with: brew install gh"
    exit 1
fi

# Check authentication
if ! gh auth status &> /dev/null; then
    echo "❌ Not authenticated. Run: gh auth login"
    exit 1
fi

echo "✅ gh CLI authenticated"
echo

# Base directory for runners
BASE_DIR="$HOME/actions-runners"
mkdir -p "$BASE_DIR"

# Add runner 2
echo "========================================="
echo "Setting up Runner #2"
echo "========================================="

RUNNER_DIR="$BASE_DIR/runner-2"
mkdir -p "$RUNNER_DIR"
cd "$RUNNER_DIR"

echo "📥 Downloading runner package..."
curl -o actions-runner-osx-arm64-2.321.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.321.0/actions-runner-osx-arm64-2.321.0.tar.gz
tar xzf actions-runner-osx-arm64-2.321.0.tar.gz
rm actions-runner-osx-arm64-2.321.0.tar.gz

echo "🔑 Generating registration token..."
TOKEN=$(gh api -X POST repos/$REPO_OWNER/$REPO_NAME/actions/runners/registration-token --jq .token)

echo "⚙️  Configuring runner..."
./config.sh --url https://github.com/$REPO_OWNER/$REPO_NAME --token "$TOKEN" --name "runner-2" --labels self-hosted,macOS,ARM64 --unattended

echo "🚀 Installing as service..."
./svc.sh install
./svc.sh start

echo "✅ Runner #2 configured and started"
echo

# Add runner 3
echo "========================================="
echo "Setting up Runner #3"
echo "========================================="

RUNNER_DIR="$BASE_DIR/runner-3"
mkdir -p "$RUNNER_DIR"
cd "$RUNNER_DIR"

echo "📥 Downloading runner package..."
curl -o actions-runner-osx-arm64-2.321.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.321.0/actions-runner-osx-arm64-2.321.0.tar.gz
tar xzf actions-runner-osx-arm64-2.321.0.tar.gz
rm actions-runner-osx-arm64-2.321.0.tar.gz

echo "🔑 Generating registration token..."
TOKEN=$(gh api -X POST repos/$REPO_OWNER/$REPO_NAME/actions/runners/registration-token --jq .token)

echo "⚙️  Configuring runner..."
./config.sh --url https://github.com/$REPO_OWNER/$REPO_NAME --token "$TOKEN" --name "runner-3" --labels self-hosted,macOS,ARM64 --unattended

echo "🚀 Installing as service..."
./svc.sh install
./svc.sh start

echo "✅ Runner #3 configured and started"
echo

# Summary
echo "========================================="
echo "✅ Setup Complete!"
echo "========================================="
echo

TOTAL_RUNNERS=$(ps aux | grep "Runner.Listener" | grep -v grep | wc -l)
echo "Total runners now: $TOTAL_RUNNERS"
echo

echo "Runners installed:"
ls -d $BASE_DIR/runner-*
echo

echo "To check status:"
echo "  ps aux | grep Runner.Listener"
echo

echo "To stop all runners:"
echo "  $BASE_DIR/runner-2/svc.sh stop"
echo "  $BASE_DIR/runner-3/svc.sh stop"
echo

echo "To view runner status on GitHub:"
echo "  https://github.com/$REPO_OWNER/$REPO_NAME/settings/actions/runners"
echo

echo "🎉 You now have 3 runners ready to handle parallel jobs!"
