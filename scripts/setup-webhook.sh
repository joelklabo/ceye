#!/bin/bash
#
# setup-webhook.sh - Automated GitHub webhook configuration
#
# Usage: ./setup-webhook.sh OWNER REPO
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

OWNER=$1
REPO=$2

if [ -z "$OWNER" ] || [ -z "$REPO" ]; then
    echo -e "${RED}❌${NC} Usage: $0 OWNER REPO"
    echo ""
    echo "Example:"
    echo "  $0 joelklabo ceye"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🪝 GitHub Webhook Setup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Repository: $OWNER/$REPO"
echo ""

# Check gh CLI
if ! command -v gh &> /dev/null; then
    echo -e "${RED}❌${NC} gh CLI not installed"
    echo "   Install: brew install gh"
    exit 1
fi
echo -e "${GREEN}✅${NC} gh CLI installed"

# Check authentication
if ! gh auth status &>/dev/null; then
    echo -e "${RED}❌${NC} Not authenticated with GitHub"
    echo "   Run: gh auth login"
    exit 1
fi
echo -e "${GREEN}✅${NC} Authenticated with GitHub"

# Detect ngrok tunnel
echo ""
echo -e "${BLUE}ℹ${NC}  Detecting ngrok tunnel..."

NGROK_URL=$(curl -s http://localhost:4040/api/tunnels 2>/dev/null | jq -r '.tunnels[] | select(.proto == "https") | .public_url' | head -1)

if [ -z "$NGROK_URL" ]; then
    echo -e "${RED}❌${NC} ngrok not running or no HTTPS tunnel found"
    echo ""
    echo "Start ngrok first:"
    echo "  ngrok http 9090"
    echo ""
    echo "Or start ceye (it will auto-start ngrok):"
    echo "  ceye --webhooks --webhook-port 9090"
    exit 1
fi

echo -e "${GREEN}✅${NC} Found ngrok tunnel: $NGROK_URL"

# Check existing webhooks
echo ""
echo -e "${BLUE}ℹ${NC}  Checking existing webhooks..."

HOOKS=$(gh api "repos/$OWNER/$REPO/hooks" 2>/dev/null || echo "ERROR")

if [ "$HOOKS" == "ERROR" ]; then
    echo -e "${RED}❌${NC} Failed to access repository webhooks"
    echo "   Check repository access permissions"
    exit 1
fi

# Check if webhook with this URL already exists
if echo "$HOOKS" | jq -e ".[] | select(.config.url | contains(\"$NGROK_URL\"))" > /dev/null 2>&1; then
    WEBHOOK_ID=$(echo "$HOOKS" | jq -r ".[] | select(.config.url | contains(\"$NGROK_URL\")) | .id")
    echo -e "${GREEN}✅${NC} Webhook already configured (ID: $WEBHOOK_ID)"
    echo "   URL: $NGROK_URL/webhooks/github"
    echo ""
    echo -e "${BLUE}ℹ${NC}  Test webhook:"
    echo "   gh api repos/$OWNER/$REPO/hooks/$WEBHOOK_ID/pings -X POST"
    exit 0
fi

# Check if OLD webhook exists (different ngrok URL)
OLD_WEBHOOK=$(echo "$HOOKS" | jq -r '.[] | select(.config.url | contains("ngrok")) | .id' | head -1)

if [ ! -z "$OLD_WEBHOOK" ]; then
    echo -e "${YELLOW}⚠️${NC}  Found old ngrok webhook (ID: $OLD_WEBHOOK)"
    echo -n "   Delete old webhook? [y/N] "
    read -r RESPONSE
    if [[ "$RESPONSE" =~ ^[Yy]$ ]]; then
        gh api "repos/$OWNER/$REPO/hooks/$OLD_WEBHOOK" -X DELETE
        echo -e "${GREEN}✅${NC} Deleted old webhook"
    fi
fi

# Create new webhook
echo ""
echo -e "${BLUE}ℹ${NC}  Creating webhook..."

WEBHOOK_URL="$NGROK_URL/webhooks/github"

RESULT=$(gh api "repos/$OWNER/$REPO/hooks" -X POST --input - << EOF
{
  "name": "web",
  "active": true,
  "events": ["workflow_run"],
  "config": {
    "url": "$WEBHOOK_URL",
    "content_type": "json",
    "insecure_ssl": "0"
  }
}
EOF
)

WEBHOOK_ID=$(echo "$RESULT" | jq -r '.id')

if [ -z "$WEBHOOK_ID" ] || [ "$WEBHOOK_ID" == "null" ]; then
    echo -e "${RED}❌${NC} Failed to create webhook"
    echo "$RESULT" | jq .
    exit 1
fi

echo -e "${GREEN}✅${NC} Webhook created successfully!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  📊 Webhook Details"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  ID: $WEBHOOK_ID"
echo "  URL: $WEBHOOK_URL"
echo "  Events: workflow_run"
echo "  Active: true"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🧪 Test Webhook"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "1. Send test ping:"
echo "   gh api repos/$OWNER/$REPO/hooks/$WEBHOOK_ID/pings -X POST"
echo ""
echo "2. Trigger workflow:"
echo "   gh workflow run CI --ref main"
echo ""
echo "3. Or push a commit:"
echo "   git commit --allow-empty -m 'Test webhook'"
echo "   git push"
echo ""
echo "4. Check ceye logs for:"
echo "   'Received GitHub webhook: workflow_run'"
echo ""
