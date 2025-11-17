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
├── cmd/ci-dash/           # Main application entry point
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
go build -o bin/ci-dash ./cmd/ci-dash
sudo cp bin/ci-dash /usr/local/bin/ci-dash
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
ci-dash --port 8080

# Demo mode
ci-dash --demo --demo-duration 5m --port 8080

# With config
ci-dash --config path/to/ceye.yaml --port 8080
```

## UI Testing Strategy

### Dashboard Testing

1. **Start web server**
```bash
./bin/ci-dash --port 8080
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

## Configuration

### Config File Format (ceye.yaml)

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

### Provider Logos

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

### Environment Variables

- `GITHUB_TOKEN` - GitHub personal access token
- `AZURE_PAT` - Azure DevOps personal access token
- `GITLAB_TOKEN` - GitLab personal access token
- `CEYE_CONFIG_ROOT` - Config directory override

## Development Plan

See [docs/plan.md](docs/plan.md) for the comprehensive master plan.

### Current Sprint (Options 2-5)

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

### Future Roadmap

- Enterprise features (auth, RBAC, audit logs)
- Additional providers (Jenkins, CircleCI, Buildkite)
- Mobile app (React Native)
- AI/ML features (failure prediction, anomaly detection)

## Testing Standards

### Required for All Code

- **Unit tests** for all new code
- **Integration tests** for cross-component features
- **Contract tests** for provider implementations
- **E2E tests** for user-facing features

### Test Coverage Goals

- Overall coverage: > 80%
- All critical paths tested
- All providers pass contract tests
- Zero known security issues
- All linters passing

### TDD Workflow

1. Write failing test
2. Implement minimal code to pass
3. Refactor
4. Commit with tests
5. Run full test suite before push

## Success Metrics

### Performance Targets

- Event processing: < 10ms p99
- Store query: < 5ms p99
- WebSocket latency: < 50ms p99
- Memory usage: < 100MB at 1000 runs
- CPU usage: < 5% idle, < 20% under load

### Reliability Targets

- Uptime: 99.9%
- Provider crash recovery: < 5s
- Zero data loss
- Zero UI freezes

### Quality Targets

- Test coverage: > 80%
- All critical paths tested
- Zero known security issues
- Zero memory leaks
- All linters passing

## React Development

### Component Performance

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

### Theme Support

**All components support light/dark modes via CSS variables**:
- Use Tailwind classes: `bg-background`, `text-foreground`
- CSS variables automatically switch with `dark` class on `<html>`
- SVG logos use `currentColor` to adapt to theme

**Adding theme toggle**:
1. ThemeContext provides `theme` and `toggleTheme()`
2. Persists to localStorage
3. Respects system preference on first load
4. Apply `dark` class to `<html>` element

### Provider Icons

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

## Common Tasks

### Adding a New Provider

1. Create provider package in `internal/providers/yourprovider/`
2. Implement `Provider` interface
3. Add configuration schema
4. Write contract compliance tests
5. Add to provider factory
6. Document setup and configuration
7. Update README

### Adding a New Feature

1. Check if it's in [docs/plan.md](docs/plan.md)
2. Create feature branch
3. Write tests first (TDD)
4. Implement feature
5. Update documentation
6. Run full test suite
7. Update plan.md with progress
8. Submit PR

### Fixing a Bug

1. Write test that reproduces bug
2. Fix bug
3. Verify test passes
4. Run full test suite
5. Commit with test
6. No changes to plan.md needed for bugs

## Key Files and Locations

### Code
- `cmd/ci-dash/main.go` - Entry point
- `internal/core/types.go` - Core types
- `internal/core/store.go` - Run store
- `internal/providers/safe.go` - SafeProvider wrapper
- `internal/server/server.go` - Web server

### Configuration
- `ceye.yaml` - Main config file
- `config.example.yaml` - Example config

### Documentation
- `docs/plan.md` - Master development plan ⭐
- `docs/README.md` - Documentation guide
- `docs/references/testing-guide.md` - Testing strategy
- `docs/references/agents.md` - UI testing guide
- `AGENTS.md` - This file

### Tests
- `internal/core/provider_contract_test.go` - Contract test suite
- `cmd/ci-dash/integration_test.go` - Integration tests
- `internal/providers/safe_test.go` - Safety tests

## Provider Implementation Guide

### Implementing the Provider Interface

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

### Provider Contract Requirements

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

## Communication Guidelines

### What to Include in Responses

- ✅ Concrete actions taken
- ✅ Test results (pass/fail counts)
- ✅ File paths modified
- ✅ Commands to verify changes
- ✅ Next steps or blockers

### What to Avoid

- ❌ Asking user to verify things you can check
- ❌ Suggesting changes without making them
- ❌ Incomplete implementations
- ❌ Tests that don't actually test the feature

### When Making Changes

1. Make the change
2. Write/update tests
3. Run tests locally
4. Verify tests pass
5. Report results with proof

## Important Context

### Provider = Agent

In ceye, **providers are the "agents"**. They monitor CI systems and report status via the `Provider` interface.

### SafeProvider Wrapper

All providers are wrapped in `SafeProvider` which:
- Catches panics and converts to errors
- Validates events before forwarding
- Logs errors with stack traces
- Ensures system stability

### Adaptive Polling

Providers poll frequently (10-15s) when runs are active and slow down (60-120s) when idle to reduce API load.

### Event-Driven Architecture

Everything flows through events:
1. Provider fetches data from CI system
2. Provider sends `RunEvent` to channel
3. Store merges events into state
4. UI updates from store

This decouples providers from UI and enables real-time updates.

## Troubleshooting

### Tests Failing

```bash
# Run specific test
go test -v ./internal/core/ -run TestStoreMerge

# Run with race detector
go test -race ./...

# Show detailed output
go test -v -count=1 ./...
```

### Provider Not Fetching Data

- Check authentication (tokens set?)
- Check network connectivity
- Check rate limits
- Look for error events in logs
- Verify configuration is correct

### WebSocket Not Connecting

- Check server is running on correct port
- Check firewall rules
- Verify WebSocket endpoint exists
- Check browser console for errors

### JavaScript Errors in Browser

#### DOM Manipulation Race Conditions

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

#### React Framer Motion Table Flicker

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

#### Playwright WebSocket Test Patterns

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

### React + Vite + Go Embedding

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

## Quick Reference

### Most Important Files

1. **docs/plan.md** - What to work on next
2. **internal/core/types.go** - Core data structures
3. **internal/providers/safe.go** - Provider safety wrapper
4. **cmd/ci-dash/main.go** - Application entry point

### Most Important Commands

```bash
# Build and test
go build -o bin/ci-dash ./cmd/ci-dash
go test ./...

# Run locally with demo
./bin/ci-dash --demo --port 8080

# Run dashboard
./bin/ci-dash --port 8080

# Install globally
sudo cp bin/ci-dash /usr/local/bin/
```

### Most Important Patterns

- **TDD**: Write tests first
- **Event-driven**: Communicate via channels
- **Provider interface**: Uniform abstraction
- **SafeProvider**: Wrap everything for safety
- **Contract tests**: Validate implementations

## Links

- **Master Plan**: [plan.md](plan.md)
- **Project README**: [readme.md](readme.md)
- **Testing Guide**: [references/testing-guide.md](references/testing-guide.md)
- **Webhook Guide**: [references/webhook-guide.md](references/webhook-guide.md)

## Symlink Structure

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

## Documentation Organization

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

## Summary

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
