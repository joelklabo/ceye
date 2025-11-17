# ceye - CI/CD Monitoring Dashboard - Agent Context

**Last Updated**: 2025-11-16  
**Version**: 1.0  
**Status**: Active Development

## Project Overview

**ceye** (CI Eye) is a web-based CI/CD monitoring dashboard that aggregates workflow runs from multiple CI providers (GitHub Actions, Azure DevOps, GitLab CI) into a unified, real-time view.

### Key Features

- **Real-time Dashboard**: Modern interface with live WebSocket updates
- **Multi-provider**: GitHub, Azure DevOps, GitLab, with extensible provider system
- **Real-time**: Live updates via event channels and WebSocket
- **Resilient**: SafeProvider wrapper with panic recovery and validation
- **Tested**: 140+ tests including integration, contract, and safety tests

## Architecture

### Core Components

```
ceye/
├── cmd/ceye/              # Main application entry point
├── internal/
│   ├── core/              # Core types (Run, RunEvent, Provider, Store)
│   ├── config/            # Configuration loading and validation
│   ├── providers/         # Provider implementations
│   │   ├── github/        # GitHub Actions provider
│   │   ├── azure/         # Azure DevOps provider (in progress)
│   │   ├── gitlab/        # GitLab CI provider
│   │   ├── demo/          # Demo provider for testing
│   │   └── manager/       # Provider lifecycle management
│   ├── server/            # HTTP/WebSocket server
├── web/                   # Static UI assets
└── docs/                  # Documentation
    ├── agents.md          # This file (SOURCE OF TRUTH)
    │                      # Symlinked from: /AGENTS.md, /CLAUDE.md, /.github/copilot-instructions.md
    ├── readme.md          # Project README (symlinked from /README.md)
    ├── plan.md            # Master development plan
    └── references/        # Reference docs and guides
        ├── doc-inventory.md
        ├── testing-guide.md
        ├── webhook-guide.md
        ├── web-ui-architecture.md
        └── ui-demo.txt
```

### Data Flow

```
Provider → RunEvent → Store → UI
   ↓
SafeProvider (panic recovery, validation)
   ↓
Manager (lifecycle, health tracking)
```

### Webhook Metadata Tracking

**Added**: 2025-11-17 (Commit: a8370d7)

The Store now tracks webhook metadata for real-time visibility:

```go
type WebhookMetadata struct {
    EventType  string // "workflow_run", "ping", etc
    DeliveryID string // GitHub delivery ID
    Payload    string // Full webhook payload as JSON string
    ReceivedAt time.Time
}

type ProviderHealth struct {
    LastError    time.Time
    ErrorCount   int
    LastSuccess  time.Time
    MessageCount int              // Total messages received
    LastWebhook  *WebhookMetadata // Last webhook received
}
```

**Key Points**:
- Webhook metadata is populated in `internal/webhooks/server.go` when webhooks arrive
- Store tracks up to 100 recent webhooks per provider in `webhookHistory`
- ProviderHealth includes `MessageCount` and `LastWebhook` for UI display
- Main event processor merges Store health (with webhooks) into server status
- Frontend displays last webhook event type and message count

**Integration**:
1. Webhook arrives → `handleGitHub()` creates `WebhookMetadata`
2. Attached to `RunEvent.WebhookMeta`
3. Store.Merge() tracks it in `webhookHistory` and updates `ProviderHealth`
4. Event processor calls `store.GetProviderHealth()` to merge webhook data
5. WebSocket sends updated health to UI
6. React component displays webhook info

**Important**: There are TWO health tracking systems:
- Store health: Has webhook metadata, message counts
- Main.go providerHealth: Has error tracking from event processor
- They are merged in the event loop before sending to UI

## ⚠️ CRITICAL: Temporary Files

**ALWAYS use `tmp/` directory for ALL temporary files, logs, working documents, test outputs, debug files, session notes, etc.**

```
tmp/                    # Gitignored, safe for any temporary work
├── .gitkeep           # Explains the folder purpose
├── *.log              # Log files from testing
├── *.md               # Working notes, test results, session summaries
├── *.txt              # Debug output, snapshots, captures
├── *.pid              # Process IDs
├── *.yaml             # Test configurations
└── ...                # Any temporary file

DO:
✅ Write ALL logs to tmp/
✅ Put ALL test outputs in tmp/
✅ Use tmp/ for debugging files
✅ Store session notes in tmp/
✅ Save temporary scripts in tmp/
✅ Put captured output in tmp/
✅ Use tmp/ for any file you might delete later

DO NOT:
❌ Write to /tmp/ (permission issues, wrong location)
❌ Create temp files in project root (clutters workspace)
❌ Put temp files in docs/ (only permanent docs there)
❌ Use system temp dirs (we have our own)
❌ Commit anything from tmp/ to git (all ignored)
```

**Why**: Keeps project organized, avoids permission errors, clear separation of permanent vs temporary files, easy cleanup.

**Examples**:
```bash
# Good - all in tmp/
cd /Users/honk/code/ceye
go build 2>&1 | tee tmp/build-output.txt
./bin/ceye --demo 2>&1 > tmp/demo-test.log
cat > tmp/session-notes.md << 'EOF'
...
EOF

# Bad - wrong locations
go build 2>&1 | tee /tmp/build.txt     # Wrong dir
./bin/ceye > test.log                  # Clutters root
cat > notes.md                         # Not temp location
```

