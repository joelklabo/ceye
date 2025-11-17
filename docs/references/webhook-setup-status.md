# Webhook Setup - Ready to Use

**Date**: 2025-11-16  
**Status**: Test server built, automation scripts created

## ✅ What's Ready

### 1. Webhook Test Server
- **Binary**: `bin/webhook-test` (8.2MB)
- **Endpoints**:
  - POST `/webhooks/github` - Receives GitHub webhooks
  - POST `/webhooks/azure` - Receives Azure webhooks
  - GET `/webhooks/health` - Health check
  - GET `/` - Status page
- **Features**: Logs all payloads, pretty-prints JSON

### 2. Automation Scripts

####  `scripts/start-with-webhooks.sh`
Automated one-command setup:
```bash
./scripts/start-with-webhooks.sh
```

**What it does**:
1. Installs ngrok (if needed)
2. Starts webhook server on port 9090
3. Starts ngrok tunnel
4. Displays webhook URL
5. Provides setup commands for GitHub
6. Tails logs
7. Cleans up on Ctrl+C

### 3. Documentation
- `docs/WEBHOOK_TEST_GUIDE.md` - Complete step-by-step guide
- `docs/references/webhook-research.md` - Full research (843 lines)
- `WEBHOOK_RESEARCH_SUMMARY.md` - Executive summary

## 🚀 Quick Start

###Option 1: Automated (Recommended)

```bash
cd /Users/honk/code/ceye
./scripts/start-with-webhooks.sh
```

Follow the displayed instructions to configure GitHub webhook.

### Option 2: Manual

**Terminal 1**:
```bash
./bin/webhook-test
```

**Terminal 2**:
```bash
ngrok http 9090
# Copy the HTTPS URL
```

**Browser**:
1. Go to: https://github.com/joelklabo/ceye/settings/hooks/new
2. Payload URL: `https://YOUR-NGROK-URL.ngrok.io/webhooks/github`
3. Content type: `application/json`
4. Events: `Workflow runs`
5. Add webhook

**Test it**:
```bash
# Trigger a workflow
git commit --allow-empty -m "Test webhook"
git push
```

Watch Terminal 1 for the webhook payload!

## 📊 UI Enhancements Needed

### Terminal UI (TUI)

**Current plan** (not yet implemented):
1. **Webhook indicator** in header/status bar:
   ```
   ⚡ LIVE | GitHub: Connected | Azure: Not configured
   ```

2. **Flash animation** when webhook received:
   ```
   ⚡ New workflow run received from GitHub
   ```

3. **Visual pulse** on new data:
   - Brief color change on affected rows
   - "⚡" icon next to updated runs
   - Fade after 2-3 seconds

**Implementation location**:
- Add `WebhookReceivedMsg` message type ✅ DONE
- Add `statusIconWebhook = "⚡"` constant ✅ DONE
- Modify `Model.View()` in `internal/ui/model.go`
- Add webhook status to `RunUpdatedMsg` ✅ DONE

### Web UI

**Current plan** (not yet implemented):
1. **Live indicator** in header:
   ```html
   <div class="webhook-status live">
     <span class="pulse"></span> LIVE
   </div>
   ```

2. **Toast notification** on webhook:
   ```
   ⚡ Workflow run updated (< 1s ago)
   ```

3. **Visual highlight** on updated rows:
   - Brief green glow
   - Fade animation
   - Row slide-in effect

**Implementation location**:
- Modify `internal/server/web/index.html`
- Add CSS animations to `internal/server/web/style.css`
- Update WebSocket handler in `internal/server/web/app.js`
- Add `webhookReceived` event type to WebSocket messages

## 🔧 Integration Steps

To integrate webhooks into main ceye app:

### 1. Wire Webhook Server into Main App (1 day)

```go
// cmd/ci-dash/main.go
import "github.com/joelklabo/ceye/internal/webhooks"

func main() {
    // ... existing setup ...
    
    // Start webhook server
    webhookServer := webhooks.NewServer(webhooks.Config{
        Port: 9090,
        GitHubSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
    })
    
    // Forward webhook events to store
    go func() {
        for event := range webhookServer.Events() {
            store.Merge(event) // Same as polling events!
        }
    }()
    
    go webhookServer.Start(ctx)
    
    // ... rest of app ...
}
```

