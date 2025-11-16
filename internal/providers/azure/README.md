# Azure DevOps Provider

Complete Azure DevOps provider implementation for ceye, monitoring pipeline builds across multiple projects.

## Features

- ✅ Full Azure DevOps REST API v7.1 support
- ✅ Personal Access Token (PAT) authentication
- ✅ Multi-project monitoring
- ✅ Per-project pipeline filtering
- ✅ Adaptive polling (fast when active, slow when idle)
- ✅ Retry logic with exponential backoff
- ✅ Rate limiting compliance
- ✅ Comprehensive status mapping
- ✅ Build timeline/stage support
- ✅ Contract compliance tests

## Configuration

### Basic Configuration

```yaml
providers:
  - type: azure
    display_name: "Azure Production"
    org: "myorganization"
    projects:
      - name: "WebApp"
        pipelines: [123, 456]  # Specific pipeline IDs
      
      - name: "API"
        pipelines: []  # Empty = all pipelines in project
```

### Environment Variables

Set one of these environment variables with your Azure DevOps Personal Access Token:

```bash
export AZURE_PAT="your-pat-token"
# OR
export AZURE_DEVOPS_PAT="your-pat-token"
```

### Creating a Personal Access Token

1. Go to Azure DevOps: `https://dev.azure.com/{your-org}`
2. Click on **User Settings** (top right) → **Personal Access Tokens**
3. Click **New Token**
4. Set:
   - **Name**: ceye-monitor
   - **Organization**: Your organization
   - **Expiration**: 90 days (or custom)
   - **Scopes**: 
     - ✅ Build (Read)
     - ✅ Code (Read) - for repository info
5. Click **Create** and copy the token
6. Set environment variable: `export AZURE_PAT="your-token"`

### Configuration Schema

```go
type Config struct {
    DisplayName  string          // Optional: friendly name for this provider
    Org          string          // Required: Azure DevOps organization name
    Projects     []ProjectConfig // Required: projects to monitor
    FastInterval time.Duration   // Optional: polling when builds are active
    SlowInterval time.Duration   // Optional: polling when builds are idle
}

type ProjectConfig struct {
    Name      string // Required: project name
    Pipelines []int  // Optional: pipeline IDs (empty = all pipelines)
}
```

## Usage

### Creating a Provider

```go
import "github.com/joelklabo/ceye/internal/providers/azure"

// From configuration
cfg := azure.Config{
    DisplayName: "Azure Prod",
    Org:         "myorg",
    Projects: []azure.ProjectConfig{
        {Name: "WebApp", Pipelines: []int{123, 456}},
        {Name: "API", Pipelines: nil}, // All pipelines
    },
}

provider := azure.NewProviderFromConfig(cfg)

// Or with custom client
client := azure.NewClient("myorg", "my-pat-token")
provider := azure.NewProvider(client, cfg)

// Start monitoring
ctx := context.Background()
events := make(chan core.RunEvent)
go provider.Start(ctx, events)

// Receive events
for evt := range events {
    if evt.Err != nil {
        log.Printf("Error: %v", evt.Err)
        continue
    }
    for _, run := range evt.Runs {
        log.Printf("Build: %s - %s", run.WorkflowName, run.Status)
    }
}
```

### Multiple Projects

Monitor multiple projects with different pipeline configurations:

```go
cfg := azure.Config{
    Org: "myorg",
    Projects: []azure.ProjectConfig{
        // Monitor specific pipelines in production
        {
            Name:      "Production",
            Pipelines: []int{123, 456, 789},
        },
        // Monitor all pipelines in development
        {
            Name:      "Development",
            Pipelines: nil,
        },
        // Monitor specific critical pipelines in staging
        {
            Name:      "Staging",
            Pipelines: []int{100, 200},
        },
    },
}
```

## API Client

### Client Methods

```go
// Create client
client := azure.NewClient("myorg", "my-pat")

// List builds (recent 50)
runs, err := client.ListBuilds("myorg", "myproject", []int{123})

// Get build details
build, err := client.GetBuildDetails("myorg", "myproject", 12345)

// List all pipelines
pipelines, err := client.ListPipelines("myorg", "myproject")

// Get build timeline (stages/jobs)
timeline, err := client.GetBuildTimeline("myorg", "myproject", 12345)
```

### Retry and Rate Limiting

The client automatically:
- Retries failed requests up to 3 times
- Uses exponential backoff (1s, 2s, 3s)
- Handles rate limiting (429 responses)
- Respects `Retry-After` header
- Times out after 30 seconds

## Status Mapping

### Azure Status → core.RunStatus

| Azure Status    | core.RunStatus       | Description |
|----------------|----------------------|-------------|
| `notStarted`   | `RunStatusQueued`    | Build queued but not started |
| `inProgress`   | `RunStatusInProgress`| Build currently running |
| `completed`    | *depends on result*  | Build finished, check result |
| `cancelling`   | `RunStatusCancelled` | Build being cancelled |
| `postponed`    | `RunStatusQueued`    | Build postponed/delayed |

### Azure Result → Status (when Completed)