## 🐛 Developer Debugging Tools

**Added**: 2025-11-17 (Task 8)

The ceye dashboard includes built-in debugging tools accessible via the **Debug Panel**.

### Accessing the Debug Panel

1. **Open the app**: `http://localhost:8080`
2. **Click the bug icon** (🐛) in the bottom-right corner
3. **Panel slides in from the right** with three tabs

### Debug Panel Features

#### 1. WebSocket Inspector

**What it shows**:
- All WebSocket messages received (last 50)
- Message type (e.g. `runs_update`, `snapshot`)
- Timestamp of receipt
- Direction indicator (↓ received)
- Full JSON payload (click to expand)

**How to use**:
```
1. Open Debug Panel (bug icon)
2. Click "WebSocket" tab
3. Watch messages arrive in real-time
4. Click any message to see full JSON
5. Click "Clear" to reset the list
```

**Example debugging session**:
```
Problem: Runs not updating in UI
Steps:
1. Open Debug Panel → WebSocket tab
2. Check if messages arriving (should see runs_update every ~15s)
3. If no messages → WebSocket connection issue
4. If messages arriving → check JSON payload for runs data
5. Expand message → verify Run objects have correct fields
6. Compare with UI → identify transformation issue
```

**What to look for**:
- ✅ Messages arriving regularly (every 15-30s)
- ✅ Message type is `runs_update`
- ✅ Runs array contains expected workflow runs
- ✅ Provider health shows webhook metadata
- ❌ No messages → Check WebSocket connection
- ❌ Empty Runs array → Check provider configuration
- ❌ Missing fields → Check backend RunEvent structure

#### 2. Logs Tab (Coming Soon)

Will capture frontend console.log() in real-time

#### 3. Events Tab (Coming Soon)

Will show visual timeline of all system events

### Additional Debugging Tools

#### Browser DevTools

**Console**:
```javascript
// In browser console
localStorage.debugPanelOpen = 'true'  // Auto-open on reload
localStorage.removeItem('debugPanelOpen')  // Close on reload
```

**Network Tab**:
- Filter by "WS" to see WebSocket traffic
- Shows connection handshake, frames, close events
- Useful for low-level WebSocket debugging

**React DevTools**:
- Install React Developer Tools extension
- Inspect component state
- View DashboardContext values
- Track re-renders

#### Server Logs

**View real-time server logs**:
```bash
# Standard output
./bin/ceye --port 8080

# With timestamp and component tags
./bin/ceye --port 8080 2>&1 | grep -E "webhook|github|store"

# Save to file
./bin/ceye --port 8080 2>&1 | tee tmp/server.log

# Follow specific provider
./bin/ceye --port 8080 2>&1 | grep "github:"
```

**Key log patterns**:
```
✅ "Web server starting on http://localhost:8080" - Server ready
✅ "github: webhook mode enabled" - Webhooks active
✅ "✅ Parsed GitHub webhook" - Webhook received
❌ "Error parsing GitHub webhook" - Malformed payload
❌ "Warning: event channel full" - System overload
```

#### Testing with Demo Mode

**Quick debugging setup**:
```bash
# Start with demo provider
./bin/ceye --demo --demo-runs 10 --port 8080

# Demo generates fake runs every 15s
# Perfect for testing UI updates without real CI data
```

#### Webhook Testing

**Test webhook delivery**:
```bash
# 1. Start ngrok
ngrok http 9090

# 2. Get tunnel URL
https://abc123.ngrok.io

# 3. Configure GitHub webhook
gh api repos/owner/repo/hooks --method POST \
  -f url="https://abc123.ngrok.io/webhooks/github" \
  -f content_type="json" \
  -f events='["workflow_run"]'

# 4. Watch Debug Panel → WebSocket tab
# Should see runs_update with WebhookMeta field

# 5. Check server logs
./bin/ceye 2>&1 | grep "webhook"
```

#### Store Inspection

**Check Store state** (add temporary debug endpoint):
```bash
# View all runs in store
curl http://localhost:8080/api/runs

# View provider health
curl http://localhost:8080/api/health
```

### Common Debugging Scenarios

#### Scenario 1: WebSocket Not Connecting

**Symptoms**: UI shows "OFFLINE", no activity updates

**Debug steps**:
1. Open Debug Panel → Should see "No messages yet" initially
2. Wait 5s → Should see first `snapshot` message
3. If no messages:
   - Check browser console for WebSocket errors
   - Verify server running on correct port
   - Check firewall/proxy settings

**Fix**: Usually browser cache - hard reload (Cmd+Shift+R)

#### Scenario 2: Webhooks Not Working

**Symptoms**: Polling works but no real-time updates

**Debug steps**:
1. Open Debug Panel → WebSocket tab
2. Trigger a workflow run on GitHub
3. Check if message arrives within 2-3 seconds
4. If no message:
   - Verify ngrok tunnel is running
   - Check GitHub webhook delivery page
   - Look for "X-GitHub-Delivery" header in server logs

**Fix**: Re-create webhook with correct ngrok URL

#### Scenario 3: UI Flicker/Performance

**Symptoms**: UI stutters, runs re-animate on updates

**Debug steps**:
1. Open React DevTools → Profiler
2. Start recording
3. Wait for WebSocket message
4. Stop recording
5. Check which components re-rendered

**Fix**: Add React.memo() to expensive components

#### Scenario 4: Missing Run Data

