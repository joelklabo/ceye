#!/bin/bash
set -e

echo "=== Starting ceye with Webhooks ==="
echo

# Install ngrok if needed
if ! command -v ngrok &> /dev/null; then
    echo "Installing ngrok..."
    brew install ngrok
fi

# Build latest
echo "Building ceye..."
go build -o bin/webhook-test ./cmd/webhook-test

# Start webhook server
echo
echo "Starting webhook server on port 9090..."
./bin/webhook-test > /tmp/ceye-webhooks.log 2>&1 &
WEBHOOK_PID=$!
echo "✓ Webhook server started (PID: $WEBHOOK_PID)"
sleep 2

# Start ngrok
echo "Starting ngrok tunnel..."
ngrok http 9090 --log=stdout > /tmp/ceye-ngrok.log 2>&1 &
NGROK_PID=$!
sleep 3

# Get URL
NGROK_URL=$(curl -s http://localhost:4040/api/tunnels 2>/dev/null | python3 -c "import sys, json; print(json.load(sys.stdin)['tunnels'][0]['public_url'])" 2>/dev/null || echo "")

if [ -z "$NGROK_URL" ]; then
    echo "❌ Failed to get ngrok URL. Check /tmp/ceye-ngrok.log"
    kill $WEBHOOK_PID $NGROK_PID 2>/dev/null
    exit 1
fi

echo "✓ ngrok URL: $NGROK_URL"
WEBHOOK_URL="$NGROK_URL/webhooks/github"

echo
echo "===================================="
echo "✓ Setup complete!"
echo "===================================="
echo
echo "Webhook endpoint: $WEBHOOK_URL"
echo "ngrok dashboard:  http://localhost:4040"
echo "Server logs:      tail -f /tmp/ceye-webhooks.log"
echo
echo "To set up GitHub webhook, run:"
echo "  gh api repos/joelklabo/ceye/hooks --method POST \\"
echo "    --field name=web --field active=true \\"
echo "    --field events[]=workflow_run \\"
echo "    --field 'config={\"url\":\"$WEBHOOK_URL\",\"content_type\":\"json\"}'"
echo
echo "Or visit: https://github.com/joelklabo/ceye/settings/hooks/new"
echo "  Payload URL: $WEBHOOK_URL"
echo "  Content type: application/json"
echo "  Events: Workflow runs"
echo
echo "Press Ctrl+C to stop all services"

trap "echo ''; echo 'Stopping services...'; kill $WEBHOOK_PID $NGROK_PID 2>/dev/null; exit 0" INT
tail -f /tmp/ceye-webhooks.log
