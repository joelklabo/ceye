# Webhook Setup Guide

Complete guide for setting up GitHub webhooks with ceye for real-time CI/CD monitoring.

## Prerequisites

- **ngrok** installed: `brew install ngrok`
- **gh CLI** installed: `brew install gh`
- **GitHub authentication**: `gh auth login`

## Quick Start (Automatic)

ceye automatically detects and starts ngrok when running in webhook mode:

```bash
ceye --webhooks --webhook-port 9090
```

Output will show:
```
🌐 ngrok tunnel: https://abc123.ngrok-free.dev
   Webhook endpoint: https://abc123.ngrok-free.dev/webhooks/github
   
   Configure webhook with:
   ./scripts/setup-webhook.sh OWNER REPO
```

Then use the helper script:
```bash
./scripts/setup-webhook.sh joelklabo ceye
```

## Manual Setup

### Step 1: Start ngrok

```bash
ngrok http 9090
```

Copy the HTTPS URL from the output (e.g., `https://abc123.ngrok-free.dev`)

### Step 2: Configure GitHub Webhook

#### Option A: Using gh CLI (Recommended)

```bash
# Set variables
OWNER="your-github-username"
REPO="your-repo-name"
NGROK_URL="https://abc123.ngrok-free.dev"  # From Step 1

# Create webhook
gh api repos/$OWNER/$REPO/hooks -X POST --input - << EOF
{
  "name": "web",
  "active": true,
  "events": ["workflow_run"],
  "config": {
    "url": "$NGROK_URL/webhooks/github",
    "content_type": "json"
  }
}
EOF
```

#### Option B: Using GitHub Web UI

1. Go to: `https://github.com/OWNER/REPO/settings/hooks/new`
2. **Payload URL**: `https://YOUR-NGROK-URL/webhooks/github`
3. **Content type**: `application/json`
4. **Events**: Select "Workflow runs"
5. **Active**: Checked
6. Click **Add webhook**

### Step 3: Verify Webhook

Send a test ping:
```bash
# Get the webhook ID from the previous command or:
WEBHOOK_ID=$(gh api repos/$OWNER/$REPO/hooks | jq -r '.[0].id')

# Send test ping
gh api repos/$OWNER/$REPO/hooks/$WEBHOOK_ID/pings -X POST
```

Check ceye logs for:
```
Received GitHub webhook: ping (delivery: xxx)
```

### Step 4: Test with Real Workflow

Trigger a GitHub Action:
```bash
# Option 1: Trigger via gh CLI
gh workflow run CI --ref main

# Option 2: Push empty commit
git commit --allow-empty -m "Test webhook"
git push
```

Within seconds, you should see in ceye logs:
```
Received GitHub webhook: workflow_run (delivery: xxx)
✅ Parsed GitHub webhook: owner/repo/.github/workflows/ci.yml
Sent RunEvent to channel
STORE: merge called - 1 runs from provider 'github'
```

## Helper Scripts

### Check Webhook Configuration

Verify webhooks are properly configured:

```bash
./scripts/check-webhook-config.sh
```

Output:
```
✅ gh CLI installed
✅ Authenticated with GitHub
✅ Webhooks configured

  581485630: https://abc123.ngrok-free.dev/webhooks/github
    Events: workflow_run
    Active: true
```

### Setup Webhook (Automated)

Automatically configure webhook with current ngrok URL:

```bash
./scripts/setup-webhook.sh OWNER REPO
```

This script:
1. Detects ngrok tunnel URL
2. Checks for existing webhooks
3. Creates webhook if needed
4. Verifies configuration

## Troubleshooting

### ngrok not found

```bash
brew install ngrok
```

### ngrok tunnel not accessible

Check ngrok is running:
```bash
curl http://localhost:4040/api/tunnels
```

Should return JSON with tunnel info.

### Webhook not receiving events

1. **Check webhook configuration**:
   ```bash
   gh api repos/OWNER/REPO/hooks
   ```

2. **Verify webhook is active**:
   Look for `"active": true` in the output

3. **Check webhook URL**:
   Should match your current ngrok URL

4. **Test webhook endpoint**:
   ```bash
   curl https://YOUR-NGROK-URL/webhooks/health
   ```
   Should return `200 OK`

5. **Check ceye logs**:
   Look for errors or webhook receipts

### Webhooks stopped working

**Common cause**: ngrok URL changed

ngrok free tier assigns new URLs on restart. You need to:

1. Get new ngrok URL:
   ```bash
   curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url'
   ```

2. Update webhook:
   ```bash
   # Get webhook ID
   WEBHOOK_ID=$(gh api repos/OWNER/REPO/hooks | jq -r '.[0].id')
   
   # Update URL
   gh api repos/OWNER/REPO/hooks/$WEBHOOK_ID -X PATCH \
     -F config[url]="https://NEW-NGROK-URL/webhooks/github"
   ```

Or delete and recreate:
```bash
# Delete old webhook
gh api repos/OWNER/REPO/hooks/$WEBHOOK_ID -X DELETE

# Create new one (see Step 2 above)
```

## Webhook Events

ceye listens for these GitHub webhook events:

- `workflow_run` - Workflow run state changes (queued, in_progress, completed)
- `ping` - Test event (logged but ignored)

### Event Payload Example

When a workflow completes, GitHub sends:

```json
{
  "action": "completed",
  "workflow_run": {
    "id": 123456,
    "name": "CI",
    "head_branch": "main",
    "head_sha": "abc123",
    "status": "completed",
    "conclusion": "success",
    "html_url": "https://github.com/owner/repo/actions/runs/123456",
    "created_at": "2025-11-17T17:00:00Z",
    "updated_at": "2025-11-17T17:05:00Z"
  },
  "repository": {
    "name": "repo",
    "full_name": "owner/repo"
  }
}
```

ceye parses this and updates the dashboard in real-time.

## Production Deployment

For production use, instead of ngrok, you should:

1. **Deploy ceye to a server with public IP**
2. **Use HTTPS** (Let's Encrypt recommended)
3. **Configure webhook URL**: `https://your-server.com:9090/webhooks/github`
4. **Set webhook secret** for security:
   ```bash
   ceye --webhook-secret "your-secret-here"
   ```

5. **Configure firewall** to allow incoming on port 9090

## Security Considerations

- **Webhook secret**: Use `--webhook-secret` flag to verify webhook authenticity
- **HTTPS only**: GitHub requires HTTPS for webhooks in production
- **Firewall**: Restrict webhook port to GitHub IP ranges if possible
- **Rate limiting**: GitHub respects your server's rate limits

## FAQ

**Q: Do I need ngrok for production?**
A: No, ngrok is only for local development. In production, use a public server.

**Q: Can I use multiple repos?**
A: Yes, configure a webhook for each repo pointing to the same ceye instance.

**Q: What if ngrok URL changes?**
A: You'll need to update the webhook configuration with the new URL.

**Q: Can I use a custom domain with ngrok?**
A: Yes, with ngrok paid plans. See: https://ngrok.com/docs/guides/custom-domains

**Q: How do I see webhook delivery history?**
A: In GitHub UI: Settings → Webhooks → Click webhook → Recent Deliveries

## Additional Resources

- [GitHub Webhooks Documentation](https://docs.github.com/en/developers/webhooks-and-events/webhooks)
- [ngrok Documentation](https://ngrok.com/docs)
- [gh CLI Documentation](https://cli.github.com/manual/)