**Symptoms**: Runs show in logs but not in UI

**Debug steps**:
1. Open Debug Panel → WebSocket tab
2. Expand latest `runs_update` message
3. Check Runs array → Should have run objects
4. Compare with UI → Identify missing data
5. Check DashboardContext → useDashboard() hook

**Fix**: Usually data transformation issue in App.tsx

### Debugging Tips

**Performance**:
- Use `tmp/` for log files (never `/tmp/` or project root)
- Debug Panel persists state in localStorage
- Clear messages regularly to avoid memory issues

**Best Practices**:
- Always check Debug Panel first (fastest way to see live data)
- Use browser DevTools for deep inspection
- Save problematic payloads to `tmp/debug-payload.json`
- Add temporary console.log() with prefix: `console.log('[DEBUG]', ...)`

**When to use what**:
- 🐛 **Debug Panel**: Real-time WebSocket message inspection
- 🔍 **Browser Console**: Quick checks, one-off debugging
- 📊 **React DevTools**: Component state, props, re-renders
- 📝 **Server Logs**: Backend issues, API errors, webhook delivery
- 🧪 **Demo Mode**: UI testing without real CI data

### Key Abstractions

**Provider Interface** (The "Agent" Interface)
```go
type Provider interface {
    Name() string
    Start(ctx context.Context, out chan<- RunEvent) error
}
```

**Core Types**
```go
type Run struct {
    ID           string
    Provider     string
    Repo         string
    WorkflowName string
    Status       RunStatus
    Conclusion   string
    Branch       string
    CommitSHA    string
    StartedAt    time.Time
    UpdatedAt    time.Time
    Duration     time.Duration
    URL          string
}

type RunEvent struct {
    Provider  string
    Runs      []Run
    Timestamp time.Time
    Err       error
    Message   string
    Health    map[string]ProviderHealth
}
```

## Current State

### ✅ Completed Features (Phases 1-5)

**Phase 1: Core Dashboard**
- Dashboard with real-time updates
- Provider abstraction layer
- Real-time event streaming
- Store with normalized Run data
- Adaptive polling (fast when active, slow when idle)
- Workspace scanning and config auto-discovery

**Phase 2: Dashboard UI**
- HTTP server with WebSocket support
- Static HTML/CSS/JS frontend
- Real-time updates via WebSocket
- All 5 dashboard panels (Runs, Active, Health, Failures, Trends)
- Provider filtering and sorting
- Responsive layout

**Phase 3: Provider Safety**
- SafeProvider wrapper with panic recovery
- Event validation (4 validation rules)
- Graceful degradation
- Clear error logging with stack traces
- 19 safety tests passing

**Phase 4: Integration Testing**
- Provider → Store → UI flow tests
- Multiple provider coordination tests
- Health panel update tests
- Panic isolation tests
- 6 integration tests passing

**Phase 5: Contract Testing**
- Provider interface contract test suite
- 8 core contract validation tests
- Provider compliance tests
- 11 contract tests passing

**Total Test Coverage**: 140+ tests, all passing ✅

### 🚧 Current Sprint

**Priority**: Azure DevOps Provider (Option 3) - 3 weeks

See [docs/plan.md](docs/plan.md) for full details on Options 2-5.

## Development Workflow

### Building and Installing

```bash
cd /Users/honk/code/ceye
go build -o bin/ceye ./cmd/ceye
sudo cp bin/ceye /usr/local/bin/ceye
```

**Important**: Rebuild after any code changes to keep the CLI on your PATH up to date.

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/core/

# With verbose output
go test -v ./internal/providers/

# With coverage
go test -cover ./...
```

### Running the Dashboard

```bash
# Web mode (default and only mode)
ceye --port 8080

# Demo mode
ceye --demo --demo-duration 5m --port 8080

# With config
ceye --config path/to/ceye.yaml --port 8080
```

## UI Testing Strategy

### Dashboard Testing

1. **Start web server**
```bash
./bin/ceye --port 8080
```

2. **Open browser**
```
http://localhost:8080
```

3. **Verify**
- WebSocket connection establishes
- Real-time updates appear
- Provider filtering works
- Status filtering works
- Search works
- Responsive layout

<h2>Configuration</h2>

<h3>Config File Format (ceye.yaml)</h3>

```yaml
providers:
  - type: github
    display_name: "GitHub Prod"
    logo: "/static/logos/github.svg"  # Optional: custom logo path
    repos:
      - owner: "myorg"
        repo: "myrepo"
  
  - type: azure
    display_name: "Azure DevOps"
    logo: "/static/logos/azure.svg"   # Optional: uses built-in if omitted
    org: "myorg"
    projects:
      - name: "MyProject"
        pipelines: [123, 456]
  
  - type: jenkins
    display_name: "Jenkins CI"
    logo: "/static/logos/jenkins.svg"  # Custom logo for unsupported providers
    # ... jenkins-specific config
  
  - type: demo
    count: 5
    interval: "10s"

server:
  port: 8080
  host: "0.0.0.0"
```

<h3>Provider Logos</h3>

**Built-in Logos**: GitHub, Azure DevOps, GitLab (automatically used based on provider type)

**Custom Logos**:
- Add `logo` field to provider config with path to SVG file
- Logo files should be placed in `web/static/logos/` directory
- Requirements:
  - Format: SVG only (scalable, theme-compatible)
  - ViewBox: 24x24 recommended
  - File size: < 10KB
  - Colors: Use `currentColor` for theme compatibility

**Example**:
```yaml
providers:
  - type: jenkins
    display_name: "Jenkins"
    logo: "/static/logos/jenkins.svg"  # Place file in web/static/logos/
