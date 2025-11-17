# Webhook Implementation Guide

**Status**: Infrastructure complete, integration pending  
**Last Updated**: 2025-11-16

## Overview

Webhooks provide real-time event notifications from CI providers (GitHub, Azure DevOps) instead of polling. This guide covers the complete webhook implementation for ceye.

## Current Status

✅ **Complete**:
- Webhook test server (`cmd/webhook-test`)
- Automation scripts (`scripts/start-with-webhooks.sh`)
- ngrok integration for localhost testing
- End-to-end testing with real GitHub webhooks
- Comprehensive logging and debugging

⚠️ **Pending**:
- Integration with main ceye application
- UI updates for webhook status
- Configuration management
- Production deployment guidance

## Quick Start

### Automated Setup (Recommended)

```bash
cd /Users/honk/code/ceye
./scripts/start-with-webhooks.sh
```

This script:
1. Starts webhook server on port 9090
2. Starts ngrok tunnel
3. Displays webhook URL
4. Provides GitHub setup commands
5. Tails webhook logs
6. Cleans up on Ctrl+C

### Manual Setup

**Terminal 1** - Start webhook server:
```bash
./bin/webhook-test
```

**Terminal 2** - Start ngrok:
```bash
ngrok http 9090
# Copy the HTTPS URL (e.g., https://abc123.ngrok-free.app)
```

**Configure GitHub webhook**:
1. Go to: `https://github.com/YOUR-ORG/YOUR-REPO/settings/hooks/new`
2. Payload URL: `https://YOUR-NGROK-URL.ngrok-free.app/webhooks/github`
3. Content type: `application/json`
4. Events: Select "Workflow runs"
5. Add webhook

**Test it**:
```bash
# Trigger a workflow
git commit --allow-empty -m "Test webhook"
git push
```

Watch Terminal 1 for the webhook payload!

## Architecture

### Webhook Server

**Endpoints**:
- `POST /webhooks/github` - GitHub webhook receiver
- `POST /webhooks/azure` - Azure DevOps webhook receiver
- `GET /webhooks/health` - Health check
- `GET /` - Status page

**Features**:
- JSON payload logging
- Pretty-printed output
- Error handling
- Event type detection

### Key Events

**GitHub** (`workflow_run` event):
- Workflow started
- Workflow completed
- Workflow failed
- Workflow cancelled

**Azure DevOps** (Build service hooks):
- Build started
- Build completed
- Build failed

## Payload Structure

### GitHub `workflow_run` Event

```json
{
  "action": "completed",
  "workflow_run": {
    "id": 123456789,
    "name": "CI",
    "status": "completed",
    "conclusion": "success",
    "head_branch": "main",
    "head_sha": "abc123",
    "created_at": "2025-11-16T00:00:00Z",
    "updated_at": "2025-11-16T00:05:00Z",
    "html_url": "https://github.com/org/repo/actions/runs/123456789"
  },
  "repository": {
    "full_name": "org/repo"
  }
}
```

### Azure DevOps Build Hook

```json
{
  "eventType": "build.complete",
  "resource": {
    "id": 12345,
    "buildNumber": "20251116.1",
    "status": "completed",
    "result": "succeeded",
    "sourceBranch": "refs/heads/main",
    "sourceVersion": "abc123",
    "startTime": "2025-11-16T00:00:00Z",
    "finishTime": "2025-11-16T00:05:00Z",
    "url": "https://dev.azure.com/org/project/_build/results?buildId=12345"
  }
}
```

## Local Development with ngrok

### Why ngrok?

- Exposes localhost to internet
- Provides HTTPS (required by GitHub)
- Free tier sufficient for development
- Easy to use

### ngrok Setup

```bash
# Install
brew install ngrok

# Run (no auth needed for basic use)
ngrok http 9090
```

### ngrok Free Tier

**Includes**:
- 1 concurrent tunnel
- HTTPS support
- Random URLs
- No auth required

**Limitations**:
- URL changes on restart
- Rate limiting (40 req/min)
- Connection time limits

**Paid tier** ($8/month):
- Custom domains
- Reserved URLs
- Higher rate limits
- Better for production

## Integration Plan

### Phase 1: Webhook Receiver (Week 1)

**1.1 Create Webhook Server** (2 days)
- Move from `cmd/webhook-test` to `internal/webhooks`
- Add to main ceye binary
- Configuration via `ceye.yaml`

**1.2 Event Parsing** (1 day)
- Parse GitHub `workflow_run` events
- Parse Azure DevOps build hooks
- Convert to `core.Run` format

