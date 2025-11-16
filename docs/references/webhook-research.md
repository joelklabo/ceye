# Webhook-Based CI/CD Monitoring Research

**Date**: 2025-11-16  
**Purpose**: Replace polling with webhooks to reduce API rate limits  
**Status**: Research Complete

## Executive Summary

Both GitHub and Azure DevOps support **webhooks** (push notifications) as an alternative to polling. This approach:

- ✅ **Eliminates rate limits** - No repeated API calls needed
- ✅ **Real-time updates** - Instant notification when workflows complete
- ✅ **Scalable** - Works for any number of repositories
- ✅ **Efficient** - Only sends data when events occur

**Recommendation**: Implement webhook receiver in ceye to replace current polling mechanism.

---

## GitHub Webhooks

### Overview

GitHub sends HTTP POST requests to your endpoint when specific events occur (e.g., workflow runs, check runs, status changes).

### Key Events for CI Monitoring

| Event | Description | Use Case |
|-------|-------------|----------|
| `workflow_run` | GitHub Actions workflow triggered/completed | Primary CI status monitoring |
| `check_run` | Check run status updated | Individual job monitoring |
| `status` | Commit status changed | Generic CI system monitoring |

### Webhook Payload Structure

**Example `workflow_run` event** (completed):

```json
{
  "action": "completed",
  "workflow_run": {
    "id": 123456789,
    "name": "CI",
    "node_id": "MDExOldvcmtmbG93UnVuMTIzNDU2Nzg5",
    "head_branch": "main",
    "head_sha": "d6fde92930d4715a2b49857d24b940956b26d2d3",
    "path": ".github/workflows/ci.yml",
    "run_number": 42,
    "event": "push",
    "status": "completed",
    "conclusion": "success",
    "workflow_id": 1234567,
    "created_at": "2021-08-30T16:07:58Z",
    "updated_at": "2021-08-30T16:09:09Z",
    "run_started_at": "2021-08-30T16:08:21Z",
    "jobs_url": "https://api.github.com/repos/owner/repo/actions/runs/123456789/jobs",
    "logs_url": "https://api.github.com/repos/owner/repo/actions/runs/123456789/logs",
    "artifacts_url": "https://api.github.com/repos/owner/repo/actions/runs/123456789/artifacts",
    "repository": {
      "id": 1234567,
      "name": "repo",
      "full_name": "owner/repo",
      "owner": { "login": "owner" }
    },
    "head_commit": {
      "id": "d6fde92930d4715a2b49857d24b940956b26d2d3",
      "message": "Fix bug"
    },
    "triggering_actor": {
      "login": "username"
    }
  },
  "repository": {
    "id": 1234567,
    "name": "repo",
    "full_name": "owner/repo"
  },
  "organization": {
    "login": "org"
  },
  "sender": {
    "login": "username"
  }
}
```

**Key fields for ceye**:
- `workflow_run.id` → `Run.ID`
- `workflow_run.name` → `Run.WorkflowName`
- `workflow_run.status` → `Run.Status`
- `workflow_run.conclusion` → `Run.Conclusion`
- `workflow_run.head_branch` → `Run.Branch`
- `workflow_run.head_sha` → `Run.CommitSHA`
- `workflow_run.created_at` → `Run.StartedAt`
- `workflow_run.updated_at` → `Run.UpdatedAt`
- Duration = `updated_at - run_started_at`
- `repository.full_name` → `Run.Repo`

### Setting Up Webhooks

#### Via GitHub UI
1. Go to repository → Settings → Webhooks
2. Click "Add webhook"
3. Set **Payload URL**: `https://your-domain.com/webhooks/github`
4. Set **Content type**: `application/json`
5. Set **Secret**: Generate secure random string
6. Select events: `workflow_run`, `check_run`, `status`
7. Save

#### Via GitHub CLI

```bash
gh api repos/<org>/<repo>/hooks \
  --method POST \
  --field name=web \
  --field active=true \
  --field events[]=workflow_run \
  --field events[]=check_run \
  --field events[]=status \
  --field config='{"url":"https://your-domain.com/webhooks/github","content_type":"json","secret":"YOUR_SECRET"}'
```

#### Via GitHub REST API