```

If no logo is specified or file is missing, the UI falls back to:
1. Built-in logo (if provider type matches: github, azure, gitlab)
2. Generic provider icon (colorful gradient circle with first letter)

<h3>Environment Variables</h3>

- `GITHUB_TOKEN` - GitHub personal access token
- `AZURE_PAT` - Azure DevOps personal access token
- `GITLAB_TOKEN` - GitLab personal access token
- `CEYE_CONFIG_ROOT` - Config directory override

<h2>Development Plan</h2>

See [docs/plan.md](docs/plan.md) for the comprehensive master plan.

<h3>Current Sprint (Options 2-5)</h3>

**Option 2: Enhanced Monitoring** (4 weeks)
- Historical data storage (SQLite)
- Trends and analytics (Chart.js)
- Alerting (Slack/Email/PagerDuty)
- Performance metrics (Prometheus/Grafana)

**Option 3: Azure DevOps Provider** (3 weeks) - **PRIORITY**
- Complete API client with auth
- Full provider implementation
- Feature parity with GitHub

**Option 4: User Experience** (4 weeks)
- Keyboard shortcuts (20+ shortcuts)
- Theme system (6 presets)
- Dashboard customization
- Advanced filtering

**Option 5: Advanced Testing** (4 weeks)
- Load testing (100 providers, 10k runs)
- Chaos engineering (fault injection)
- E2E tests (Playwright)
- Performance benchmarks

<h3>Future Roadmap</h3>

- Enterprise features (auth, RBAC, audit logs)
- Additional providers (Jenkins, CircleCI, Buildkite)
- Mobile app (React Native)
- AI/ML features (failure prediction, anomaly detection)

<h2>Testing Standards</h2>

<h3>Required for All Code</h3>

- **Unit tests** for all new code
- **Integration tests** for cross-component features
- **Contract tests** for provider implementations
- **E2E tests** for user-facing features

<h3>Test Coverage Goals</h3>

- Overall coverage: > 80%
- All critical paths tested
- All providers pass contract tests
- Zero known security issues
- All linters passing

<h3>TDD Workflow</h3>

1. Write failing test
2. Implement minimal code to pass
3. Refactor
4. Commit with tests
5. Run full test suite before push

<h2>Success Metrics</h2>

<h3>Performance Targets</h3>

- Event processing: < 10ms p99
- Store query: < 5ms p99
- WebSocket latency: < 50ms p99
- Memory usage: < 100MB at 1000 runs
- CPU usage: < 5% idle, < 20% under load

<h3>Reliability Targets</h3>

- Uptime: 99.9%
- Provider crash recovery: < 5s
- Zero data loss
- Zero UI freezes

<h3>Quality Targets</h3>

- Test coverage: > 80%
- All critical paths tested
- Zero known security issues
- Zero memory leaks
- All linters passing

<h2>React Development</h2>

<h3>Component Performance</h3>

**Use React.memo for expensive components**:
```tsx
export const StatsCards = memo(function StatsCards({ stats }: Props) {
  // Component implementation
})
```

**Use useMemo for expensive calculations**:
```tsx
const activityItems = useMemo(() => 
  runs.slice(0, 10).map(/* transform */),
  [runs]
)
```

<h3>Theme Support</h3>

**All components support light/dark modes via CSS variables**:
- Use Tailwind classes: `bg-background`, `text-foreground`
- CSS variables automatically switch with `dark` class on `<html>`
- SVG logos use `currentColor` to adapt to theme

**Adding theme toggle**:
1. ThemeContext provides `theme` and `toggleTheme()`
2. Persists to localStorage
3. Respects system preference on first load
4. Apply `dark` class to `<html>` element

<h3>Provider Icons</h3>

**Adding built-in provider logos**:
1. Create SVG component in `web/src/components/icons/logos/`
2. Use `currentColor` for fill (theme-adaptive)
3. Add to ProviderIcon component's switch statement
4. Test in both light and dark modes

**Using ProviderIcon**:
```tsx
<ProviderIcon provider="github" size="md" />
<ProviderIcon provider="custom" logoPath="/logos/custom.svg" />
<ProviderIcon provider="unknown" fallback="monogram" />
```

<h2>Common Tasks</h2>

<h3>Adding a New Provider</h3>

1. Create provider package in `internal/providers/yourprovider/`
2. Implement `Provider` interface
3. Add configuration schema
4. Write contract compliance tests
5. Add to provider factory
6. Document setup and configuration
7. Update README

<h3>Adding a New Feature</h3>

1. Check if it's in [docs/plan.md](docs/plan.md)
2. Create feature branch
3. Write tests first (TDD)
4. Implement feature
5. Update documentation
6. Run full test suite
7. Update plan.md with progress
8. Submit PR

<h3>Fixing a Bug</h3>

1. Write test that reproduces bug
2. Fix bug
3. Verify test passes
4. Run full test suite
5. Commit with test
6. No changes to plan.md needed for bugs

<h2>Troubleshooting Bash Sessions</h2>

When AI agents have issues with bash sessions (timeouts, hangs, unresponsive), use these steps:

<h3>Diagnosing Issues</h3>

1. **List active bash sessions**: Check what's running
2. **Check for hung processes**: `ps aux | grep -E "(playwright|node|ceye)"`
3. **Check port conflicts**: `lsof -i :8080` to see if port is already in use
4. **Check process tree**: See if background processes are blocking

<h3>Resolving Issues</h3>

1. **Kill specific processes**: `pkill -f "process-name"` (e.g., `pkill -f "ceye"`)
2. **Kill port users**: `kill -9 <PID>` for processes blocking ports
3. **Use new session IDs**: Don't reuse a problematic sessionId
4. **Stop bash sessions**: Use `stop_bash` tool with the sessionId (kills the whole session)

<h3>Prevention</h3>

- **Unique sessionIds**: Always use unique sessionIds for different operations
- **Appropriate timeouts**: Set `initial_wait` to 60-120s for builds/tests
- **Detached mode for servers**: Use `mode="detached"` for long-running servers
- **Clean up after**: Always `pkill -f "process-name"` after testing
- **Check before starting**: Verify port is free before starting server

<h3>Example Cleanup Sequence</h3>

```bash
# Stop any running ceye instances
pkill -f "ceye"