**1.3 Integration with Store** (1 day)
- Send webhook events to store
- Merge with polled data
- Deduplication logic

**1.4 Testing** (1 day)
- Unit tests for parsing
- Integration tests with mock webhooks
- End-to-end tests with ngrok

### Phase 2: UI Enhancements (Week 2)

**2.1 Webhook Status Panel** (2 days)
- Show webhook connection status
- Display last received event
- Show event rate
- Error indicators

**2.2 Hybrid Mode** (2 days)
- Support webhooks + polling
- Fallback to polling if webhook fails
- Configuration per provider

**2.3 Migration Path** (1 day)
- Detect if webhooks configured
- Auto-enable hybrid mode
- Documentation for users

### Phase 3: Production Deployment (Week 3)

**3.1 Security** (2 days)
- HMAC signature verification (GitHub)
- Basic auth (Azure DevOps)
- Rate limiting
- Request validation

**3.2 Configuration** (2 days)
- Add webhook URLs to config
- Secret management
- Multi-repo support

**3.3 Documentation** (1 day)
- Setup guides per provider
- Troubleshooting
- Best practices

## Configuration Example

```yaml
# ceye.yaml
providers:
  - type: github
    display_name: "GitHub"
    repos:
      - owner: "myorg"
        repo: "myrepo"
    webhook:
      enabled: true
      secret: "${GITHUB_WEBHOOK_SECRET}"
      path: "/webhooks/github"
  
  - type: azure
    display_name: "Azure DevOps"
    org: "myorg"
    projects:
      - name: "MyProject"
    webhook:
      enabled: true
      secret: "${AZURE_WEBHOOK_SECRET}"
      path: "/webhooks/azure"

server:
  port: 8080
  webhooks_enabled: true
```

## Security Best Practices

### GitHub Webhooks

1. **Use webhook secrets** - Verify HMAC signature
2. **HTTPS only** - Required by GitHub
3. **Validate payloads** - Check event type and structure
4. **Rate limiting** - Prevent abuse
5. **Log suspicious requests** - Monitor for attacks

### Azure DevOps

1. **Use basic auth** - Username/password for service hooks
2. **HTTPS only** - Encrypt in transit
3. **Validate payloads** - Check event type
4. **IP whitelist** - Azure DevOps IPs only (optional)

## Troubleshooting

### Webhook Not Receiving Events

1. **Check ngrok** - Is tunnel running?
2. **Check GitHub config** - Is webhook enabled?
3. **Check logs** - Any errors in webhook server?
4. **Test with curl**:
   ```bash
   curl -X POST http://localhost:9090/webhooks/github \
     -H "Content-Type: application/json" \
     -d '{"action":"completed","workflow_run":{"id":123}}'
   ```

### Events Delayed

1. **GitHub delivery** - Check webhook delivery page
2. **ngrok rate limit** - Upgrade if needed
3. **Network issues** - Check internet connection

### Invalid Signature

1. **Check secret** - Must match GitHub webhook config
2. **Check algorithm** - Should be SHA-256
3. **Check header** - `X-Hub-Signature-256`

## Webhooks vs Polling

### When to Use Webhooks

✅ **Good for**:
- Real-time updates (< 10s latency)
- Many repositories
- Active development
- Reducing API calls

❌ **Not needed for**:
- Few repositories (< 5)
- Infrequent updates
- Stable projects
- Simple setups

### Recommendation

**Start with polling** - Simpler, works everywhere
**Add webhooks later** - When you need real-time updates

## Testing

### Test Webhook Locally

```bash
# Start webhook test server
./bin/webhook-test

# In another terminal, send test event
curl -X POST http://localhost:9090/webhooks/github \
  -H "Content-Type: application/json" \
  -d '{
    "action": "completed",
    "workflow_run": {
      "id": 123,
      "name": "Test",
      "status": "completed",
      "conclusion": "success"
    }
  }'
```

### Test with Real GitHub Events

1. Set up ngrok tunnel
2. Configure webhook in GitHub
3. Push a commit or trigger workflow
4. Watch logs for webhook payload

## References

- **Test server**: `cmd/webhook-test/main.go`
- **Automation**: `scripts/start-with-webhooks.sh`
- **GitHub webhook docs**: https://docs.github.com/webhooks
- **Azure service hooks**: https://learn.microsoft.com/azure/devops/service-hooks/

## Next Steps

1. Review this guide
2. Test webhook server locally
3. Set up ngrok tunnel
4. Configure GitHub webhook
5. Verify events arrive
6. Plan integration into main app

---

**Status**: Infrastructure ready, waiting for integration into main ceye application.