```bash
curl -X POST \
  -H "Authorization: token YOUR_GITHUB_TOKEN" \
  -H "Content-Type: application/json" \
  https://api.github.com/repos/OWNER/REPO/hooks \
  -d '{
    "name": "web",
    "active": true,
    "events": ["workflow_run", "check_run", "status"],
    "config": {
      "url": "https://your-domain.com/webhooks/github",
      "content_type": "json",
      "secret": "YOUR_SECRET"
    }
  }'
```

### Security: HMAC Signature Verification

GitHub signs webhook payloads with HMAC-SHA256.

**Go Example**:

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "io"
    "net/http"
    "strings"
)

func ValidateGitHubWebhook(r *http.Request, secret []byte) ([]byte, bool) {
    sigHeader := r.Header.Get("X-Hub-Signature-256")
    if !strings.HasPrefix(sigHeader, "sha256=") {
        return nil, false
    }
    sig := sigHeader[len("sha256="):]
    
    payload, err := io.ReadAll(r.Body)
    if err != nil {
        return nil, false
    }
    
    mac := hmac.New(sha256.New, secret)
    mac.Write(payload)
    expectedMAC := mac.Sum(nil)
    expectedSig := hex.EncodeToString(expectedMAC)
    
    // Constant-time comparison to prevent timing attacks
    return payload, hmac.Equal([]byte(sig), []byte(expectedSig))
}
```

**Usage in webhook handler**:

```go
func HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
    secret := []byte(os.Getenv("GITHUB_WEBHOOK_SECRET"))
    payload, valid := ValidateGitHubWebhook(r, secret)
    if !valid {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Parse payload
    var event WorkflowRunEvent
    if err := json.Unmarshal(payload, &event); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }
    
    // Process event...
    w.WriteHeader(http.StatusOK)
}
```

### Rate Limits Comparison

| Method | Rate Limit | Latency | Scalability |
|--------|------------|---------|-------------|
| **Polling** | 5,000 requests/hour (authenticated) | 10-60s | Limited by API quota |
| **Webhooks** | No limit | < 1s | Unlimited repositories |

---

## Azure DevOps Service Hooks

### Overview

Azure DevOps "Service Hooks" send HTTP POST requests to external endpoints when events occur (builds, releases, work items, etc.).

### Key Events for CI Monitoring

| Event | Description | Use Case |
|-------|-------------|----------|
| `build.complete` | Build finished | Primary CI status monitoring |
| `build.started` | Build started | Track in-progress builds |
| `release.deployment.completed` | Release deployed | Deployment monitoring |

### Webhook Payload Structure

**Example `build.complete` event**:

```json
{
  "subscriptionId": "aaaa0a0a-bb1b-cc2c-dd3d-eeeeee4e4e4e",
  "notificationId": 1,
  "id": "a0a0a0a0-bbbb-cccc-dddd-e1e1e1e1e1e1",
  "eventType": "build.complete",
  "publisherId": "tfs",
  "message": {
    "text": "Build 20241202.1 succeeded",
    "html": "Build <a href=\"...\">20241202.1</a> succeeded",
    "markdown": "Build [20241202.1](...) succeeded"
  },
  "detailedMessage": {
    "text": "Build completed successfully.",
    "html": "Build completed <strong>successfully</strong>.",
    "markdown": "Build completed **successfully**."
  },
  "resource": {
    "id": 2727068,
    "buildNumber": "20241202.1",
    "status": "succeeded",
    "result": "succeeded",
    "definition": {
      "id": 69,
      "name": "Sample-CI"
    },
    "project": {
      "id": "e4e4e4e4-ffff-aaaa-bbbb-c5c5c5c5c5c5",
      "name": "FabrikamFiber"
    },
    "uri": "vstfs:///Build/Build/2727068",
    "url": "https://dev.azure.com/FabrikamFiber/_apis/build/builds/2727068",
    "startTime": "2024-12-02T21:17:55.123Z",
    "finishTime": "2024-12-02T21:19:21.456Z",
    "requests": [],
    "triggerInfo": {}
  },
  "resourceVersion": "1.0",
  "resourceContainers": {
    "collection": { "id": "f1f1f1f1-aaaa-bbbb-cccc-eeeeeeeeeeee" },
    "account": { "id": "00000000-0000-0000-0000-000000000000" },
    "project": { "id": "e4e4e4e4-ffff-aaaa-bbbb-c5c5c5c5c5c5" }
  },
  "createdDate": "2024-12-02T21:19:21.987Z"
}
```

**Key fields for ceye**:
- `resource.id` → `Run.ID`
- `resource.definition.name` → `Run.WorkflowName`
- `resource.status` + `resource.result` → `Run.Status` + `Run.Conclusion`
- `resource.project.name` → Part of `Run.Repo`
- `resource.buildNumber` → Display name
- `resource.startTime` → `Run.StartedAt`
- `resource.finishTime` → `Run.UpdatedAt`
- Duration = `finishTime - startTime`
- `resource.url` → `Run.URL`

### Setting Up Service Hooks

#### Via Azure DevOps UI
1. Go to Project Settings → Service Hooks
2. Click "+ Create subscription"
3. Choose **Web Hooks** service
4. Select trigger event (e.g., `Build completed`)
5. Filter by pipeline/status if needed
6. Configure action:
   - **URL**: `https://your-domain.com/webhooks/azure`
   - **Resource details**: Select "All" for complete payload
   - **Messages to send**: JSON
   - **Basic authentication**: Optional username/password