# Check if port 8080 is free
lsof -i :8080

# If occupied, kill it
kill -9 <PID>

# Now start fresh
go build -o bin/ceye ./cmd/ceye
./bin/ceye --demo --port 8080
```

<h2>Key Files and Locations</h2>

<h3>Code</h3>
- `cmd/ceye/main.go` - Entry point
- `internal/core/types.go` - Core types
- `internal/core/store.go` - Run store
- `internal/providers/safe.go` - SafeProvider wrapper
- `internal/server/server.go` - Web server

<h3>Configuration</h3>
- `ceye.yaml` - Main config file
- `config.example.yaml` - Example config

<h3>Documentation</h3>
- `docs/plan.md` - Master development plan ⭐
- `docs/README.md` - Documentation guide
- `docs/references/testing-guide.md` - Testing strategy
- `docs/references/agents.md` - UI testing guide
- `AGENTS.md` - This file

<h3>Tests</h3>
- `internal/core/provider_contract_test.go` - Contract test suite
- `cmd/ceye/integration_test.go` - Integration tests
- `internal/providers/safe_test.go` - Safety tests

<h2>Provider Implementation Guide</h2>

<h3>Implementing the Provider Interface</h3>

```go
package myprovider

import (
    "context"
    "time"
    "github.com/joelklabo/ceye/internal/core"
)

type Provider struct {
    name string
    // Your provider-specific fields
}

func New(config Config) *Provider {
    return &Provider{
        name: "myprovider",
    }
}

func (p *Provider) Name() string {
    return p.name
}

func (p *Provider) Start(ctx context.Context, out chan<- core.RunEvent) error {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            runs, err := p.fetchRuns()
            if err != nil {
                // Send error event
                out <- core.RunEvent{
                    Provider:  p.name,
                    Timestamp: time.Now(),
                    Err:       err,
                    Message:   "Failed to fetch runs",
                }
                continue
            }
            
            // Send runs event
            out <- core.RunEvent{
                Provider:  p.name,
                Runs:      runs,
                Timestamp: time.Now(),
            }
        }
    }
}
```

<h3>Provider Contract Requirements</h3>

Every provider MUST:

1. Return a non-empty name via `Name()`
2. Respect context cancellation in `Start()`
3. Send well-formed events to the channel
4. Be safe for concurrent access
5. Handle multiple `Start()` calls gracefully
6. Not deadlock on event channel operations
7. Maintain name stability (same name every call)
8. Handle context timeout properly

Use the contract test suite to validate:

```go
func TestMyProviderContract(t *testing.T) {
    suite := core.NewProviderContractTestSuite(t, "MyProvider", func() core.Provider {
        return New(Config{})
    })
    
    suite.RunAll()
}
```

<h2>Communication Guidelines</h2>

<h3>What to Include in Responses</h3>

- ✅ Concrete actions taken
- ✅ Test results (pass/fail counts)
- ✅ File paths modified
- ✅ Commands to verify changes
- ✅ Next steps or blockers

<h3>What to Avoid</h3>

- ❌ Asking user to verify things you can check
- ❌ Suggesting changes without making them
- ❌ Incomplete implementations
- ❌ Tests that don't actually test the feature

<h3>When Making Changes</h3>

1. Make the change
2. Write/update tests
3. Run tests locally
4. Verify tests pass
5. Report results with proof

<h2>Important Context</h2>

<h3>Provider = Agent</h3>

In ceye, **providers are the "agents"**. They monitor CI systems and report status via the `Provider` interface.

<h3>SafeProvider Wrapper</h3>

All providers are wrapped in `SafeProvider` which:
- Catches panics and converts to errors
- Validates events before forwarding
- Logs errors with stack traces
- Ensures system stability

<h3>Adaptive Polling</h3>

Providers poll frequently (10-15s) when runs are active and slow down (60-120s) when idle to reduce API load.

<h3>Event-Driven Architecture</h3>

Everything flows through events:
1. Provider fetches data from CI system
2. Provider sends `RunEvent` to channel
3. Store merges events into state
4. UI updates from store

This decouples providers from UI and enables real-time updates.

<h2>Troubleshooting</h2>

<h3>Tests Failing</h3>

```bash
# Run specific test
go test -v ./internal/core/ -run TestStoreMerge

# Run with race detector
go test -race ./...

