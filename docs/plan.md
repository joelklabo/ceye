# ceye Development Plan

**Last Updated**: 2025-11-16  
**Status**: Active Development

## Table of Contents

1. [Completed Features](#completed-features)
2. [Current Sprint](#current-sprint)
3. [Option 2: Enhanced Monitoring](#option-2-enhanced-monitoring)
4. [Option 3: Azure DevOps Provider](#option-3-azure-devops-provider)
5. [Option 4: User Experience](#option-4-user-experience)
6. [Option 5: Advanced Testing](#option-5-advanced-testing)
7. [Future Roadmap](#future-roadmap)

---

## Completed Features

### ✅ Phase 1: Core Dashboard (Complete)
- [x] TUI with Bubble Tea
- [x] Provider abstraction (GitHub, Azure, GitLab, Demo)
- [x] Real-time updates via event channels
- [x] Store with normalized Run data
- [x] Provider filtering and status display
- [x] Adaptive polling (fast when active, slow when idle)
- [x] Workspace scanning and config auto-discovery

### ✅ Phase 2: Web UI (Complete)
- [x] HTTP server with WebSocket support
- [x] Static HTML/CSS/JS frontend
- [x] Real-time updates via WebSocket
- [x] All 5 dashboard panels
- [x] Provider filtering and sorting
- [x] Responsive layout
- [x] Comprehensive web server tests (13 tests)

### ✅ Phase 3: Provider Safety (Complete)
- [x] SafeProvider wrapper with panic recovery
- [x] Event validation (4 validation rules)
- [x] Graceful degradation
- [x] Clear error logging with stack traces
- [x] Safety tests (19 tests passing)

### ✅ Phase 4: Integration Testing (Complete)
- [x] Provider → Store → UI flow tests
- [x] Multiple provider coordination tests
- [x] Health panel update tests
- [x] Panic isolation tests
- [x] Integration tests (6 tests passing)

### ✅ Phase 5: Contract Testing (Complete)
- [x] Provider interface contract test suite
- [x] 8 core contract validation tests
- [x] Provider compliance tests
- [x] Contract tests (11 tests passing)

**Total Test Coverage**: 140+ tests, all passing ✅

---

## Current Sprint

### Sprint Status (Updated: 2025-11-16 18:48 UTC)

**Current Work**: Option 3 - Azure DevOps Provider ✅ **WEEKS 1-2 COMPLETE**

**Progress**:
- ✅ Week 1, Phase 1: API Client Completion (COMPLETE)
  - Created comprehensive Azure REST API client (`client.go`, 400 lines)
  - Implemented retry logic, rate limiting, error handling
  - Added status/conclusion mapping for all Azure states
  - Created test suite (`client_test.go`, 360 lines, 20+ tests)
- ✅ Week 1, Phase 2: Response Parsing (COMPLETE)
  - All Azure statuses and results mapped to core types
  - Build parsing with duration calculation
  - Branch name normalization
  - In-progress build handling
- ✅ Week 1, Phase 3: Testing (COMPLETE)
  - Unit tests: 25+ tests covering all mappings and parsing
  - Contract tests: Provider interface compliance
  - Multi-project configuration tests
  - Custom display name tests
- ✅ Week 2, Phase 1: Provider Core (COMPLETE)
  - Enhanced provider with multi-project support
  - Configurable display names
  - Environment variable integration (AZURE_PAT)
  - Adaptive polling with custom intervals
  - Comprehensive documentation (README.md, 10KB)
- ✅ Week 2, Phase 2: Config Integration (COMPLETE - 2025-11-16)
  - Updated factory to support multi-project configuration
  - Added configurable fast/slow polling intervals via config
  - Fixed main.go integration with new client interface
  - All tests passing (140+)
  - Binary builds successfully
- ✅ **Phase 2.1: Historical Data Storage (COMPLETE - 2025-11-16)**
  - SQLite storage backend with 11 storage tests
  - Store integration with async persistence
  - Query interface with 4 query methods
  - 155+ tests passing
- 🎯 Next: Phase 2.2 - Trends and Analytics

### Sprint Goals
Implement enhanced monitoring, complete Azure DevOps provider, improve UX, and add advanced testing capabilities.

### Priority Order
1. ✅ Azure DevOps Provider (foundational) - **COMPLETE** (Weeks 1-2 done, Week 3 deferred)
2. 🎯 **Enhanced Monitoring** (high value) - **STARTING NOW**
3. Advanced Testing (quality assurance)
4. User Experience (polish)

---

## Option 2: Enhanced Monitoring

### Overview
Add historical data tracking, trends analysis, alerting, and performance metrics to provide deeper insights into CI/CD health.

### 2.1 Historical Data Storage (Week 1)

**Goal**: Store run history for trend analysis and metrics

#### Implementation Plan

**Phase 2.1.1: Storage Layer** (2 days)
- [ ] Create `internal/storage` package
- [ ] Design schema for historical runs
  ```go
  type HistoricalRun struct {
      Run         core.Run
      StoredAt    time.Time
      Duration    time.Duration
      WaitTime    time.Duration  // Time from queued to running
  }
  ```
- [ ] Implement SQLite backend (lightweight, embedded)
  - Table: `runs` (id, provider, repo, workflow, status, started_at, completed_at, duration, etc.)
  - Table: `metrics` (provider, repo, timestamp, success_rate, avg_duration, etc.)
- [ ] Add configuration for storage location
- [ ] Create migration system for schema changes

**Phase 2.1.2: Data Persistence** (1 day)
- [ ] Hook into Store.Merge() to persist completed runs
- [ ] Add retention policy (e.g., keep 30/90 days)
- [ ] Implement periodic cleanup job
- [ ] Add storage health checks

**Phase 2.1.3: Query Interface** (1 day)
- [ ] Create query methods:
  - `GetRunHistory(repo, workflow, since)` - Get historical runs
  - `GetProviderMetrics(provider, since)` - Aggregate metrics
  - `GetFailureRate(repo, workflow, period)` - Calculate failure rates
  - `GetAverageDuration(repo, workflow, period)` - Duration trends
- [ ] Add pagination for large result sets
- [ ] Optimize with indexes

**Testing** - ✅ COMPLETE
- [x] Unit tests for storage layer (11 tests)
- [x] Integration tests for persistence (3 tests)
- [x] All 155+ tests passing

### 2.2 Trends and Analytics (Week 2)

**Goal**: Provide insights into CI/CD performance over time

#### Implementation Plan

**Phase 2.2.1: Trend Calculation** (2 days)
- [ ] Implement trend analysis engine
  ```go
  type TrendAnalyzer struct {
      storage Storage
  }
  
  type Trend struct {
      Metric      string          // "success_rate", "duration", "frequency"
      Period      time.Duration   // Last 24h, 7d, 30d
      Current     float64
      Previous    float64
      Change      float64         // Percentage change
      Direction   TrendDirection  // Up, Down, Stable
  }
  ```
- [ ] Calculate metrics:
  - Success rate trends
  - Average duration trends
  - Build frequency trends
  - Failure patterns
- [ ] Add statistical analysis (moving averages, standard deviation)
- [ ] Detect anomalies (sudden spikes, recurring failures)

**Phase 2.2.2: Trends Panel** (2 days)
- [ ] Add trends panel to TUI
  ```
  ┌─ Trends (Last 7 Days) ────────────┐
  │ Success Rate:    95% (↑ 3%)       │
  │ Avg Duration:    4m32s (↓ 15s)    │
  │ Builds/Day:      12 (→ stable)    │
  │ Failure Rate:    5% (↓ 2%)        │
  └────────────────────────────────────┘
  ```
- [ ] Add trends to web UI with charts (Chart.js)
- [ ] Add drill-down capability (click to see details)
- [ ] Add export to CSV/JSON

**Phase 2.2.3: Analytics Dashboard** (2 days)
- [ ] Create `/analytics` endpoint in web server
- [ ] Build analytics page with:
  - Success rate over time (line chart)
  - Duration distribution (histogram)
  - Failure reasons breakdown (pie chart)
  - Busiest times (heatmap)
  - Top failing workflows (bar chart)
- [ ] Add date range selector
- [ ] Add provider/repo filters

**Testing**
- [ ] Unit tests for trend calculations (15+ tests)
- [ ] Edge case tests (no data, single point, etc.)
- [ ] UI tests for trend display

### 2.3 Alerting and Notifications (Week 3)

**Goal**: Notify users of important CI/CD events

#### Implementation Plan

**Phase 2.3.1: Alert Engine** (2 days)
- [ ] Create `internal/alerting` package
- [ ] Define alert rules:
  ```go
  type AlertRule struct {
      Name        string
      Condition   AlertCondition
      Channels    []AlertChannel
      Cooldown    time.Duration
  }
  
  type AlertCondition interface {
      Check(run core.Run, history []HistoricalRun) bool
  }
  ```
- [ ] Implement built-in conditions:
  - Failure after N successes
  - Duration exceeds threshold
  - Success rate drops below X%
  - Build frequency anomaly
  - Specific workflow fails
- [ ] Add alert state management (fired, resolved, suppressed)
- [ ] Implement cooldown to prevent alert spam

**Phase 2.3.2: Notification Channels** (2 days)
- [ ] Implement notification channels:
  - Email (SMTP)
  - Slack webhooks
  - Discord webhooks
  - PagerDuty
  - Generic webhooks
- [ ] Add channel configuration
- [ ] Implement retry logic for failed notifications
- [ ] Add notification templates

**Phase 2.3.3: Alert Configuration** (1 day)
- [ ] Add alerting section to config:
  ```yaml
  alerting:
    rules:
      - name: "Production failures"
        condition:
          type: "workflow_failed"
          workflow: "deploy-prod"
        channels: ["slack-ops", "pagerduty"]
        cooldown: "30m"
    
    channels:
      slack-ops:
        type: "slack"
        webhook: "${SLACK_WEBHOOK}"
      pagerduty:
        type: "pagerduty"
        api_key: "${PD_API_KEY}"
  ```
- [ ] Add alert management UI
- [ ] Add alert history view

**Testing**
- [ ] Unit tests for alert conditions (20+ tests)
- [ ] Integration tests for notifications (10+ tests)
- [ ] Mock notification channels for testing
- [ ] Alert cooldown tests

### 2.4 Performance Metrics Dashboard (Week 4)

**Goal**: Visualize system and CI/CD performance metrics

#### Implementation Plan

**Phase 2.4.1: Metrics Collection** (2 days)
- [ ] Implement metrics collector
  ```go
  type Metrics struct {
      Runs          RunMetrics
      Providers     ProviderMetrics
      System        SystemMetrics
  }
  
  type RunMetrics struct {
      Total         int64
      InProgress    int64
      Queued        int64
      Completed     int64
      Failed        int64
      AvgDuration   time.Duration
  }
  ```
- [ ] Collect provider-specific metrics
- [ ] Collect system metrics (memory, goroutines)
- [ ] Add metrics sampling and aggregation

**Phase 2.4.2: Prometheus Integration** (2 days)
- [ ] Add Prometheus exporter endpoint `/metrics`
- [ ] Export standard metrics:
  - `ceye_runs_total{provider,status}` - Counter
  - `ceye_run_duration_seconds{provider,repo}` - Histogram
  - `ceye_provider_errors_total{provider}` - Counter
  - `ceye_provider_poll_duration_seconds{provider}` - Histogram
  - `ceye_active_providers` - Gauge
- [ ] Add custom metrics support
- [ ] Document Prometheus setup

**Phase 2.4.3: Metrics Dashboard** (1 day)
- [ ] Add `/metrics/dashboard` page to web UI
- [ ] Display real-time metrics
- [ ] Add graphs using Chart.js
- [ ] Add Grafana dashboard template

**Testing**
- [ ] Metrics collection tests (10+ tests)
- [ ] Prometheus format validation tests
- [ ] Metrics accuracy tests

---

## Option 3: Azure DevOps Provider

### Overview
Complete the Azure DevOps provider implementation to match GitHub provider feature parity.

### 3.1 Azure DevOps Client Enhancement (Week 1)

**Goal**: Full-featured Azure DevOps API client

#### Implementation Plan

**Phase 3.1.1: API Client Completion** (2 days)
- [x] Review Azure DevOps REST API documentation
- [x] Implement missing API endpoints:
  - [x] List pipeline runs with filters
  - [x] Get pipeline run details
  - [x] Get build timeline
  - [ ] Get test results (future enhancement)
  - [ ] Get artifact information (future enhancement)
- [x] Add authentication methods:
  - [x] Personal Access Token (PAT)
  - [ ] OAuth (future enhancement)
  - [ ] Service Principal (future enhancement)
- [x] Implement pagination handling (top N results)
- [x] Add retry logic with exponential backoff
- [x] Add rate limiting compliance

**Phase 3.1.2: Response Parsing** (1 day) - COMPLETE ✅
- [x] Map Azure API responses to core.Run:
  - [x] Implemented `parseAzureBuilds()` function
  - [x] ID, provider, repo (project/repository)
  - [x] Workflow name, status, conclusion
  - [x] Branch (with refs/ prefix removal)
  - [x] Commit SHA, timestamps, duration
  - [x] Build URL
- [x] Handle all Azure status values:
  - [x] notStarted, inProgress, completed, cancelling, postponed
- [x] Handle all Azure result values:
  - [x] succeeded, partiallySucceeded, failed, canceled/cancelled
- [x] Add error handling for malformed responses
- [x] Handle in-progress builds (no finish time)
- [ ] Validate with live Azure API (needs credentials)

**Phase 3.1.3: Testing** (1 day) - IN PROGRESS 🚧
- [x] Create comprehensive unit tests:
  - [x] Status mapping tests (8 test cases)
  - [x] Conclusion mapping tests (6 test cases)
  - [x] Branch name cleaning tests (5 test cases)
  - [x] Build parsing tests (2 comprehensive scenarios)
  - [x] In-progress build handling test
- [ ] Create mock Azure API server for integration tests
- [ ] Add authentication flow tests (PAT)
- [ ] Test error handling (4xx, 5xx, rate limits, timeouts)
- [ ] Test retry logic with transient failures
- [ ] Test rate limiting with Retry-After header

### 3.2 Azure Provider Implementation (Week 2)

**Goal**: Production-ready Azure DevOps provider

#### Implementation Plan

**Phase 3.2.1: Provider Core** (2 days) - COMPLETE ✅
- [x] Implement full Azure provider with multi-project support
- [x] Add configurable display names
- [x] Support per-project pipeline filtering
- [x] Implement environment variable integration (AZURE_PAT)
- [x] Add custom polling interval support
- [x] Create factory function `NewProviderFromConfig()`
- [x] Update provider tests for new configuration
      refreshCh    chan struct{}
  }
  
  type ProjectConfig struct {
      Name      string
      Pipelines []int  // Empty = all pipelines
  }
  ```
- [ ] Implement Start() with adaptive polling
- [ ] Handle multiple projects
- [ ] Handle multiple pipelines per project
- [ ] Add refresh trigger support

**Phase 3.2.2: Configuration** (1 day)
- [ ] Add Azure configuration schema:
  ```yaml
  providers:
    - type: azure
      display_name: "Azure Prod"
      org: "myorg"
      projects:
        - name: "MyProject"
          pipelines: [123, 456]  # Empty = all
        - name: "AnotherProject"
          pipelines: []  # All pipelines
  ```
- [ ] Add validation for Azure config
- [ ] Add environment variable support for PAT
- [ ] Document Azure setup

**Phase 3.2.3: Features** (2 days)
- [ ] Add build filtering (branch, tags, status)
- [ ] Add support for build definitions vs. YAML pipelines
- [ ] Add support for classic and multi-stage pipelines
- [ ] Add release pipeline support (if needed)
- [ ] Handle Azure-specific fields (stages, jobs, tasks)

**Testing**
- [ ] Provider contract compliance tests (✓ reuse existing suite)
- [ ] Azure-specific tests (20+ tests)
- [ ] Multi-project tests
- [ ] Multi-pipeline tests
- [ ] Integration tests with mock API

### 3.3 Azure Provider Polish (Week 3)

**Goal**: Feature parity with GitHub provider

#### Implementation Plan

**Phase 3.3.1: Advanced Features** (2 days)
- [ ] Add Azure-specific metadata:
  - Stage information
  - Job details
  - Test results summary
  - Artifact count
- [ ] Add tooltip/detail view with Azure info
- [ ] Add links to Azure DevOps portal
- [ ] Add Azure status badges

**Phase 3.3.2: Performance** (1 day)
- [ ] Optimize API calls (batch requests)
- [ ] Add caching for static data (project/pipeline names)
- [ ] Profile and optimize Azure provider
- [ ] Add performance tests

**Phase 3.3.3: Documentation** (1 day)
- [ ] Write Azure provider documentation
- [ ] Add setup guide (PAT creation, permissions)
- [ ] Add configuration examples
- [ ] Add troubleshooting guide
- [ ] Add comparison with GitHub provider

**Testing**
- [ ] End-to-end tests with real Azure account
- [ ] Performance benchmarks
- [ ] Load tests (many pipelines)

---

## Option 4: User Experience

### Overview
Enhance user experience with keyboard shortcuts, theming, customization, and saved preferences.

### 4.1 Keyboard Shortcuts (Week 1)

**Goal**: Comprehensive keyboard navigation for web UI

#### Implementation Plan

**Phase 4.1.1: Shortcut System** (2 days)
- [ ] Implement JavaScript keyboard handler:
  ```javascript
  const shortcuts = {
      'r': refresh,
      '/': focusSearch,
      'Escape': clearSearch,
      '1-9': switchProvider,
      'g': goToTop,
      'G': goToBottom,
      'j': moveDown,
      'k': moveUp,
      'Enter': openRun,
      'y': copyURL,
      '?': showHelp
  };
  ```
- [ ] Add keyboard event listener
- [ ] Handle key combinations (Ctrl+R, etc.)
- [ ] Prevent conflicts with browser shortcuts
- [ ] Add visual feedback for shortcuts

**Phase 4.1.2: Help Overlay** (1 day)
- [ ] Create help modal with all shortcuts
- [ ] Show on `?` key press
- [ ] Group shortcuts by category
- [ ] Make dismissible (ESC key)
- [ ] Add "Show shortcuts" button

**Phase 4.1.3: TUI Enhancements** (1 day)
- [ ] Review existing TUI shortcuts
- [ ] Add missing shortcuts to match web UI
- [ ] Ensure consistency across UIs
- [ ] Update TUI help text

**Testing**
- [ ] Browser automation tests for shortcuts (Playwright)
- [ ] Test all shortcut combinations
- [ ] Test conflict resolution
- [ ] Accessibility tests

### 4.2 Theme System (Week 2)

**Goal**: Dark/light theme toggle and customization

#### Implementation Plan

**Phase 4.2.1: Theme Engine** (2 days)
- [ ] Define theme structure:
  ```javascript
  const themes = {
      dark: {
          background: '#1a1b26',
          foreground: '#a9b1d6',
          primary: '#7aa2f7',
          success: '#9ece6a',
          warning: '#e0af68',
          error: '#f7768e',
          border: '#414868',
          // ... more colors
      },
      light: {
          background: '#ffffff',
          foreground: '#333333',
          // ...
      }
  };
  ```
- [ ] Implement CSS variables for theming
- [ ] Create theme switcher component
- [ ] Add theme persistence (localStorage)
- [ ] Add smooth transitions

**Phase 4.2.2: Additional Themes** (1 day)
- [ ] Create theme presets:
  - Dark (default)
  - Light
  - Solarized Dark
  - Solarized Light
  - Dracula
  - Nord
- [ ] Add theme preview
- [ ] Add theme import/export

**Phase 4.2.3: TUI Theming** (2 days)
- [ ] Add theme support to TUI
- [ ] Use lipgloss adaptive colors
- [ ] Add `--theme` flag
- [ ] Support terminal color schemes
- [ ] Document terminal color requirements

**Testing**
- [ ] Visual regression tests
- [ ] Theme persistence tests
- [ ] Accessibility tests (contrast ratios)

### 4.3 Dashboard Customization (Week 3)

**Goal**: Customizable dashboard layouts and saved views

#### Implementation Plan

**Phase 4.3.1: Layout Engine** (2 days)
- [ ] Implement flexible layout system:
  ```javascript
  const layouts = {
      default: {
          panels: ['runs', 'active', 'health', 'failures', 'trends'],
          sizes: ['full', 'half', 'half', 'half', 'half']
      },
      compact: {
          panels: ['runs', 'active'],
          sizes: ['full', 'full']
      },
      analytics: {
          panels: ['runs', 'trends', 'analytics', 'health'],
          sizes: ['half', 'half', 'half', 'half']
      }
  };
  ```
- [ ] Add drag-and-drop panel reordering
- [ ] Add panel visibility toggles
- [ ] Add panel size controls
- [ ] Save layout to localStorage

**Phase 4.3.2: Saved Views** (2 days)
- [ ] Implement view persistence:
  ```javascript
  const views = {
      name: 'Production Watch',
      filters: {
          providers: ['github-prod'],
          status: ['failed', 'in_progress'],
          repos: ['api', 'web']
      },
      layout: 'compact',
      sortBy: 'updated_at',
      sortDir: 'desc'
  };
  ```
- [ ] Add view management UI (create, edit, delete)
- [ ] Add quick view switcher
- [ ] Add view sharing (export/import JSON)
- [ ] Add default view setting

**Phase 4.3.3: Preferences** (1 day)
- [ ] Add preferences panel:
  - Auto-refresh interval
  - Notification settings
  - Default view
  - Theme preference
  - Table density (compact/comfortable)
  - Date format
  - Time format
- [ ] Save preferences to localStorage
- [ ] Add preferences import/export

**Testing**
- [ ] Layout persistence tests
- [ ] View CRUD tests
- [ ] Preferences tests
- [ ] Cross-browser compatibility tests

### 4.4 Enhanced Filtering (Week 4)

**Goal**: Advanced filtering and search capabilities

#### Implementation Plan

**Phase 4.4.1: Advanced Filters** (2 days)
- [ ] Implement filter builder:
  - Multiple conditions (AND/OR)
  - Field-specific filters
  - Regex support
  - Date range filters
  - Duration filters
- [ ] Add filter UI component
- [ ] Add filter presets
- [ ] Add filter saving

**Phase 4.4.2: Search Enhancement** (1 day)
- [ ] Implement fuzzy search
- [ ] Add search highlighting
- [ ] Add search history
- [ ] Add search suggestions
- [ ] Search across all fields

**Phase 4.4.3: Quick Filters** (1 day)
- [ ] Add one-click filters:
  - "My repos"
  - "Failed today"
  - "Long running"
  - "Recently completed"
- [ ] Add filter chips with counts
- [ ] Add filter clear all

**Testing**
- [ ] Filter logic tests (30+ tests)
- [ ] Search accuracy tests
- [ ] UI interaction tests

---

## Option 5: Advanced Testing

### Overview
Comprehensive testing strategy including load testing, chaos engineering, E2E tests, and performance benchmarks.

### 5.1 Load Testing (Week 1)

**Goal**: Validate system performance under high load

#### Implementation Plan

**Phase 5.1.1: Load Test Framework** (2 days)
- [ ] Create `internal/loadtest` package
- [ ] Implement load test scenarios:
  ```go
  type LoadTestScenario struct {
      Name          string
      Providers     int        // Number of concurrent providers
      EventsPerSec  int        // Events generated per second
      RunCount      int        // Total runs in system
      Duration      time.Duration
      Assertions    []Assertion
  }
  ```
- [ ] Add test scenarios:
  - 10 providers, 100 runs
  - 50 providers, 1000 runs
  - 100 providers, 10000 runs
  - Sustained load (1 hour)
  - Spike load (sudden burst)

**Phase 5.1.2: Metrics Collection** (1 day)
- [ ] Collect performance metrics during load tests:
  - Memory usage
  - CPU usage
  - Goroutine count
  - Event processing latency
  - Store query performance
  - WebSocket message latency
- [ ] Add metrics visualization
- [ ] Add performance regression detection

**Phase 5.1.3: Load Test Suite** (2 days)
- [ ] Create load test runners
- [ ] Add automated load tests to CI
- [ ] Set performance budgets
- [ ] Add performance reports
- [ ] Document load test results

**Testing**
- [ ] Baseline performance tests
- [ ] Memory leak tests
- [ ] Resource exhaustion tests
- [ ] Recovery tests

### 5.2 Chaos Engineering (Week 2)

**Goal**: Validate system resilience under failure conditions

#### Implementation Plan

**Phase 5.2.1: Chaos Test Framework** (2 days)
- [ ] Create `internal/chaos` package
- [ ] Implement chaos scenarios:
  ```go
  type ChaosScenario interface {
      Name() string
      Execute(ctx context.Context, sys *System) error
      Verify() error
  }
  ```
- [ ] Add chaos scenarios:
  - Random provider crashes
  - Network failures
  - Slow responses
  - Corrupted events
  - Store failures
  - Memory pressure
  - CPU saturation

**Phase 5.2.2: Fault Injection** (2 days)
- [ ] Implement fault injection:
  - Provider panics
  - API timeouts
  - Network partitions
  - Disk full
  - Rate limit exceeded
- [ ] Add failure probability controls
- [ ] Add failure recovery verification
- [ ] Add blast radius measurement

**Phase 5.2.3: Chaos Tests** (1 day)
- [ ] Create chaos test suite
- [ ] Add automated chaos tests
- [ ] Verify graceful degradation
- [ ] Verify error recovery
- [ ] Document chaos test results

**Testing**
- [ ] System resilience tests (20+ scenarios)
- [ ] Recovery time tests
- [ ] Partial failure tests
- [ ] Cascading failure tests

### 5.3 End-to-End Testing (Week 3)

**Goal**: Browser automation tests for web UI

#### Implementation Plan

**Phase 5.3.1: E2E Test Setup** (1 day)
- [ ] Set up Playwright
- [ ] Create test fixtures
- [ ] Add test helpers:
  ```javascript
  class DashboardPage {
      async goto() { /* ... */ }
      async waitForRuns() { /* ... */ }
      async filterByProvider(name) { /* ... */ }
      async selectRun(id) { /* ... */ }
      async verifyRunCount(expected) { /* ... */ }
  }
  ```
- [ ] Configure browsers (Chrome, Firefox, Safari)
- [ ] Set up CI integration

**Phase 5.3.2: Critical Path Tests** (2 days)
- [ ] Test user flows:
  - Dashboard loads and displays runs
  - WebSocket connection establishes
  - Real-time updates appear
  - Provider filtering works
  - Status filtering works
  - Search works
  - Run details modal opens
  - Copy URL works
  - Refresh works
- [ ] Test edge cases:
  - No runs
  - Connection loss
  - Slow network
  - Many runs (scrolling)

**Phase 5.3.3: Visual Testing** (2 days)
- [ ] Add screenshot comparison tests
- [ ] Test responsive layouts
- [ ] Test theme switching
- [ ] Test keyboard navigation
- [ ] Test accessibility

**Testing**
- [ ] E2E tests (30+ scenarios)
- [ ] Visual regression tests
- [ ] Cross-browser tests
- [ ] Mobile browser tests

### 5.4 Performance Benchmarks (Week 4)

**Goal**: Establish performance baselines and detect regressions

#### Implementation Plan

**Phase 5.4.1: Benchmark Suite** (2 days)
- [ ] Create benchmark tests:
  ```go
  func BenchmarkStoreInsert(b *testing.B) { /* ... */ }
  func BenchmarkStoreQuery(b *testing.B) { /* ... */ }
  func BenchmarkEventProcessing(b *testing.B) { /* ... */ }
  func BenchmarkWebSocketBroadcast(b *testing.B) { /* ... */ }
  func BenchmarkProviderPolling(b *testing.B) { /* ... */ }
  ```
- [ ] Benchmark critical paths
- [ ] Add memory allocation benchmarks
- [ ] Add concurrent operation benchmarks

**Phase 5.4.2: Performance Tracking** (1 day)
- [ ] Set up benchstat for comparison
- [ ] Add benchmark results to CI
- [ ] Track performance over time
- [ ] Alert on regressions > 10%
- [ ] Generate performance reports

**Phase 5.4.3: Profiling** (2 days)
- [ ] Add CPU profiling
- [ ] Add memory profiling
- [ ] Add goroutine profiling
- [ ] Add block profiling
- [ ] Create profiling guides

**Testing**
- [ ] Benchmark stability tests
- [ ] Profile correctness verification
- [ ] Regression detection tests

---

## Future Roadmap

### Phase 6: Enterprise Features (Future)
- [ ] Multi-user support with authentication
- [ ] Role-based access control (RBAC)
- [ ] Team/organization management
- [ ] Audit logging
- [ ] SSO integration (SAML, OAuth)

### Phase 7: Additional Providers (Future)
- [ ] Jenkins provider
- [ ] CircleCI provider
- [ ] Buildkite provider
- [ ] Travis CI provider
- [ ] Drone CI provider
- [ ] GitLab CI (enhancement)

### Phase 8: Mobile (Future)
- [ ] React Native app
- [ ] Push notifications
- [ ] Mobile-optimized UI
- [ ] Offline support

### Phase 9: AI/ML Features (Future)
- [ ] Failure prediction
- [ ] Anomaly detection
- [ ] Build time optimization suggestions
- [ ] Root cause analysis

---

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

---

## Development Process

### Workflow
1. Create feature branch from `main`
2. Implement with TDD (tests first)
3. Run full test suite locally
4. Create PR with description
5. CI validates all tests + linters
6. Code review
7. Merge to `main`
8. Deploy and verify

### Testing Standards
- Unit tests for all new code
- Integration tests for cross-component features
- E2E tests for user-facing features
- Performance tests for critical paths
- All tests must pass before merge

### Documentation Standards
- Update plan.md with progress
- Document new features in references/
- Update README with user-facing changes
- Add inline code comments for complex logic
- Keep architecture docs up to date

---

## Notes

**File Organization**
- `docs/plan.md` - This file (implementation plan)
- `docs/plans/` - Detailed implementation plans (archived after completion)
- `docs/references/` - Reference documentation and guides

**Naming Conventions**
- All files use lowercase with hyphens: `my-feature.md`
- Directories use lowercase: `plans/`, `references/`
- Code follows Go conventions

**Priority System**
- P0: Critical (blocks other work)
- P1: High (current sprint)
- P2: Medium (next sprint)
- P3: Low (future)

**Status Indicators**
- [ ] Not started
- [x] Complete
- [~] In progress
- [!] Blocked