7. Test and finish

#### Via Azure CLI (REST API)

Azure CLI doesn't have direct service hook commands, but you can use the REST API:

```bash
az rest --method POST \
  --uri "https://dev.azure.com/{organization}/{project}/_apis/hooks/subscriptions?api-version=6.0" \
  --headers "Content-Type=application/json" \
  --body '{
    "publisherId": "tfs",
    "eventType": "build.complete",
    "resourceVersion": "1.0",
    "consumerId": "webHooks",
    "consumerActionId": "httpRequest",
    "publisherInputs": {
      "definitionName": "YOUR_PIPELINE_NAME",
      "projectId": "YOUR_PROJECT_GUID"
    },
    "consumerInputs": {
      "url": "https://your-domain.com/webhooks/azure",
      "basicAuthUsername": "",
      "basicAuthPassword": ""
    }
  }'
```

### Security: Basic Auth + HTTPS

Azure DevOps webhooks support:
- **Basic Authentication**: Username/password in webhook config
- **HTTPS**: Always use HTTPS endpoints
- **Custom Headers**: Can add authentication headers

**Go Example**:

```go
func HandleAzureWebhook(w http.ResponseWriter, r *http.Request) {
    // Verify basic auth if configured
    username, password, ok := r.BasicAuth()
    expectedUser := os.Getenv("AZURE_WEBHOOK_USER")
    expectedPass := os.Getenv("AZURE_WEBHOOK_PASS")
    
    if !ok || username != expectedUser || password != expectedPass {
        w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Parse payload
    var event AzureBuildCompleteEvent
    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }
    
    // Process event...
    w.WriteHeader(http.StatusOK)
}
```

---

## Local Development & Testing

### Challenge
Webhooks require a public HTTPS endpoint, but during development you're running on `localhost`.

### Solutions

#### 1. ngrok (Recommended)

**Setup**:
```bash
# Install
brew install ngrok

# Authenticate (optional, for persistent URLs)
ngrok authtoken YOUR_AUTH_TOKEN

# Expose local port
ngrok http 8080
```

**Output**:
```
Forwarding  https://abc123.ngrok.io -> http://localhost:8080
```

**Features**:
- ✅ HTTPS by default
- ✅ Inspect/replay requests at `http://localhost:4040`
- ✅ Static domains (paid plan)
- ✅ Request authentication support

**Configuration in webhooks**:
- Set webhook URL to `https://abc123.ngrok.io/webhooks/github`
- Update URL when ngrok restarts (unless using static domain)

#### 2. smee.io (GitHub-Specific)

**Setup**:
```bash
# Create channel at https://smee.io/
# Install CLI
npm install -g smee-client

# Forward to localhost
smee -u https://smee.io/abc123 -t http://localhost:8080/webhooks/github
```

**Features**:
- ✅ Purpose-built for GitHub webhooks
- ✅ Web UI to inspect payloads
- ✅ No installation needed (browser-based)
- ❌ Limited to GitHub

#### 3. LocalTunnel

**Setup**:
```bash
npx localtunnel --port 8080
```