# Show detailed output
go test -v -count=1 ./...
```

<h3>Provider Not Fetching Data</h3>

- Check authentication (tokens set?)
- Check network connectivity
- Check rate limits
- Look for error events in logs
- Verify configuration is correct

<h3>WebSocket Not Connecting</h3>

**Common Error**: "WebSocket connection failed: Insufficient resources"

**Root Cause**: Infinite WebSocket connection loop in React useEffect

**Critical lesson from Phase 0.7 (commit 2bde007)**:

```typescript
// ❌ BAD - Infinite loop creating hundreds of connections
const connect = useCallback(() => {
  const ws = new WebSocket(url)
  // ... setup
}, [url, onMessage, onError, ...]) // Dependencies cause recreation

useEffect(() => {
  connect()
  return () => ws?.close()
}, [connect]) // ⚠️ connect recreated every render → infinite loop!

// ✅ GOOD - Only connect on mount
useEffect(() => {
  connect()
  return () => ws?.close()
  // eslint-disable-next-line react-hooks/exhaustive-deps
}, []) // Only run once on mount
```

**Why this happens**:
1. useCallback dependencies cause `connect` to be recreated on every render
2. useEffect sees new `connect` function → runs again
3. New WebSocket created → state changes → component re-renders
4. Back to step 1 → infinite loop
5. Hundreds of connections exhaust browser resources ("Insufficient resources")

**How to debug**:
- Check browser console for rapid "WebSocket connection failed" errors
- Check server logs for many connection attempts
- Use React DevTools to see if component re-renders constantly
- Add logging inside useEffect to count how many times it runs

**Prevention**:
- Don't include callback functions in useEffect dependencies if they shouldn't change
- Use empty dependency array `[]` for setup that should only run on mount
- Use `useRef` for functions that need to be stable across renders

**Other WebSocket issues**:
- Check server is running on correct port
- Check firewall rules
- Verify WebSocket endpoint exists
- Check browser console for errors

<h3>JavaScript Errors in Browser</h3>

<h4>DOM Manipulation Race Conditions</h4>

Common error: `Cannot read properties of undefined (reading 'X')`

**Root cause**: Accessing DOM elements during or after modification
**Critical lesson from Phase 2 (commit 6e5f61b)**:

When manipulating DOM with checks/removes/appends, ORDER MATTERS:

```javascript
// ❌ BAD - Race condition
if (log.children.length >= maxItems) {
    log.removeChild(log.firstChild);  // Modifies DOM
}
if (log.children.length === 1 && log.firstChild.classList.contains('muted')) {
    log.innerHTML = '';  // firstChild changed after removal!
}
log.appendChild(newItem);

// ✅ GOOD - Check/clear BEFORE modifying
if (log.children.length === 1) {
    const firstChild = log.firstChild;  // Store reference
    if (firstChild && firstChild.classList && firstChild.classList.contains('muted')) {
        log.innerHTML = '';  // Clear before append
    }
}
log.appendChild(newItem);  // Append
if (log.children.length > maxItems) {  // Check AFTER append
    log.removeChild(log.firstChild);
}
```

**Key principles**:
1. Store DOM references in variables before checks (they may change)
2. Do checks/clears BEFORE adding new elements
3. Do limit checks AFTER adding new elements
4. Always null-check before accessing properties

This is especially important for:
- `classList` operations
- `firstChild`, `lastChild`, `parentNode` access
- Operations that modify children array during iteration
- Custom properties that may not always be present

<h4>React Framer Motion Table Flicker</h4>

Common issue: **Table rows flicker/re-animate on every state update**

**Root cause**: Using `initial` prop on table rows causes re-animation on every render
**Critical lesson from Phase 0.6 (commit ad65646)**:

```typescript
// ❌ BAD - Re-animates on EVERY WebSocket update
{sortedRuns.map((run, index) => (
  <motion.tr
    key={run.ID}
    initial={{ opacity: 0, x: -20 }}  // Triggers animation on EVERY render!
    animate={{ opacity: 1, x: 0 }}
    transition={{ duration: 0.2, delay: index * 0.02 }}
  >
    {/* row content */}
  </motion.tr>
))}

// ✅ GOOD - Only animates on mount, smooth updates
const RunRow = memo(({ run }: RunRowProps) => (
  <motion.tr
    layout  // Smooth position changes
    initial={{ opacity: 0 }}  // Only on mount
    animate={{ opacity: 1 }}
    exit={{ opacity: 0 }}
    transition={{ duration: 0.15, layout: { duration: 0.2 } }}
  >
    {/* row content */}
  </motion.tr>
))

// Wrap in AnimatePresence for enter/exit
<tbody>
  <AnimatePresence mode="popLayout">
    {sortedRuns.map((run) => (
      <RunRow key={run.ID} run={run} />
    ))}
  </AnimatePresence>
