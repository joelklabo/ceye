# ceye - CI/CD Monitoring Dashboard 🔍

[![CI Status](https://github.com/joelklabo/ceye/workflows/CI/badge.svg)](https://github.com/joelklabo/ceye/actions)
[![Tests](https://github.com/joelklabo/ceye/workflows/Comprehensive%20Tests/badge.svg)](https://github.com/joelklabo/ceye/actions/workflows/tests.yml)
[![Security](https://github.com/joelklabo/ceye/workflows/Security%20Checks/badge.svg)](https://github.com/joelklabo/ceye/actions/workflows/security.yml)
[![Code Quality](https://github.com/joelklabo/ceye/workflows/Code%20Quality/badge.svg)](https://github.com/joelklabo/ceye/actions/workflows/code-quality.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/joelklabo/ceye)](https://goreportcard.com/report/github.com/joelklabo/ceye)
[![codecov](https://codecov.io/gh/joelklabo/ceye/branch/main/graph/badge.svg)](https://codecov.io/gh/joelklabo/ceye)
[![Go Version](https://img.shields.io/github/go-mod/go-version/joelklabo/ceye)](https://github.com/joelklabo/ceye/blob/main/go.mod)
[![Release](https://img.shields.io/github/v/release/joelklabo/ceye)](https://github.com/joelklabo/ceye/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**ceye** (CI Eye) is a production-ready CI/CD monitoring dashboard that aggregates workflow runs from multiple CI providers into a unified, real-time web interface.

**Why ceye?**
- 🎯 **One Dashboard, All Pipelines** - Stop context-switching between GitHub Actions, Azure DevOps, and GitLab
- ⚡ **Real-Time Updates** - See build status changes as they happen
- 🔔 **Smart Alerts** - Get notified when critical builds fail or queues back up
- 📊 **Historical Trends** - Track success rates and build times over weeks
- 🎨 **Beautiful UI** - Modern web dashboard with real-time updates
- 🔌 **Extensible** - Simple Provider interface for adding new CI systems
- 🛡️ **Production Ready** - 175+ tests, panic recovery, graceful degradation

## ✨ Features

### 🎯 Core Capabilities
- **Web UI**: Modern web interface with real-time WebSocket updates
- **Real-time Updates**: Live status changes across all providers
- **Multi-Provider**: GitHub Actions, Azure DevOps, GitLab CI, and Demo mode
- **Historical Data**: SQLite storage with trends and analytics
- **Smart Alerting**: Configurable rules with Slack/webhook notifications
- **Professional UX**: Themes, keyboard shortcuts, workspaces, and more

### 🎨 User Experience
- **Dark Theme**: Clean, professional dark mode interface
- **Advanced Filtering**: Filter by provider, status, and search
- **Real-time Updates**: Instant WebSocket updates as builds complete
- **Responsive Design**: Works on desktop, tablet, and mobile

### 🔔 Alerting
- **4 Alert Conditions**: workflow_failed, high_failure_rate, duration_spike, build_queued_too_long
- **3 Notification Channels**: Slack webhooks, generic webhooks, logging
- **Smart Cooldowns**: Prevent alert spam
- **Alert History**: Track all alerts in web UI
- **Rule Statistics**: Monitor alert fires per hour

## 📸 Screenshots

## 📸 Screenshots

### Web UI

Modern, responsive web interface accessible from any browser:

```
┌─────────────────────────────────────────────────────────────────────────┐
│  ceye - CI/CD Monitoring                    [Dark ▾] [@] [⚙] [?]       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Filters: [github ×] [azure ×] [✓ success ×]  🔍 Search...   [Clear]   │
│                                                                         │
│  Workspace: [All Providers ▾]  [💾 Save]  [📥 Load]                     │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│  Status    Repository          Workflow            Branch    Updated   │
│  ──────────────────────────────────────────────────────────────────    │
│  🟢 Success  myorg/api          Build & Test        main      2m ago   │
│  🔵 Running  myorg/api          Deploy Production   main      1m ago   │
│  🟢 Success  myorg/web          Frontend Build      main      5m ago   │
│  🔴 Failed   MyProject          Integration Test    develop   3m ago   │
│  🟡 Queued   example/service    Build               main      now      │
├─────────────────────────────────────────────────────────────────────────┤
│  📊 Dashboard Stats                                                     │
│  ┌────────────┬────────────┬────────────┬────────────┐                 │
│  │  Total: 45 │ Success: 38│ Running: 2 │ Failed: 5  │                 │
│  │  Success Rate: 84%       │ Avg Duration: 3m 24s   │                 │
│  └────────────┴────────────┴────────────┴────────────┘                 │
│                                                                         │
│  📈 Trends (Last 7 Days)                                                │
│  [Chart showing success rate, build frequency, and duration trends]    │
│                                                                         │
│  🔔 Recent Alerts (2)                                                   │
│  • Production Deploy Failed (2m ago)                                    │
│  • High Failure Rate on myorg/api (5m ago)                              │
└─────────────────────────────────────────────────────────────────────────┘
```

**Web UI Features:**
- 🎨 **4 Beautiful Themes** (Dark, Light, Solarized, Dracula)
- 🔌 **Real-time WebSocket Updates** - No page refresh needed
- 🏷️ **Multi-select Filtering** - Filter by provider, status, repo
- 💾 **Workspaces** - Save and switch between filter presets
- ⌨️ **Keyboard Shortcuts** - `r` refresh, `/` search, `Esc` clear
- ⚙️ **Settings Page** - Centralized configuration
- 📱 **Responsive Design** - Works on mobile, tablet, desktop
- 📊 **Interactive Charts** - Trends with Chart.js
- 🔔 **Alert Notifications** - Desktop and in-app alerts

## 🚀 Quick Start

### Installation

```bash
# Install via Go
go install github.com/joelklabo/ceye/cmd/ceye@latest

# Or clone and build
git clone https://github.com/joelklabo/ceye
cd ceye
go build -o bin/ceye ./cmd/ceye
sudo cp bin/ceye /usr/local/bin/
```

### Try it Out (No Config Required)

```bash
# Web UI with demo data (default mode)
ceye --demo --port 8080

# Then open http://localhost:8080 in your browser
```

## 📖 Usage

### Web UI

The web UI is the default and only mode. It provides a modern, real-time dashboard accessible from any browser.

```bash
# Basic usage
ceye --port 8080

# With specific config
ceye --config /path/to/ceye.yaml --port 8080

# Demo mode (no credentials needed)
ceye --demo --port 8080

# Webhook support (for push-based updates)
ceye --webhooks --webhook-port 9090 --port 8080
```

**Web UI Features:**
- Real-time WebSocket updates
- Provider and status filtering  
- Search across all runs
- Detailed run information
- Provider health monitoring
- Responsive design

### Web UI

```bash
# Start web server on default port 8080
ceye --web

# Custom port
ceye --web --port 9000

# With config file
ceye --web --config ceye.yaml --port 8080

# Demo mode for testing
ceye --web --demo --port 8080

# Bind to specific host (default is localhost)
ceye --web --host 0.0.0.0 --port 8080

# Production mode with historical data
ceye --web --storage /var/lib/ceye/ceye.db
```

**Accessing the Web UI:**

1. Open http://localhost:8080 in your browser
2. Click the theme selector (top right) to choose your theme
3. Use the filter pills to narrow down runs
4. Click "Workspace" to save your current filters
5. Click the settings icon (⚙) to configure preferences

**Web UI Keyboard Shortcuts:**
- `r` - Refresh data
- `/` - Focus search box
- `Esc` - Clear filters
- `a` - Toggle alerts panel
- `d` - Toggle dark mode
- `?` - Show help

**WebSocket API:**

The web UI connects via WebSocket for real-time updates:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === 'runs') {
    // Update runs display
    updateRuns(data.runs);
  } else if (data.type === 'health') {
    // Update provider health
    updateHealth(data.health);
  } else if (data.type === 'alert') {
    // Show alert notification
    showAlert(data.alert);
  }
};
```

### Common Workflows

#### Monitor Production Deployments

```bash
# Create prod.yaml
cat > prod.yaml << EOF
providers:
  - type: github
    display_name: "Production"
    repos:
      - owner: "myorg"
        repo: "api"
      - owner: "myorg"
        repo: "web"

alerting:
  rules:
    - name: "Production Deploy Failed"
      condition: "workflow_failed"
      providers: ["Production"]
      workflows: ["Deploy to Production"]
      severity: "critical"
  
  channels:
    - type: "slack"
      webhook: "${SLACK_WEBHOOK_URL}"
      enabled: true
EOF

# Run with alert monitoring
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."
ceye --config prod.yaml
```

#### Compare Environments

```bash
# Create multi-env.yaml with multiple providers
cat > multi-env.yaml << EOF
providers:
  - type: github
    display_name: "Production"
    repos:
      - owner: "myorg"
        repo: "api"
  
  - type: github
    display_name: "Staging"
    repos:
      - owner: "myorg"
        repo: "api-staging"
  
  - type: github
    display_name: "Development"
    repos:
      - owner: "myorg"
        repo: "api-dev"
EOF

# View all environments
ceye --config multi-env.yaml --web
```

#### Track Historical Trends

```bash
# Enable SQLite storage for trends
ceye --web --storage ceye.db --config ceye.yaml

# After running for a while, you'll see:
# - 7-day trend charts
# - Success rate history
# - Duration trends
# - Failure rate patterns
```

#### Debug CI Issues

```bash
# Enable debug logging and event capture
ceye --log-level debug --log-events debug.jsonl

# In another terminal, watch the events
tail -f debug.jsonl | jq .

# You'll see:
# - Raw events from providers
# - Store merge operations
# - Alert evaluations
# - WebSocket messages
```

#### Multi-Provider Dashboard

```bash
# Monitor GitHub, Azure, and GitLab simultaneously
cat > all-providers.yaml << EOF
providers:
  - type: github
    display_name: "GitHub"
    repos:
      - owner: "myorg"
        repo: "api"
  
  - type: azure
    display_name: "Azure"
    org: "myorg"
    projects:
      - name: "MyProject"
        pipelines: [123, 456]
  
  - type: gitlab
    display_name: "GitLab"
    base_url: "https://gitlab.com"
    projects:
      - "myorg/myproject"
EOF

export GITHUB_TOKEN="ghp_..."
export AZURE_PAT="..."
export GITLAB_TOKEN="..."

ceye --config all-providers.yaml --web --port 8080
```

## ⚙️ Configuration

Create `ceye.yaml`:

```yaml
providers:
  # GitHub Actions
  - type: github
    display_name: "GitHub Production"
    repos:
      - owner: "myorg"
        repo: "api"
      - owner: "myorg"
        repo: "web"
  
  # Azure DevOps
  - type: azure
    display_name: "Azure CI"
    org: "myorg"
    projects:
      - name: "MyProject"
        pipelines: [123, 456]
  
  # GitLab CI
  - type: gitlab
    display_name: "GitLab"
    base_url: "https://gitlab.com"
    projects:
      - "myorg/myproject"
  
  # Demo (for testing)
  - type: demo
    count: 4
    interval: "10s"

# Alerting (optional)
alerting:
  rules:
    - name: "Production Failures"
      condition: "workflow_failed"
      providers: ["github"]
      workflows: ["deploy-prod"]
      severity: "critical"
    
    - name: "High Failure Rate"
      condition: "high_failure_rate"
      threshold: 0.5
      window: "1h"
      severity: "warning"
    
    - name: "Slow Builds"
      condition: "duration_spike"
      threshold: 2.0
      severity: "warning"
    
    - name: "Queue Backup"
      condition: "build_queued_too_long"
      threshold: "10m"
      severity: "warning"
  
  channels:
    - type: "slack"
      webhook: "${SLACK_WEBHOOK_URL}"
      enabled: true
    
    - type: "webhook"
      url: "${CUSTOM_WEBHOOK_URL}"
      enabled: true
    
    - type: "log"
      enabled: true

# Web server (optional)
server:
  port: 8080
  host: "0.0.0.0"
```

### Environment Variables

```bash
# Provider credentials
export GITHUB_TOKEN="ghp_..."
export AZURE_PAT="..."
export GITLAB_TOKEN="..."

# Alert channels
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."
export CUSTOM_WEBHOOK_URL="https://..."

# Config location (optional)
export CEYE_CONFIG="/path/to/ceye.yaml"
```

## 🏗️ Architecture

### Overview

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Providers  │───▶│    Store    │───▶│   Web UI    │
│ (GitHub,    │    │ (Normalized │    │ (WebSocket) │
│  Azure, etc)│    │    Runs)    │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
       │                  │                   │
       │                  ▼                   │
       │           ┌─────────────┐            │
       │           │  Alerting   │            │
       │           │   Engine    │            │
       │           └─────────────┘            │
       │                  │                   │
       ▼                  ▼                   ▼
┌─────────────────────────────────────────────────┐
│              Event Stream (Channels)             │
└─────────────────────────────────────────────────┘
```

### Key Components

- **Providers**: Poll CI systems, emit RunEvents
- **SafeProvider**: Wraps providers with panic recovery
- **Store**: Thread-safe state management
- **Alert Engine**: Rule evaluation and notifications
- **Web Server**: HTTP + WebSocket server for real-time UI
- **Historical Storage**: SQLite for trends

### Data Flow

1. Providers poll CI systems (10-60s intervals)
2. SafeProvider validates and forwards events
3. Store merges events into normalized state
4. Alert engine evaluates rules
5. UIs receive updates via channels/WebSocket
6. Historical data saved to SQLite

## 🔌 Provider Interface (Agent Interface)

**ceye** uses a simple but powerful **Provider Interface** that acts as the "agent" abstraction. Each provider is an independent agent that monitors a CI system and reports status updates.

### The Provider Interface

```go
// Provider is the core abstraction - each provider is an "agent"
// that monitors a CI/CD system and emits events
type Provider interface {
    // Name returns the provider's unique identifier
    Name() string
    
    // Start begins monitoring and sends events to the channel
    // Runs until context is cancelled
    Start(ctx context.Context, out chan<- RunEvent) error
}
```

### Run Event Structure

```go
type RunEvent struct {
    Provider  string                    // Provider name
    Runs      []Run                     // Current runs
    Timestamp time.Time                 // When event was generated
    Err       error                     // Error if fetch failed
    Message   string                    // Human-readable status
    Health    map[string]ProviderHealth // Health information
}

type Run struct {
    ID           string        // Unique identifier
    Provider     string        // Source provider
    Repo         string        // Repository name
    WorkflowName string        // Workflow/pipeline name
    Status       RunStatus     // queued, in_progress, completed
    Conclusion   string        // success, failure, cancelled, etc.
    Branch       string        // Git branch
    CommitSHA    string        // Git commit hash
    StartedAt    time.Time     // When run started
    UpdatedAt    time.Time     // Last update time
    Duration     time.Duration // How long it took/is taking
    URL          string        // Link to run in CI system
}
```

### Implementing a Custom Provider

Here's how to create a custom provider for your CI system:

```go
package myprovider

import (
    "context"
    "time"
    "github.com/joelklabo/ceye/internal/core"
)

type Provider struct {
    name       string
    apiClient  *MyAPIClient
    pollInterval time.Duration
}

func New(config Config) *Provider {
    return &Provider{
        name:         config.Name,
        apiClient:    NewAPIClient(config.Token),
        pollInterval: 30 * time.Second,
    }
}

// Name returns the provider's unique identifier
func (p *Provider) Name() string {
    return p.name
}

// Start begins monitoring and sending events
func (p *Provider) Start(ctx context.Context, out chan<- core.RunEvent) error {
    ticker := time.NewTicker(p.pollInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            // Fetch runs from your CI system
            runs, err := p.fetchRuns(ctx)
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
                Health: map[string]core.ProviderHealth{
                    p.name: {Status: "healthy", LastSuccess: time.Now()},
                },
            }
        }
    }
}

func (p *Provider) fetchRuns(ctx context.Context) ([]core.Run, error) {
    // Call your CI system's API
    data, err := p.apiClient.GetRuns(ctx)
    if err != nil {
        return nil, err
    }
    
    // Convert to ceye's Run format
    runs := make([]core.Run, len(data))
    for i, item := range data {
        runs[i] = core.Run{
            ID:           item.ID,
            Provider:     p.name,
            Repo:         item.Repository,
            WorkflowName: item.Pipeline,
            Status:       mapStatus(item.State),
            Conclusion:   item.Result,
            Branch:       item.Branch,
            CommitSHA:    item.Commit,
            StartedAt:    item.StartTime,
            UpdatedAt:    item.UpdateTime,
            Duration:     item.Duration,
            URL:          item.WebURL,
        }
    }
    
    return runs, nil
}
```

### SafeProvider Wrapper

All providers are automatically wrapped with `SafeProvider` for safety:

```go
// SafeProvider wraps any provider with:
// - Panic recovery
// - Event validation
// - Error logging
// - Health tracking

safeProvider := providers.NewSafeProvider(myProvider)
```

**SafeProvider provides:**
- ✅ **Panic Recovery** - Converts panics to errors
- ✅ **Event Validation** - Ensures data integrity
- ✅ **Error Logging** - Clear error messages with stack traces
- ✅ **Graceful Degradation** - One provider failure doesn't crash the system

### Provider Contract Tests

Validate your provider implementation:

```go
func TestMyProviderContract(t *testing.T) {
    suite := core.NewProviderContractTestSuite(t, "MyProvider", func() core.Provider {
        return New(Config{Name: "test", Token: "test-token"})
    })
    
    // Runs 8 contract validation tests:
    // 1. Provider returns non-empty name
    // 2. Respects context cancellation
    // 3. Sends well-formed events
    // 4. Safe for concurrent access
    // 5. Handles multiple Start() calls
    // 6. Doesn't deadlock on channels
    // 7. Maintains stable name
    // 8. Handles context timeout
    suite.RunAll()
}
```

### Example: GitHub Provider

The GitHub Actions provider shows a complete implementation:

```go
// Polls GitHub Actions API
// Converts workflow runs to normalized Run format
// Handles rate limiting and pagination
// Emits events every 10-60 seconds based on activity
```

See `internal/providers/github/provider.go` for the full implementation.

### Adding Your Provider to ceye

1. **Implement the interface** in `internal/providers/yourprovider/`
2. **Write contract tests** to validate compliance
3. **Add configuration** to `internal/config/config.go`
4. **Register in factory** at `cmd/ceye/provider_cmd.go`
5. **Update docs** with setup instructions

### Built-in Providers

- **GitHub Actions** (`github`) - Full support ✅
- **Azure DevOps** (`azure`) - Full support ✅
- **GitLab CI** (`gitlab`) - Full support ✅
- **Demo** (`demo`) - For testing and demos ✅

### Provider Health Monitoring

Each provider reports health status:

```go
type ProviderHealth struct {
    Status      string    // "healthy", "degraded", "unhealthy"
    LastSuccess time.Time // When last successful fetch occurred
    LastError   error     // Most recent error
    ErrorCount  int       // Consecutive errors
}
```

The dashboard shows health status for all providers in real-time.

## 🧪 Testing

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Specific package
go test ./internal/core/

# Verbose output
go test -v ./...
```

**Test Coverage**: 175+ tests, all passing ✅

## 🛠️ Development

### Prerequisites

- Go 1.21+
- make (optional)

### Build

```bash
# Build
go build -o bin/ceye ./cmd/ceye

# With make
make build

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o bin/ceye-linux-amd64 ./cmd/ceye
```

### Run Tests

```bash
# All tests
make test

# Without make
go test ./...
```

### Project Structure

```
ceye/
├── cmd/
│   └── ceye/           # Main application entry point
├── internal/
│   ├── core/           # Core types and store
│   ├── providers/      # Provider implementations
│   │   ├── github/     # GitHub Actions
│   │   ├── azure/      # Azure DevOps
│   │   ├── gitlab/     # GitLab CI
│   │   └── demo/       # Demo provider
│   ├── alerting/       # Alert engine
│   ├── storage/        # Historical storage (SQLite)
│   ├── server/         # Web server + WebSocket
│   │   └── web/        # Web UI assets
├── docs/               # Documentation
└── .github/
    └── workflows/      # CI/CD workflows
```

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Run tests and linters
5. Submit a pull request

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket support
- [Chart.js](https://www.chartjs.org/) - Web UI charts
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## 📚 Documentation

- [Development Plan](docs/plan.md)
- [Testing Guide](docs/references/testing-guide.md)
- [Webhook Guide](docs/references/webhook-guide.md)
- [Web UI Architecture](docs/references/web-ui-architecture.md)

## 🆘 Support

- [Report Issues](https://github.com/joelklabo/ceye/issues)
- [Discussions](https://github.com/joelklabo/ceye/discussions)

## 🎯 Status

**Production Ready** ✅

- All core features complete
- 175+ tests passing
- Zero known bugs
- Comprehensive documentation
- Ready for deployment

---

Made with ❤️ by [@joelklabo](https://github.com/joelklabo)
