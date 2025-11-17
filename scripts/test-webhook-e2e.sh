#!/bin/bash
#
# test-webhook-e2e.sh - End-to-end webhook testing script
#
# Tests the complete webhook pipeline:
# 1. Verify GitHub webhook configuration
# 2. Test local webhook endpoint
# 3. Trigger GitHub Action
# 4. Verify webhook receipt
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
OWNER=${GITHUB_OWNER:-"joelklabo"}
REPO=${GITHUB_REPO:-"ceye"}
WEBHOOK_PORT=9090
WEB_PORT=8080
TEST_DURATION=30

# Functions
log_info() {
    echo -e "${BLUE}ℹ${NC}  $1"
}

log_success() {
    echo -e "${GREEN}✅${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠️${NC}  $1"
}

log_error() {
    echo -e "${RED}❌${NC} $1"
}

cleanup() {
    log_info "Cleaning up..."
    if [ ! -z "$CEYE_PID" ]; then
        kill $CEYE_PID 2>/dev/null || true
    fi
    pkill -f "ceye.*--webhook-port" 2>/dev/null || true
}

trap cleanup EXIT

# Main test flow
main() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  🔍 Webhook End-to-End Test"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Testing repository: $OWNER/$REPO"
    echo "Webhook port: $WEBHOOK_PORT"
    echo "Web port: $WEB_PORT"
    echo ""

    # Step 1: Check webhook configuration on GitHub
    log_info "Step 1: Checking GitHub webhook configuration..."
    if command -v gh &> /dev/null; then
        HOOKS=$(gh api "repos/$OWNER/$REPO/hooks" 2>/dev/null || echo "[]")
        
        if [ "$HOOKS" == "[]" ] || [ -z "$HOOKS" ]; then
            log_warning "No webhooks configured on GitHub"
            echo ""
            echo "To configure webhooks, you need:"
            echo "1. A public URL (use ngrok: ngrok http $WEBHOOK_PORT)"
            echo "2. Run: gh api repos/$OWNER/$REPO/hooks -X POST \\"
            echo "     -F url='https://YOUR-NGROK-URL/webhooks/github' \\"
            echo "     -F content_type='application/json' \\"
            echo "     -F events[]='workflow_run' \\"
            echo "     -F active=true"
            echo ""
        else
            log_success "Webhooks configured on GitHub"
            echo "$HOOKS" | jq -r '.[] | "  - \(.config.url) (\(.events | join(", ")))"' 2>/dev/null || echo "$HOOKS"
        fi
    else
        log_warning "gh CLI not installed, skipping GitHub check"
        echo "Install: brew install gh"
    fi
    echo ""

    # Step 2: Test local webhook endpoint
    log_info "Step 2: Starting ceye with webhook server..."
    
    mkdir -p tmp
    ceye --port $WEB_PORT --webhook-port $WEBHOOK_PORT > tmp/ceye-webhook-test.log 2>&1 &
    CEYE_PID=$!
    
    sleep 3
    
    if ! ps -p $CEYE_PID > /dev/null; then
        log_error "ceye failed to start"
        cat tmp/ceye-webhook-test.log
        exit 1
    fi
    
    log_success "ceye started (PID: $CEYE_PID)"
    echo ""

    # Step 3: Test webhook health endpoint
    log_info "Step 3: Testing webhook health endpoint..."
    
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        http://localhost:$WEBHOOK_PORT/webhooks/health 2>/dev/null || echo "000")
    
    if [ "$RESPONSE" == "200" ]; then
        log_success "Webhook endpoint healthy (HTTP $RESPONSE)"
    else
        log_error "Webhook endpoint not responding (HTTP $RESPONSE)"
        log_info "Check tmp/ceye-webhook-test.log for details"
        exit 1
    fi
    echo ""

    # Step 4: Send test webhook
    log_info "Step 4: Sending test webhook payload..."
    
    # Create minimal test payload
    cat > tmp/test-webhook-payload.json << 'EOF'
{
  "action": "completed",
  "workflow_run": {
    "id": 99999999,
    "name": "Test Workflow",
    "head_branch": "main",
    "head_sha": "abc123def456",
    "status": "completed",
    "conclusion": "success",
    "html_url": "https://github.com/test/test/actions/runs/99999999",
    "created_at": "2025-11-17T17:00:00Z",
    "updated_at": "2025-11-17T17:01:00Z"
  },
  "repository": {
    "name": "test-repo",
    "full_name": "test/test-repo"
  }
}
EOF
    
    curl -X POST "http://localhost:$WEBHOOK_PORT/webhooks/github" \
        -H "Content-Type: application/json" \
        -H "X-GitHub-Event: workflow_run" \
        -d @tmp/test-webhook-payload.json \
        > /dev/null 2>&1
    
    sleep 2
    
    if grep -q "workflow_run" tmp/ceye-webhook-test.log; then
        log_success "Test webhook received"
    else
        log_warning "Test webhook may not have been processed"
        log_info "Check tmp/ceye-webhook-test.log"
    fi
    echo ""

    # Step 5: Check for real webhooks (if running long enough)
    log_info "Step 5: Monitoring for real webhooks (${TEST_DURATION}s)..."
    echo "   Trigger a GitHub Action now to test real webhook delivery"
    echo "   Run: gh workflow run CI --ref main"
    echo ""
    
    sleep $TEST_DURATION
    
    if grep -q "Received webhook" tmp/ceye-webhook-test.log || \
       grep -q "workflow_run" tmp/ceye-webhook-test.log; then
        log_success "Webhook activity detected!"
    else
        log_warning "No webhook activity detected in ${TEST_DURATION}s"
        echo ""
        echo "Possible reasons:"
        echo "1. No webhooks configured on GitHub (see Step 1)"
        echo "2. No GitHub Actions triggered during test"
        echo "3. Webhook URL not publicly accessible (need ngrok)"
        echo ""
    fi

    # Step 6: Show summary
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  📊 Test Summary"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    log_info "Recent webhook activity:"
    grep -E "(Received webhook|workflow_run|Webhook)" tmp/ceye-webhook-test.log | tail -5 || echo "   (none)"
    
    echo ""
    log_info "Full logs: tmp/ceye-webhook-test.log"
    echo ""
    
    # Provide next steps
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  🎯 Next Steps"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "To enable real webhook delivery:"
    echo ""
    echo "1. Start ngrok in another terminal:"
    echo "   ngrok http $WEBHOOK_PORT"
    echo ""
    echo "2. Copy the HTTPS URL (e.g., https://abc123.ngrok.io)"
    echo ""
    echo "3. Configure GitHub webhook:"
    echo "   gh api repos/$OWNER/$REPO/hooks -X POST \\"
    echo "     -F url='https://YOUR-NGROK-URL/webhooks/github' \\"
    echo "     -F content_type='application/json' \\"
    echo "     -F events[]='workflow_run' \\"
    echo "     -F active=true"
    echo ""
    echo "4. Trigger a workflow:"
    echo "   gh workflow run CI --ref main"
    echo ""
    echo "5. Watch for webhooks in ceye logs"
    echo ""
}

main "$@"
