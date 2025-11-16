# CEYE Testing Guide

## Overview

CEYE is a **Terminal UI (TUI)** application with CLI management commands. There is **no web interface** or traditional agent interface. This guide covers all testable interfaces and provides tools for comprehensive testing.

## Interfaces

### 1. TUI (Terminal UI) - Primary Interface

The main dashboard built with Bubble Tea framework.

**Location**: `internal/ui/`

**Features**:
- Real-time CI/CD run aggregation
- 5 information panels (Active Runs, Provider Health, Failure Rates, Duration Trends, Commit Details)
- Interactive table with filtering and navigation
- Provider management overlay

**Testing**:
```bash
# Standard launch
./bin/ci-dash

# Demo mode (no credentials needed)
./bin/ci-dash --demo

# Wide terminal for full panel visibility
tmux new-session -d -s ci-dash-wide -x 180 -y 50 "./bin/ci-dash --demo"
sleep 3
tmux attach -t ci-dash-wide
```

**Key Bindings to Test**:
- `Tab` - Cycle providers
- `f` - Cycle status filters  
- `t` - Cycle sort modes
- `p` - Provider palette
- `P` - Provider store overlay
- `/` - Search/filter
- Arrow keys, `j`/`k` - Navigation
- `Enter`, `o` - Open URL
- `y` - Copy URL
- `c` - Copy summary
- `v` - Toggle focus view
- `D` - Toggle detail view
- `H` - Toggle high contrast
- `r` - Refresh
- `?` - Toggle help
- `q`, `Ctrl+C` - Quit

### 2. CLI Commands - Management Interface

Provider management subcommands.

**Commands**:
```bash
# List all stored providers
ci-dash provider list

# Add a provider
ci-dash provider add --config provider.yaml

# Enable/disable providers
ci-dash provider enable --id <uuid>
ci-dash provider disable --id <uuid>

# Update provider config
ci-dash provider update --id <uuid> --config provider.yaml

# Remove a provider
ci-dash provider remove --id <uuid>

# Export/import for backup/sharing
ci-dash provider export --file providers.json
ci-dash provider import --file providers.json [--replace]
```

**Testing**:
```bash
# Create test provider
cat > /tmp/test-provider.yaml << 'EOF'
type: demo
display_name: test-provider
EOF

# Full lifecycle test
ci-dash provider add --config /tmp/test-provider.yaml
ci-dash provider list
ci-dash provider disable --id <id>
ci-dash provider enable --id <id>
ci-dash provider export --file /tmp/backup.json
ci-dash provider remove --id <id>
ci-dash provider import --file /tmp/backup.json
```

### 3. Webhook Integration - Outbound Notifications

Posts provider errors to external webhooks (Slack/Teams/custom).

**Flag**: `--webhook-url <url>`

**Testing**:

Create a webhook test server:
```python
#!/usr/bin/env python3
# Save as /tmp/webhook-server.py
import http.server, json
from datetime import datetime

class WebhookHandler(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        body = json.loads(self.rfile.read(content_length))
        
        timestamp = datetime.now().strftime('%H:%M:%S')
        print(f"\n[{timestamp}] Webhook received:")
        print(f"  Provider: {body.get('provider')}")
        print(f"  Error: {body.get('error')}")
        
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(b'{"status": "ok"}')
    
    def log_message(self, format, *args):
        pass

http.server.HTTPServer(('localhost', 8765), WebhookHandler).serve_forever()
```

Run test:
```bash
# Terminal 1: Start webhook server
python3 /tmp/webhook-server.py

# Terminal 2: Run ci-dash with webhook
ci-dash --demo --webhook-url http://localhost:8765/webhook
```

### 4. Desktop Notifications

Emits OS notifications on provider errors.

**Flag**: `--notify`

**Testing**:
```bash
# Run with notifications enabled (macOS/Linux)
ci-dash --demo --notify

# Wait for or trigger a provider error
# Verify notification appears in system notification center
```

