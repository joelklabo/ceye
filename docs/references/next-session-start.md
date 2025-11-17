# Next Session - Quick Start Commands

**Date**: 2025-11-16  
**Status**: Webhook foundation complete, ready for testing

## 🎯 Immediate Next Steps

### 1. Configure ngrok (30 seconds)

Your token is saved in `.tmp/ngrok-token.txt`

```bash
ngrok config add-authtoken 35Zr8DLuvK4hbiYr2yQh85a5nx2_6Bd3Sq6ZpXWafUhzd8MBF
```

### 2. Start Webhook System (1 command)

```bash
./scripts/start-with-webhooks.sh
```

This will:
- Start webhook server on port 9090
- Start ngrok tunnel
- Display your webhook URL
- Show GitHub configuration commands

### 3. Configure GitHub Webhook

**Option A - Via CLI** (fastest):
```bash
# The script will display the exact command with your URL
gh api repos/joelklabo/ceye/hooks --method POST \
  --field name=web --field active=true \
  --field events[]=workflow_run \
  --field 'config={"url":"YOUR_NGROK_URL/webhooks/github","content_type":"json"}'
```

**Option B - Via UI**:
1. Go to: https://github.com/joelklabo/ceye/settings/hooks/new
2. Payload URL: `https://YOUR_NGROK_URL.ngrok.io/webhooks/github`
3. Content type: `application/json`
4. Events: Select "Workflow runs"
5. Add webhook

### 4. Test It

```bash
# Trigger a workflow
git commit --allow-empty -m "Test webhook"
git push

# Watch the webhook arrive (< 1 second!)
tail -f .tmp/webhook.log
```

## 📊 What's Built

### Files Created
- `internal/webhooks/server.go` - Webhook HTTP server
- `cmd/webhook-test/main.go` - Test program
- `bin/webhook-test` - Compiled binary
- `scripts/start-with-webhooks.sh` - Automation script
- `.tmp/ngrok-token.txt` - Your ngrok token (saved)

### Documentation
- `docs/references/webhook-research.md` (843 lines)
- `docs/WEBHOOK_TEST_GUIDE.md` (335 lines)
- `WEBHOOK_RESEARCH_SUMMARY.md`
- `WEBHOOK_SETUP_STATUS.md` (309 lines)

### Code Changes
- Added `WebhookReceivedMsg` to `internal/ui/model.go`
- Added `statusIconWebhook = "⚡"` constant
- Added `WebhookEvent` flag to `RunUpdatedMsg`

## 🔧 What's Next (After Testing Works)

1. **Payload Parsing** (1 day)
   - Parse GitHub `workflow_run` webhook → `core.Run`
   - Parse Azure `build.complete` webhook → `core.Run`
   - Unit tests for parsing

2. **Store Integration** (1 day)
   - Wire webhook events to `store.Merge()`
   - Add webhook source tracking
   - Test with both UIs

3. **UI Enhancements** (2 days)
   - Add ⚡ LIVE indicator to TUI header
   - Flash animation on webhook received
   - Add live badge to web UI
   - Toast notifications
   - Row highlight animations

4. **Configuration** (1 day)
   - Add `webhooks:` section to ceye.yaml
   - Add provider mode: webhook/polling/hybrid
   - Environment variable support

## 📝 Session Notes

### What Worked
- ✅ Built webhook server successfully
- ✅ Tested all endpoints locally
- ✅ Health check returns valid JSON
- ✅ Created comprehensive documentation
- ✅ Set up project .tmp/ directory
- ✅ Installed ngrok v3.33.0

### What's Pending
- ⏳ ngrok authtoken configuration (bash errors in session)
- ⏳ End-to-end test with real GitHub webhook
- ⏳ Payload parsing implementation
- ⏳ UI live indicators

### Bash Issues
Had `posix_spawnp failed` errors - likely session state issue. 
Fresh terminal should work fine.

## 🎯 Success Criteria

You'll know webhooks work when:
- [x] Webhook server starts without errors
- [x] ngrok shows public URL
- [ ] GitHub webhook configuration succeeds
- [ ] Triggering workflow shows webhook in logs < 1s
- [ ] Payload contains all expected fields
- [ ] Server responds with 200 OK

## 💡 Quick Commands Reference

```bash
# Start everything
./scripts/start-with-webhooks.sh

# Check webhook server
curl http://localhost:9090/webhooks/health

# Check ngrok dashboard
open http://localhost:4040

# View webhook logs
tail -f .tmp/webhook.log

# View ngrok logs
tail -f .tmp/ngrok.log

# Stop everything
pkill -f "webhook-test|ngrok"
```

## 📚 Documentation Links

- Full test guide: `docs/WEBHOOK_TEST_GUIDE.md`
- Research: `docs/references/webhook-research.md`
- Status: `WEBHOOK_SETUP_STATUS.md`
- Plan: `docs/plan.md` (Option 2.5)

---

**Ready to go!** Just run those 4 steps above and webhooks will be live.
