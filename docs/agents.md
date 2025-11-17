# ceye - CI/CD Monitoring Dashboard - Agent Context

**Last Updated**: 2025-11-16  
**Version**: 1.0  
**Status**: Active Development

## Project Overview

**ceye** (CI Eye) is a terminal and web-based CI/CD monitoring dashboard that aggregates workflow runs from multiple CI providers (GitHub Actions, Azure DevOps, GitLab CI) into a unified, real-time view.

### Key Features

- **Dual UI**: Terminal UI (Bubble Tea) and Web UI (WebSocket)
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
│   ├── server/            # HTTP/WebSocket server for web UI
│   └── ui/                # Terminal UI (Bubble Tea)
├── web/                   # Static web UI assets
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
Provider → RunEvent → Store → UI (TUI/Web)
   ↓
SafeProvider (panic recovery, validation)
   ↓
Manager (lifecycle, health tracking)
```

## ⚠️ CRITICAL: Temporary Files

**ALWAYS use `tmp/` directory for temporary files**

```
tmp/                    # Gitignored, safe for any temporary work
├── .gitkeep           # Explains the folder purpose
├── *.log              # Log files from testing
├── *.md               # Working notes, test results
├── *.txt              # Debug output, snapshots
├── *.pid              # Process IDs
└── ...                # Any temporary file

DO:
✅ Write all logs to tmp/
✅ Put test outputs in tmp/
✅ Use tmp/ for debugging files
✅ Store session notes in tmp/
✅ Save temporary scripts in tmp/

DO NOT:
❌ Write to /tmp/ (permission issues)
❌ Create temp files in project root
❌ Put temp files in docs/
❌ Commit anything from tmp/ to git
```

**Why**: Avoids permission errors, keeps project organized, all temp files in one place.

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
- TUI with Bubble Tea
- Provider abstraction layer
- Real-time event streaming
- Store with normalized Run data
- Adaptive polling (fast when active, slow when idle)
- Workspace scanning and config auto-discovery

**Phase 2: Web UI**
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
# TUI mode (default)
ci-dash

# Web mode
ci-dash --web --port 8080

# Demo mode
ci-dash --demo --demo-duration 5m

# With config
ci-dash --config path/to/ceye.yaml
```

## UI Testing Strategy

### Terminal UI Testing

**Always verify TUI changes yourself before completing the task.**

1. **Build the binary**
```bash
cd /Users/honk/code/ceye
go build -o bin/ci-dash ./cmd/ci-dash
```

2. **Start in tmux** (allows capture without user interaction)
```bash
tmux kill-session -t ci-dash-live 2>/dev/null
tmux new-session -d -s ci-dash-live "cd /Users/honk/code/ceye && ./bin/ci-dash"
sleep 3
```

3. **Capture and verify output**
```bash
tmux capture-pane -t ci-dash-live -p | head -50
```

4. **Test with demo mode**
```bash
./bin/ci-dash --demo --demo-duration 5s 2>&1 | cat
```

### Common UI Issues to Check

- Text overflow and truncation
- Column alignment in tables
- Status indicators (✓ ✗ ▸ …)
- Width calculation with ANSI codes
- Terminal size compatibility (80x24 and wider)
- Border rendering and panel alignment

### Web UI Testing

1. **Start web server**
```bash
./bin/ci-dash --web --port 8080
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
    repos:
      - owner: "myorg"
        repo: "myrepo"
  
  - type: azure
    display_name: "Azure DevOps"
    org: "myorg"
    projects:
      - name: "MyProject"
        pipelines: [123, 456]
  
  - type: demo
    count: 5
    interval: "10s"

server:
  port: 8080
  host: "0.0.0.0"
```

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
- `internal/ui/model.go` - TUI model

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

### TUI Not Rendering Correctly

```bash
# Check in tmux
tmux new-session -d -s test "./bin/ci-dash --demo"
sleep 2
tmux capture-pane -t test -p

# Check terminal size
echo $COLUMNS $LINES

# Test with different sizes
COLUMNS=80 LINES=24 ./bin/ci-dash --demo --demo-duration 5s
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

# Run locally
./bin/ci-dash --demo

# Run web UI
./bin/ci-dash --web --port 8080

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
- Dual UI (terminal + web)
- Multi-provider support
- Real-time updates
- Comprehensive testing (140+ tests)
- Clear development plan (15+ weeks mapped out)

**Current priority**: Complete Azure DevOps provider (Option 3 in plan.md)

**Key principle**: Providers are agents. The Provider interface is the agent interface. Everything flows through events.

---

**For AI Agents**: This file provides complete context about the ceye project. Read this first before working on tasks. Always check docs/plan.md for current priorities.