**Features**:
- ✅ No signup required
- ✅ Open source
- ❌ URLs expire quickly
- ❌ Less reliable

### Development Workflow

1. **Start local ceye server** on port 8080 with webhook endpoint
2. **Start ngrok**: `ngrok http 8080`
3. **Configure webhook** with ngrok URL: `https://abc123.ngrok.io/webhooks/github`
4. **Trigger events** in GitHub/Azure DevOps
5. **Inspect payloads** in ngrok dashboard: `http://localhost:4040`
6. **Debug/replay** failed requests from ngrok UI

---

## Implementation Plan for ceye

### Phase 1: Webhook Receiver (Week 1)

**Goal**: Add HTTP server to receive and validate webhooks

#### 1.1 Create Webhook Server (2 days)

**New files**:
- `internal/webhooks/server.go` - HTTP server with webhook endpoints
- `internal/webhooks/github.go` - GitHub webhook handler
- `internal/webhooks/azure.go` - Azure webhook handler
- `internal/webhooks/validation.go` - HMAC signature validation

**Features**:
- HTTP server on configurable port (default 9090)
- Routes:
  - `POST /webhooks/github` - GitHub webhook receiver
  - `POST /webhooks/azure` - Azure webhook receiver
  - `GET /webhooks/health` - Health check
- HMAC-SHA256 signature validation (GitHub)
- Basic auth validation (Azure)
- Request logging and error handling

#### 1.2 Event Parsing (1 day)

**Goal**: Convert webhook payloads to `core.RunEvent`

```go
// internal/webhooks/github.go
func ParseGitHubWorkflowRun(payload []byte) (core.RunEvent, error) {
    var event GitHubWorkflowRunEvent
    if err := json.Unmarshal(payload, &event); err != nil {
        return core.RunEvent{}, err
    }
    
    run := core.Run{
        ID:           fmt.Sprintf("%d", event.WorkflowRun.ID),
        Provider:     "github",
        Repo:         event.Repository.FullName,
        WorkflowName: event.WorkflowRun.Name,
        Status:       mapGitHubStatus(event.WorkflowRun.Status),
        Conclusion:   event.WorkflowRun.Conclusion,
        Branch:       event.WorkflowRun.HeadBranch,
        CommitSHA:    event.WorkflowRun.HeadSHA,
        StartedAt:    event.WorkflowRun.RunStartedAt,
        UpdatedAt:    event.WorkflowRun.UpdatedAt,
        Duration:     event.WorkflowRun.UpdatedAt.Sub(event.WorkflowRun.RunStartedAt),
        URL:          fmt.Sprintf("https://github.com/%s/actions/runs/%d", 
                                  event.Repository.FullName, event.WorkflowRun.ID),
    }
    
    return core.RunEvent{
        Provider:  "github",
        Runs:      []core.Run{run},
        Timestamp: time.Now(),
    }, nil
}
```

#### 1.3 Integration with Store (1 day)

**Goal**: Forward webhook events to existing store

```go
// cmd/ci-dash/main.go
func main() {
    // ... existing setup ...
    
    // Create webhook server
    webhookServer := webhooks.NewServer(webhooks.Config{
        Port:          9090,
        GitHubSecret:  os.Getenv("GITHUB_WEBHOOK_SECRET"),
        AzureUser:     os.Getenv("AZURE_WEBHOOK_USER"),
        AzurePassword: os.Getenv("AZURE_WEBHOOK_PASS"),
    })
    
    // Forward webhook events to store
    go func() {
        for event := range webhookServer.Events() {
            store.Merge(event)
        }
    }()
    
    // Start webhook server
    go webhookServer.Start(ctx)
    
    // ... rest of app ...
}
```

#### 1.4 Testing (1 day)

- Unit tests for payload parsing
- Integration tests with mock webhooks
- Test with ngrok + real GitHub/Azure DevOps
- Verify signature validation works
- Test error handling

### Phase 2: Hybrid Mode (Week 2)

**Goal**: Support both polling and webhooks simultaneously

#### 2.1 Configuration (1 day)

**Add to ceye.yaml**:

