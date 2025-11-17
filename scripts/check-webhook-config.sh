#!/bin/bash
#
# check-webhook-config.sh - Quick webhook configuration checker
#

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

OWNER=${GITHUB_OWNER:-"joelklabo"}
REPO=${GITHUB_REPO:-"ceye"}

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🔍 Webhook Configuration Check"
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
echo ""

# Check webhook configuration
echo "Checking webhooks on $OWNER/$REPO..."
HOOKS=$(gh api "repos/$OWNER/$REPO/hooks" 2>/dev/null || echo "ERROR")

if [ "$HOOKS" == "ERROR" ]; then
    echo -e "${RED}❌${NC} Failed to fetch webhooks (check repo access)"
    exit 1
fi

if [ "$HOOKS" == "[]" ] || [ -z "$HOOKS" ]; then
    echo -e "${RED}❌${NC} NO WEBHOOKS CONFIGURED"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  🔧 How to Fix"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Option 1: Using ngrok (Recommended for testing)"
    echo ""
    echo "1. Install ngrok:"
    echo "   brew install ngrok"
    echo ""
    echo "2. Start ngrok tunnel:"
    echo "   ngrok http 9090"
    echo ""
    echo "3. Copy the HTTPS URL (e.g., https://abc123.ngrok.io)"
    echo ""
    echo "4. Configure webhook:"
    echo "   gh api repos/$OWNER/$REPO/hooks -X POST \\"
    echo "     -F url='https://YOUR-NGROK-URL/webhooks/github' \\"
    echo "     -F content_type='application/json' \\"
    echo "     -F events[]='workflow_run' \\"
    echo "     -F active=true"
    echo ""
    echo "Option 2: Using GitHub UI"
    echo ""
    echo "1. Go to: https://github.com/$OWNER/$REPO/settings/hooks/new"
    echo "2. Payload URL: https://YOUR-NGROK-URL/webhooks/github"
    echo "3. Content type: application/json"
    echo "4. Events: Select 'Workflow runs'"
    echo "5. Click 'Add webhook'"
    echo ""
    echo "Option 3: Using a public server"
    echo ""
    echo "1. Deploy ceye to a server with public IP"
    echo "2. Configure webhook URL: https://your-server.com:9090/webhooks/github"
    echo ""
else
    echo -e "${GREEN}✅${NC} Webhooks configured"
    echo ""
    echo "$HOOKS" | jq -r '.[] | "  \(.id): \(.config.url)\n    Events: \(.events | join(", "))\n    Active: \(.active)"' 2>/dev/null
    
    # Check if any webhook is for workflow_run
    HAS_WORKFLOW=$(echo "$HOOKS" | jq -r '.[] | select(.events | contains(["workflow_run"])) | .id' 2>/dev/null)
    
    if [ -z "$HAS_WORKFLOW" ]; then
        echo ""
        echo -e "${YELLOW}⚠️${NC}  No webhook is configured for 'workflow_run' events"
        echo "   Update webhook to include workflow_run events"
    fi
fi

echo ""