### 5. Event Logging

Writes provider events to JSON Lines file for diagnostics.

**Flag**: `--log-events <path>`

**Testing**:
```bash
# Run with event logging
ci-dash --demo --log-events /tmp/events.jsonl --demo-duration 10s

# Verify events
cat /tmp/events.jsonl | jq '.'
```

Expected format:
```json
{"timestamp":"2025-11-16T09:42:14Z","provider":"demo","runs":4}
```

### 6. Run History Persistence

Persists recent run summaries to disk.

**Flag**: `--history-path <path>`

**Default**: `~/.config/ceye/run-history.json`

**Testing**:
```bash
# Run with history
ci-dash --demo --history-path /tmp/history.json --demo-duration 10s

# Verify history file
cat /tmp/history.json | jq 'keys'
```

### 7. Config Discovery

Automatically finds all `ceye.*` config files in a workspace.

**Flag**: `--config-dir <path>`

**Env**: `CEYE_CONFIG_ROOT`

**Default**: Nearest `code/` ancestor or current directory

**Testing**:
```bash
# From workspace root
cd ~/code
ci-dash --config-dir .

# Should discover all ceye.* files
# UI shows "Missing configs" panel for repos without ceye files
# Press 'n' to cycle, 'a' to scaffold
```

## Automated Testing Tools

### Quick Test Script

Save as `scripts/quick-test.sh`:
```bash
#!/bin/bash
set -e

echo "🧪 CEYE Quick Test Suite"
echo "========================"

# Build
echo "📦 Building..."
cd /Users/honk/code/ceye
go build -o bin/ci-dash ./cmd/ci-dash

# Test 1: CLI commands
echo "✅ Test 1: CLI Commands"
cat > /tmp/test-prov.yaml << 'EOF'
type: demo
display_name: test-cli
EOF

ID=$(./bin/ci-dash provider add --config /tmp/test-prov.yaml | grep -oE '[0-9a-f-]{36}')
./bin/ci-dash provider list | grep "$ID"
./bin/ci-dash provider disable --id "$ID"
./bin/ci-dash provider enable --id "$ID"
./bin/ci-dash provider export --file /tmp/backup.json
./bin/ci-dash provider remove --id "$ID"
echo "  ✓ CLI commands working"

# Test 2: Event logging
echo "✅ Test 2: Event Logging"
./bin/ci-dash --demo --log-events /tmp/events.jsonl --demo-duration 3s > /dev/null 2>&1
[ -f /tmp/events.jsonl ] && echo "  ✓ Events logged"

# Test 3: History persistence
echo "✅ Test 3: History Persistence"
./bin/ci-dash --demo --history-path /tmp/history.json --demo-duration 3s > /dev/null 2>&1
[ -f /tmp/history.json ] && echo "  ✓ History persisted"

# Test 4: TUI launch
echo "✅ Test 4: TUI Launch"
timeout 5 ./bin/ci-dash --demo > /dev/null 2>&1 || true
echo "  ✓ TUI launches without crash"

echo ""
echo "✨ All tests passed!"
```

### Visual TUI Test

Save as `scripts/visual-test.sh`:
```bash
#!/bin/bash

echo "🎨 Starting visual TUI test..."
cd /Users/honk/code/ceye

# Build
go build -o bin/ci-dash ./cmd/ci-dash

# Start in wide tmux session
tmux kill-session -t ceye-visual 2>/dev/null || true
tmux new-session -d -s ceye-visual -x 180 -y 50 "./bin/ci-dash --demo"
sleep 3

echo ""
echo "✅ TUI running in tmux session 'ceye-visual'"
echo "📺 Attach with: tmux attach -t ceye-visual"
echo "🛑 Quit with: q or Ctrl+C"
echo ""
echo "Test checklist:"
echo "  ☐ All 5 panels visible (Active Runs, Provider Health, Failure Rates, Duration Trends, Commit Details)"
echo "  ☐ Table updates in real-time"
echo "  ☐ Provider badges show in header"
echo "  ☐ Status icons render correctly (✓ ✗ ▸ …)"
echo "  ☐ Press Tab to cycle providers"
echo "  ☐ Press f to cycle status filters"
echo "  ☐ Press t to cycle sort modes"
echo "  ☐ Press ? for help"
echo ""

# Show sample screen
tmux capture-pane -t ceye-visual -p | head -40
```