| Azure Result       | core.RunStatus       | Conclusion |
|-------------------|----------------------|------------|
| `succeeded`       | `RunStatusCompleted` | `success`  |
| `partiallySucceeded` | `RunStatusCompleted` | `partial_success` |
| `failed`          | `RunStatusFailed`    | `failure`  |
| `canceled`        | `RunStatusCancelled` | `cancelled` |

## Adaptive Polling

The provider adjusts polling frequency based on build activity:

- **Fast Interval** (15s): When builds are queued or in progress
- **Slow Interval** (60s): When all builds are completed/idle

This reduces API load while maintaining responsiveness during active builds.

### Custom Intervals

```go
cfg := azure.Config{
    Org:          "myorg",
    Projects:     projects,
    FastInterval: 10 * time.Second,  // Poll every 10s when active
    SlowInterval: 120 * time.Second, // Poll every 2m when idle
}
```

## Testing

### Unit Tests

```bash
go test ./internal/providers/azure/
```

Tests include:
- Status/conclusion mapping (14 tests)
- Branch name cleaning (5 tests)
- Build parsing (comprehensive)
- Provider lifecycle (2 tests)
- Contract compliance (8 tests)

### Contract Tests

The Azure provider passes all Provider interface contract tests:

- ✅ Non-empty name
- ✅ Context cancellation respected
- ✅ Well-formed events
- ✅ Concurrent access safe
- ✅ Multiple Start() calls handled
- ✅ No channel deadlocks
- ✅ Name stability
- ✅ Context timeout handling

### Integration Tests

```bash
# With real Azure DevOps API (requires AZURE_PAT)
export AZURE_PAT="your-token"
go test -tags=integration ./internal/providers/azure/
```

## Architecture

```
┌─────────────────┐
│  Azure Provider │
└────────┬────────┘
         │ uses
         ▼
┌─────────────────┐      ┌──────────────────┐
│  Azure Client   │──────▶│ Azure DevOps API │
└─────────────────┘      └──────────────────┘
         │
         │ returns
         ▼
┌─────────────────┐
│   core.Run[]    │
└─────────────────┘
```

### Components

1. **Provider** (`provider.go`)
   - Implements `core.Provider` interface
   - Manages polling lifecycle
   - Handles multi-project coordination
   - Emits `core.RunEvent` via channel

2. **Client** (`client.go`)
   - Azure DevOps REST API wrapper
   - Authentication (PAT)
   - Retry and rate limiting
   - Response parsing

3. **Types** (`client.go`)
   - `AzureBuild` - Build response structure
   - `Pipeline` - Pipeline definition
   - `Timeline` - Build stages/jobs
   - Mapping functions to `core.Run`

## Error Handling

### Provider Errors

The provider handles errors gracefully:

```go
// API errors are emitted as events
evt := core.RunEvent{
    Provider:  "azure",
    Err:       err,
    Message:   "Failed to fetch builds",
    Timestamp: time.Now(),
}
```

### Client Errors

The client provides detailed error messages:

- `"org and project required"` - Missing configuration
- `"api error 401: ..."` - Authentication failure
- `"api error 404: ..."` - Project/pipeline not found
- `"rate limited"` - Too many requests
- `"request failed after 3 attempts: ..."` - Persistent failure

## Performance

### API Call Optimization

- Fetches top 50 recent builds per project
- Parallel project queries (future enhancement)
- Caches pipeline definitions (future enhancement)
- Efficient status change detection (future enhancement)

### Resource Usage

- Memory: ~2-5MB per provider instance
- CPU: Minimal (polling + JSON parsing)
- Network: ~1-2 KB per request
- Rate limit: ~300 builds/project/poll

## Roadmap

### Planned Enhancements

- [ ] OAuth authentication support
- [ ] Service Principal authentication
- [ ] Build artifact information
- [ ] Test result summaries
- [ ] Parallel project queries
- [ ] Pipeline definition caching
- [ ] Change detection (only emit on changes)
- [ ] Configurable build limit per project
- [ ] Stage-level status tracking
- [ ] Build log streaming (future)

### Compatibility

- **Azure DevOps API**: v7.1
- **Go**: 1.21+
- **ceye**: v1.0+

## Troubleshooting

### Common Issues

**"api error 401: ..."**
- Check `AZURE_PAT` environment variable is set
- Verify PAT has Build (Read) scope
- Ensure PAT hasn't expired

**"api error 404: project not found"**
- Verify project name is correct (case-sensitive)
- Check you have access to the project
- Ensure project exists in the organization

**"rate limited"**
- Increase polling intervals
- Reduce number of monitored pipelines
- Check Azure DevOps rate limits

**No builds appearing**
- Verify pipelines have recent runs
- Check pipeline IDs are correct
- Ensure PAT has correct permissions

### Debug Logging

```go
import "log"

// Enable debug logging
log.SetFlags(log.LstdFlags | log.Lshortfile)

// Provider logs errors with context
// Look for: "azure provider error for org/project: ..."
```

## Examples

See:
- `provider_test.go` - Unit tests with examples
- `contract_test.go` - Contract compliance examples
- `client_test.go` - API client usage examples

## Contributing

When adding features:

1. Add tests (unit + contract)
2. Update this README
3. Update `AGENTS.md` if architecture changes
4. Follow existing code patterns
5. Run full test suite before committing

## License

Part of the ceye project. See project LICENSE.