</tbody>
```

**Key principles**:
1. **Memoize row components** with `React.memo()` to prevent re-renders
2. **Use `layout` prop** for smooth position transitions (not `initial/animate`)
3. **Use `AnimatePresence`** only at container level, not per-row
4. **Remove stagger delays** (`delay: index * 0.02`) - causes cascading flicker
5. **Only use `initial/animate`** for mount/unmount, not updates

**Performance metrics**:
- Before: Visible flicker on every WebSocket message (every 15s)
- After: Smooth 60fps updates, no flicker

**Testing pattern**:
```typescript
test('should not flicker on updates', async ({ page }) => {
  await page.goto('http://localhost:8080')
  await page.waitForSelector('table tbody tr')
  
  // All rows should have full opacity (not animating)
  const rows = page.locator('table tbody tr')
  for (let i = 0; i < Math.min(await rows.count(), 5); i++) {
    await expect(rows.nth(i)).toHaveCSS('opacity', '1')
  }
  
  // Wait for WebSocket updates
  await page.waitForTimeout(2000)
  
  // Still full opacity (no flicker)
  for (let i = 0; i < Math.min(await rows.count(), 5); i++) {
    await expect(rows.nth(i)).toHaveCSS('opacity', '1')
  }
})
```

**Reference**: See `e2e/flicker-test.spec.ts` for complete test suite

<h4>Playwright WebSocket Test Patterns</h4>

When testing WebSocket messages in Playwright:

```javascript
// ❌ BAD - Misses initial messages
await page.goto(URL);
await page.evaluate(() => {
    // Inject interceptor AFTER page loads - too late!
    window.WebSocket = function() { ... };
});

// ✅ GOOD - Captures all messages
await page.addInitScript(() => {
    // Inject interceptor BEFORE page loads
    window.WebSocket = function() { ... };
});
await page.goto(URL);
```

**Why**: WebSocket connections establish during page load. If you inject the interceptor after `goto()`, you miss the initial snapshot.

<h4>Migrating Playwright Tests from Static HTML to React</h4>

**Problem**: After migrating from static HTML to a React SPA, 25/95 tests fail with "element not found" errors.

**Root Cause**: Tests written for static HTML use specific selectors that don't exist in React components.

**Critical Lesson from Test Migration (commit 0055187)**:

When migrating to React:

1. **Delete obsolete tests** - Tests for features that don't exist (settings page, alerts, workspace selector)
2. **Simplify selectors** - Use text content, not specific HTML structure
3. **Add timeouts** - React needs time to render (2-3s)
4. **Use flexible patterns** - Match multiple states (LIVE|OFFLINE)
5. **Remove old test dependencies** - Delete tests for old architecture (Tailwind direct tests)

```typescript
// ❌ BAD - Tests old HTML structure
test('shows connection status', async ({ page }) => {
  await page.goto('http://localhost:8080')
  await page.waitForSelector('h1:has-text("ceye")')  // Specific structure
  await expect(page.locator('#connection-status')).toBeVisible()  // Specific ID
})

// ✅ GOOD - Tests React app behavior
test('shows connection status', async ({ page }) => {
  await page.goto('http://localhost:8080')
  await page.waitForTimeout(3000)  // Wait for React to render
  
  // Look for either connected or disconnected state
  const status = page.locator('text=/LIVE|OFFLINE/').first()
  await expect(status).toBeVisible({ timeout: 15000 })
})
```

**Migration Strategy**:
1. Run tests and identify failure patterns
2. Check if feature still exists in React app
3. If yes: Update selectors to match React components
4. If no: Delete the test (feature removed)
5. Add React-specific tests (component loading, state updates)

**Key Changes**:
- Static HTML: Immediate rendering → React: Async rendering
- Old: Fixed structure (`#id`, `.class`) → New: Content-based (`text=/pattern/`)
- Old: 5s timeout → New: 15s timeout (React takes longer)
- Old: 95 tests → New: 20 tests (removed obsolete)

**Files to Review**:
- Old tests backed up in `tmp/old-tests/`
- New tests: `e2e/react-app.spec.ts`, `e2e/dashboard-react.spec.ts`, `e2e/connection-indicator.spec.ts`

**Result**: 20/22 tests passing (2 skipped for cross-browser screenshots)

<h3>React + Vite + Go Embedding</h3>

**Problem**: Need to embed a React app built with Vite into a Go binary.

**Solution**: Three-step build process (commit 2d76988)

```bash
# 1. Build React app
cd web && npm run build  # Creates web/dist/

# 2. Copy dist to Go embed location
cp -r web/dist cmd/ceye/web/

# 3. Go embeds from cmd/ceye/web.go
//go:embed web/dist
var webAssets embed.FS
```

**Key Issues Solved**:

1. **go:embed doesn't support `../` paths**
   - ❌ `//go:embed ../../web/dist` - ERROR
   - ✅ Solution: Create cmd/ceye/web/ and copy dist there
   
2. **go:embed doesn't follow symlinks**
   - ❌ `ln -s ../../web web` - ERROR  
   - ✅ Solution: Actual copy in build process

3. **Server needs access to embedded FS**
   - ❌ Can't export embed.FS across packages
   - ✅ Solution: Pass fs.FS to Server.New()

**Makefile Integration**:
```make
web-build:
	cd web && npm run build
	rm -rf cmd/ceye/web
	mkdir -p cmd/ceye/web
	cp -r web/dist cmd/ceye/web/

build: web-build
	go build -o bin/ceye ./cmd/ceye
```

**Development Workflow**:
- Dev: `make web-dev` → Vite dev server (localhost:5173)
- Prod: `make build` → Builds web + embeds + Go binary
- Result: Single binary serves React SPA

**Tailwind v3 vs v4**:
- Shadcn/ui requires Tailwind v3
- If you see v4 installed, downgrade: `npm install -D tailwindcss@^3`

<h2>Quick Reference</h2>

<h3>Most Important Files</h3>