```yaml
webhooks:
  enabled: true
  port: 9090
  github_secret: "${GITHUB_WEBHOOK_SECRET}"
  azure_user: "${AZURE_WEBHOOK_USER}"
  azure_password: "${AZURE_WEBHOOK_PASS}"

providers:
  - type: github
    display_name: "GitHub"
    mode: webhook  # or "polling" or "hybrid"
    repos:
      - owner: "myorg"
        repo: "myrepo"
    # Polling fallback settings
    fallback_interval: "5m"
  
  - type: azure
    display_name: "Azure"
    mode: webhook
    org: "myorg"
    projects:
      - name: "MyProject"
    fallback_interval: "5m"
```

#### 2.2 Provider Modes (2 days)

**Modes**:
1. **webhook**: Only receive webhooks (no polling)
2. **polling**: Traditional polling (current behavior)
3. **hybrid**: Webhooks + periodic polling fallback (recommended)

**Hybrid mode logic**:
- Receive webhooks for real-time updates
- Poll every 5-10 minutes as backup (catches missed webhooks)
- Reset poll timer on webhook receipt (reduces API calls)

```go
type WebhookProvider struct {
    name          string
    mode          ProviderMode
    pollInterval  time.Duration
    lastWebhook   time.Time
    mu            sync.Mutex
}

func (p *WebhookProvider) Start(ctx context.Context, out chan<- RunEvent) error {
    if p.mode == ModeWebhook {
        // Wait for webhooks only
        <-ctx.Done()
        return ctx.Err()
    }
    
    // Hybrid or polling mode
    ticker := time.NewTicker(p.pollInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            // Skip poll if recent webhook
            p.mu.Lock()
            timeSinceWebhook := time.Since(p.lastWebhook)
            p.mu.Unlock()
            
            if p.mode == ModeHybrid && timeSinceWebhook < p.pollInterval {
                continue
            }
            
            // Perform poll...
        }
    }
}

func (p *WebhookProvider) OnWebhook() {
    p.mu.Lock()
    p.lastWebhook = time.Now()
    p.mu.Unlock()
}
```

#### 2.3 Migration Path (1 day)

**For users**:
1. Update ceye to version with webhook support
2. Optionally enable webhooks in config
3. Set up webhooks in GitHub/Azure DevOps
4. Monitor logs to verify webhooks work
5. Switch from polling to hybrid mode
6. Eventually disable polling entirely

**Backward compatibility**:
- Default to polling mode if webhooks not configured
- Graceful degradation if webhook server fails
- Clear error messages for configuration issues

### Phase 3: Setup Tools (Week 3)

**Goal**: Automate webhook setup for users

#### 3.1 CLI Commands (2 days)

```bash
# Setup GitHub webhooks for all configured repos
ci-dash webhooks setup github

# Setup Azure service hooks for all projects
ci-dash webhooks setup azure

# Verify webhooks are configured correctly
ci-dash webhooks verify

# List configured webhooks
ci-dash webhooks list

# Remove webhooks
ci-dash webhooks remove github
```

**Implementation**:
```go
// cmd/ci-dash/webhooks.go
func SetupGitHubWebhooks(cfg Config) error {
    client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
    
    webhookURL := fmt.Sprintf("https://%s/webhooks/github", cfg.Webhooks.PublicURL)
    
    for _, repo := range cfg.Providers.GitHub.Repos {
        hook := &github.Hook{
            Name:   github.String("web"),
            Active: github.Bool(true),
            Events: []string{"workflow_run", "check_run", "status"},
            Config: map[string]interface{}{
                "url":          webhookURL,
                "content_type": "json",
                "secret":       cfg.Webhooks.GitHubSecret,
            },
        }
        
        _, _, err := client.Repositories.CreateHook(ctx, repo.Owner, repo.Repo, hook)
        if err != nil {
            return err
        }
        
        fmt.Printf("✓ Created webhook for %s/%s\n", repo.Owner, repo.Repo)
    }
    
    return nil
}
```

#### 3.2 Documentation (1 day)

**New docs**:
- `docs/webhooks-setup.md` - Complete setup guide
- `docs/webhooks-troubleshooting.md` - Common issues
- Update `README.md` with webhook benefits

**Content**:
- Why webhooks are better than polling
- Step-by-step setup for GitHub and Azure
- ngrok setup for local development
- Security best practices
- Troubleshooting common issues

---

## Benefits Summary