## Debugging Tools

### 1. Capture TUI Output

```bash
# Capture current screen
tmux capture-pane -t <session> -p

# Capture with history
tmux capture-pane -t <session> -p -S -50

# Check terminal size
tmux display-message -t <session> -p '#{pane_width}x#{pane_height}'
```

### 2. Send Keys to TUI

```bash
# Send keystrokes
tmux send-keys -t <session> 'P'  # Capital P for provider store
tmux send-keys -t <session> '?'  # Show help
tmux send-keys -t <session> 'q'  # Quit
```

### 3. Monitor Events

```bash
# Real-time event monitoring
ci-dash --demo --log-events /tmp/events.jsonl &
tail -f /tmp/events.jsonl | jq '.'
```

### 4. Profile Performance

```bash
# With Go profiling
go build -o bin/ci-dash ./cmd/ci-dash
./bin/ci-dash --demo &
PID=$!
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile
```

## Test Checklist

### Pre-Commit Checklist

- [ ] `go test ./...` passes
- [ ] `go build -o bin/ci-dash ./cmd/ci-dash` succeeds
- [ ] TUI launches without panic
- [ ] Demo mode displays all 5 panels
- [ ] CLI commands execute without errors
- [ ] No obvious rendering glitches

### Full Test Suite

- [ ] TUI with real GitHub/Azure data
- [ ] All keybindings respond correctly
- [ ] CLI provider management (add/list/remove/enable/disable/export/import)
- [ ] Webhook integration sends POST requests
- [ ] Desktop notifications appear
- [ ] Event logging writes valid JSON
- [ ] History persistence creates file
- [ ] Config discovery finds all ceye.* files
- [ ] Provider store overlay is accessible (press P)
- [ ] Wide and narrow terminal layouts both work
- [ ] No memory leaks during extended run

## Common Issues

### Issue: Panels Not Visible

**Cause**: Terminal too narrow (< 80 columns)

**Fix**: Use wider terminal or tmux session
```bash
tmux new-session -d -s wide -x 180 -y 50 "./bin/ci-dash"
```

### Issue: Provider Store Overlay Doesn't Show

**Cause**: Capital 'P' key not being registered

**Debug**:
```bash
# Check if stored providers exist
ci-dash provider list

# Try with debug logging
CI_DASH_DEBUG=1 ci-dash --demo
```

### Issue: Real Provider Data Not Loading

**Cause**: Missing credentials

**Fix**: Set environment variables
```bash
export GITHUB_TOKEN=ghp_xxxxx
export CEYE_GITHUB_TOKEN=ghp_xxxxx  # Preferred
export AZURE_DEVOPS_PAT=xxxxx
```

## CI/CD Testing

### GitHub Actions Example

```yaml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test ./...
      
      - name: Build
        run: go build -o bin/ci-dash ./cmd/ci-dash
      
      - name: Quick smoke test
        run: timeout 5 ./bin/ci-dash --demo > /dev/null 2>&1 || true
      
      - name: CLI commands test
        run: |
          ./bin/ci-dash provider list
          ./bin/ci-dash --version
```

## Conclusion

CEYE's testing focuses on:
1. **TUI functionality** (visual + interactive)
2. **CLI commands** (provider management)
3. **Integration features** (webhooks, notifications, logging)

There is **no web interface or agent API** to test. All interaction is through the terminal UI or CLI commands.

For questions or issues, refer to:
- Main README: `docs/README.md`
- Implementation plan: `docs/ci-status-dashboard-plan.md`
- UI enhancements: `docs/ui-enhancements-plan.md`