### 2. Add Webhook Status to UI (1 day)

**TUI changes**:
```go
// internal/ui/model.go - Update() method
case WebhookReceivedMsg:
    // Flash webhook indicator
    // Mark affected runs with ⚡
    // Trigger brief animation

// View() method
func (m Model) View() string {
    header := m.renderHeader()
    webhookStatus := m.renderWebhookStatus() // NEW
    // ... rest of view
}
```

**Web changes**:
```javascript
// internal/server/web/app.js
ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    
    if (data.webhookEvent) {
        showWebhookNotification(); // NEW
        highlightUpdatedRows(data.runs); // NEW
    }
    
    updateTable(data.runs);
}
```

### 3. Add Configuration (1 day)

```yaml
# ceye.yaml
webhooks:
  enabled: true
  port: 9090
  github_secret: "${GITHUB_WEBHOOK_SECRET}"
  
providers:
  - type: github
    mode: webhook  # NEW: webhook, polling, or hybrid
    repos:
      - owner: joelklabo
        repo: ceye
```

## 📋 Current Status

| Component | Status | Notes |
|-----------|--------|-------|
| Webhook server | ✅ Built | `internal/webhooks/server.go` |
| Test binary | ✅ Ready | `bin/webhook-test` |
| Automation script | ✅ Ready | `scripts/start-with-webhooks.sh` |
| Documentation | ✅ Complete | 3 docs created |
| TUI integration | ⏳ Planned | Message types added |
| Web UI integration | ⏳ Planned | Needs implementation |
| Main app integration | ⏳ Planned | ~3 days work |

## 🎯 Next Steps

**Immediate** (you can do now):
1. Run `./scripts/start-with-webhooks.sh`
2. Configure GitHub webhook
3. Trigger a workflow
4. Verify webhook received in logs
5. Check ngrok dashboard at http://localhost:4040

**After validation**:
1. Implement TUI webhook indicator (1 day)
2. Implement Web UI webhook indicator (1 day)
3. Integrate into main ceye app (1 day)
4. Test end-to-end
5. Document for users

## 💡 Design Notes

### Why Two Terminal Windows?
- **Terminal 1**: Webhook server logs (see payloads)
- **Terminal 2**: ngrok (provides public URL)
- Could combine into one with tmux/screen

### Why ngrok?
- Localhost isn't internet-accessible
- ngrok creates secure tunnel: `internet → ngrok → localhost:9090`
- Free tier works fine for testing
- $8/month for static URL (optional)

### Webhook vs Polling UX
**Polling** (current):
- Updates every 15-60 seconds
- Feels "slow"
- No visual indication of data source

**Webhooks** (with indicators):
- Updates < 1 second
- ⚡ Live indicator shows real-time connection
- Flash/animation on updates
- Users know it's "live"

### Animation Ideas
**TUI**:
- Brief color inversion on updated row
- ⚡ icon fades in/out
- Status bar pulse
- Sound (terminal bell) optional

**Web**:
- CSS keyframe animation on row
- Toast notification slides in
- Green pulse on live indicator
- Confetti on success (fun!)

## 🐛 Troubleshooting

**ngrok not found**:
```bash
brew install ngrok
```

**Port 9090 already in use**:
```bash
lsof -i :9090
kill <PID>
```

**Webhook not receiving**:
- Check ngrok is running: http://localhost:4040
- Verify webhook URL is correct
- Check GitHub webhook "Recent Deliveries"
- Look for errors in webhook server logs

**GitHub CLI auth error**:
```bash
gh auth login
```

## 📚 Resources

- Test guide: `docs/WEBHOOK_TEST_GUIDE.md`
- Research: `docs/references/webhook-research.md`
- GitHub webhooks docs: https://docs.github.com/en/webhooks
- ngrok docs: https://ngrok.com/docs