| Aspect | Polling | Webhooks | Improvement |
|--------|---------|----------|-------------|
| **Rate Limits** | 5,000/hour GitHub | No limit | ∞ |
| **Latency** | 10-60 seconds | < 1 second | 10-60x faster |
| **API Costs** | High | Zero | 100% reduction |
| **Scalability** | Limited by quotas | Unlimited repos | ∞ |
| **CPU Usage** | Constant polling | Event-driven | ~90% reduction |
| **Accuracy** | May miss events | Real-time | 100% |

---

## Risks & Mitigations

### Risk 1: Missed Webhooks
**Issue**: Webhook delivery isn't 100% reliable (network issues, downtime)

**Mitigation**:
- Implement **hybrid mode** with periodic polling fallback
- GitHub retries failed webhooks up to 3 times
- Log all received webhooks for debugging
- Alert if no webhooks received for X minutes

### Risk 2: Webhook Authentication
**Issue**: Exposed webhook endpoint could receive spoofed requests

**Mitigation**:
- **HMAC signature validation** (GitHub) - cryptographically secure
- **Basic auth** (Azure) - username/password
- **HTTPS only** - encrypted transport
- **IP whitelisting** (optional) - limit to GitHub/Azure IPs
- **Rate limiting** - prevent webhook spam

### Risk 3: Public Endpoint Requirement
**Issue**: Need publicly accessible HTTPS endpoint

**Mitigation**:
- Document cloud deployment options (fly.io, Railway, render.com)
- Provide docker-compose with Caddy for auto-SSL
- ngrok for development/testing
- Support for webhook proxy services (smee.io)

### Risk 4: Configuration Complexity
**Issue**: Setting up webhooks is more complex than polling

**Mitigation**:
- Provide `ci-dash webhooks setup` CLI command
- Clear documentation with screenshots
- Automatic webhook creation via APIs
- Validation tool to test webhook configuration

---

## Recommended Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         GitHub                              │
│                    workflow_run event                       │
└────────────────┬────────────────────────────────────────────┘
                 │ HTTPS POST
                 │ X-Hub-Signature-256: sha256=...
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                    Webhook Receiver                         │
│                  (internal/webhooks)                        │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Validate   │→ │    Parse     │→ │   Convert    │    │
│  │  Signature   │  │   Payload    │  │  to RunEvent │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
└────────────────┬────────────────────────────────────────────┘
                 │ RunEvent channel
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                         Store                               │
│                    (internal/core)                          │
│                                                             │
│              Merge event → Update state                     │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ├─→ TUI update
                 └─→ WebSocket broadcast
```

---

## References

### GitHub
- [About webhooks](https://docs.github.com/get-started/customizing-your-github-workflow/exploring-integrations/about-webhooks)
- [Webhook events and payloads](https://docs.github.com/en/webhooks/webhook-events-and-payloads)
- [Creating webhooks](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks)
- [Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)
- [REST API webhooks endpoint](https://docs.github.com/en/rest/repos/webhooks)

### Azure DevOps
- [Webhooks with Azure DevOps](https://learn.microsoft.com/en-us/azure/devops/service-hooks/services/webhooks?view=azure-devops)
- [Service Hook Events](https://learn.microsoft.com/en-us/azure/devops/service-hooks/events?view=azure-devops)
- [Create service hook subscription API](https://learn.microsoft.com/en-us/rest/api/azure/devops/servicehooks/subscriptions/create?view=azure-devops-rest-6.0)

### Tools
- [ngrok](https://ngrok.com/) - HTTP tunnel for local development
- [smee.io](https://smee.io/) - GitHub webhook proxy
- [alexellis/hmac](https://github.com/alexellis/hmac) - Go HMAC validation library

---

## Next Steps

1. **Add to plan.md** as Phase 2.5 (Week 4-6)
2. **Create TDD tests** for webhook parsing
3. **Implement webhook server** with GitHub support first
4. **Test with ngrok** + real GitHub webhooks
5. **Add Azure DevOps support**
6. **Document setup process**
7. **Create CLI setup tools**
8. **Deploy and migrate users**

**Estimated effort**: 3-4 weeks for complete implementation
**Priority**: High - Solves major pain point (rate limits)
**Impact**: Massive improvement in responsiveness and scalability