1. **docs/plan.md** - What to work on next
2. **internal/core/types.go** - Core data structures
3. **internal/providers/safe.go** - Provider safety wrapper
4. **cmd/ceye/main.go** - Application entry point

<h3>Most Important Commands</h3>

```bash
# Build and test
go build -o bin/ceye ./cmd/ceye
go test ./...

# Run locally with demo
./bin/ceye --demo --port 8080

# Run dashboard
./bin/ceye --port 8080

# Install globally
sudo cp bin/ceye /usr/local/bin/
```

<h3>Most Important Patterns</h3>

- **TDD**: Write tests first
- **Event-driven**: Communicate via channels
- **Provider interface**: Uniform abstraction
- **SafeProvider**: Wrap everything for safety
- **Contract tests**: Validate implementations

<h2>Links</h2>

- **Master Plan**: [plan.md](plan.md)
- **Project README**: [readme.md](readme.md)
- **Testing Guide**: [references/testing-guide.md](references/testing-guide.md)
- **Webhook Guide**: [references/webhook-guide.md](references/webhook-guide.md)

<h2>Symlink Structure</h2>

**This file (`docs/agents.md`) is the SOURCE OF TRUTH for agent context.**

It is symlinked to multiple locations for different AI tools:

```
/AGENTS.md                           → docs/agents.md (GitHub Copilot)
/CLAUDE.md                           → docs/agents.md (Claude)
/.github/copilot-instructions.md     → docs/agents.md (GitHub Copilot workspace)
```

Similarly, the project README:
```
/README.md                           → docs/readme.md (GitHub display)
```

**Never edit the symlinks directly - always edit `docs/agents.md` or `docs/readme.md`.**

<h2>Documentation Organization</h2>

All documentation lives in the `docs/` folder:

```
docs/
├── plan.md                          # Master development plan
├── README.md                        # Documentation index
└── references/                      # Reference documents
    ├── agents.md                    # This file (symlinked to /AGENTS.md)
    ├── doc-inventory.md             # Documentation structure guide
    ├── readme.md                    # Project README
    ├── testing-guide.md             # Testing standards and practices
    ├── webhook-guide.md             # Webhook implementation guide
    └── web-ui-architecture.md       # Web UI design decisions
```

**Key principles**:
- All markdown files use lowercase-with-hyphens naming
- No markdown files in project root (except AGENTS.md symlink)
- Implementation plans deleted after completion
- Only active/useful reference docs kept
- Consolidate related docs (6 webhook docs → 1 guide)

<h2>Summary</h2>

ceye is a production-ready CI/CD monitoring dashboard with:
- Real-time interface with WebSocket updates
- Multi-provider support (GitHub, Azure, GitLab)
- Real-time event streaming
- Comprehensive testing (140+ tests)
- Clear development plan

**Current priority**: Remove TUI and fix critical bugs (see plan.md)

**Key principle**: Providers are agents. The Provider interface is the agent interface. Everything flows through events.

---

**For AI Agents**: This file provides complete context about the ceye project. Read this first before working on tasks. Always check docs/plan.md for current priorities.
## Learnings from Phase 0.6: Automatic ngrok Setup & Webhook Documentation

### Debugging Go Build Errors: `undefined` for Standard Library Packages

**Problem**: Encountered `undefined: strconv` error during `go test`.
**Root Cause**: The `strconv` package was used in `manager_test.go` but was not explicitly imported in that file. While `manager.go` correctly imported it, the test file needed its own import.
**Lesson**: Even if a package is imported in a related file, each Go file (`.go` or `_test.go`) must explicitly import all packages it uses. The Go compiler reports `undefined` for missing imports, which can be misleading if you expect a "missing import" error. Always double-check imports in the specific file reporting the error.

### Testing External Processes in Go: Challenges and Best Practices

**Problem**: Writing tests for `internal/ngrok/manager.go` which interacts with the external `ngrok` executable.
**Challenges**:
1.  **External Dependency**: `ngrok` must be installed and in the system's PATH for tests to run. Handled by `isNgrokInstalled()` and `t.Skip()`.
2.  **Process Management**: Starting and stopping `ngrok` processes reliably. Used `exec.Command` and `cmd.Process.Kill()`.
3.  **API Interaction**: Querying `ngrok`'s local API (`http://localhost:4040`) to get tunnel details.
4.  **Simulating Failure Modes**: Difficult to simulate scenarios like "ngrok starts but fails to provide a tunnel URL" without complex mocking or environment manipulation.

**Key Learnings**:
-   **Clean State**: Always ensure a clean state before and after tests by killing any lingering external processes (e.g., `killNgrokProcesses()` using `pkill`).
-   **Heuristic Delays**: `time.Sleep()` is often necessary when waiting for external processes to initialize, but it can introduce flakiness. More robust solutions involve polling the external process's API or output.
-   **Test Premise Validation**: Critically evaluate the premise of your tests. `TestManager_Start_NgrokStartsButNoTunnel` was removed because it attempted to test a failure mode that the `Start` method (and its `getTunnelURL` helper) was not designed to detect directly, given `ngrok`'s normal operation. If `ngrok` successfully starts and exposes its local API, `getTunnelURL` will find a tunnel. Simulating a "no tunnel" scenario would require `ngrok` itself to misbehave in a specific way, which is beyond the scope of unit/integration testing for the `Manager`.
-   **Process Verification**: Use `pgrep` or similar tools to verify that external processes are indeed running or stopped as expected.