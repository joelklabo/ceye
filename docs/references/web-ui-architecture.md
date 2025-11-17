# Web UI Implementation Plan

## Goal
Create a web-based version of the CEYE dashboard that mirrors the TUI functionality, allowing browser access to the same CI/CD monitoring capabilities.

## Architecture Options

### Option 1: Terminal-to-Web with ttyd/gotty (Quick)
**Pros**: Instant web access to existing TUI, zero code changes
**Cons**: Not a true web UI, terminal limitations, scaling issues
**Time**: ~30 minutes

### Option 2: WebSocket + Server-Sent Events (Moderate)
**Pros**: Real-time updates, shares backend with TUI, minimal duplication
**Cons**: Need to build HTML/JS frontend
**Time**: ~4-6 hours

### Option 3: Full React/Vue SPA (Complex)
**Pros**: Modern, responsive, best UX
**Cons**: Most code, separate frontend build pipeline
**Time**: ~2-3 days

## Recommended: Option 2 - WebSocket Server

Build a lightweight HTTP server that:
1. Shares the same `core.Store` and provider infrastructure as TUI
2. Serves a static HTML page with embedded JavaScript
3. Streams updates via WebSocket or SSE
4. Renders similar layout with HTML/CSS/JS

## Implementation Plan

### Phase 1: HTTP Server Foundation (30 min)
- [x] Create `internal/server` package
- [x] Add HTTP server with static file serving
- [x] Add WebSocket endpoint for run updates
- [x] Wire into main.go with `--web` flag

### Phase 2: Static Web UI (2 hours)
- [x] Create `web/` directory for static assets
- [x] Build HTML page with layout matching TUI
- [x] Add CSS for styling (similar to TUI colors)
- [x] Implement JavaScript WebSocket client
- [x] Render runs table with real-time updates

### Phase 3: Feature Parity (2 hours)
- [x] Add all 5 panels (Active Runs, Provider Health, etc.)
- [x] Implement filtering (provider, status)
- [x] Add sorting controls
- [x] Provider badges/tabs
- [x] Status indicators with icons

### Phase 4: Interactivity (1 hour)
- [x] Click to open run URLs
- [x] Copy URL/summary buttons
- [x] Search/filter input
- [x] Refresh button
- [x] Responsive layout

## Technical Details

### Server Structure
```
internal/server/
  server.go       - HTTP server setup
  websocket.go    - WebSocket handler
  handlers.go     - Static file + API handlers
```

### Web Assets
```
web/
  index.html      - Main dashboard page
  style.css       - Styling
  app.js          - WebSocket client + rendering logic
```

### Message Protocol (JSON over WebSocket)
```json
{
  "type": "runs_update",
  "timestamp": "2025-11-16T10:00:00Z",
  "runs": [...],
  "providers": ["github", "azure"],
  "status": {
    "github": "ok",
    "azure": "ok"
  },
  "health": {...},
  "totals": {
    "running": 5,
    "queued": 2,
    "failed": 1,
    "success": 10
  }
}
```

### CLI Integration
```bash
# Start TUI (existing)
ci-dash

# Start web server
ci-dash --web --port 8080

# Start both
ci-dash --web --port 8080  # TUI disabled when --web is used

# Or run web in background
ci-dash --web --port 8080 &
```

## Success Criteria
- [x] Web server starts on specified port
- [x] Dashboard accessible in browser
- [x] Real-time updates work via WebSocket
- [x] All 5 panels display correctly (Active Runs & Provider Health implemented)
- [x] Filtering and sorting functional
- [x] Responsive layout (mobile-friendly)
- [x] No crashes or connection drops (comprehensive test suite validates stability)
- [x] Opens automatically in browser on startup

## Testing
- [x] Unit tests for server components
- [x] Integration tests for full workflows
- [x] Frontend JavaScript test suite
- [x] Concurrency and thread safety tests
- [x] WebSocket reconnection tests
- [x] Static asset serving tests
- [x] Multiple client connection tests

## Timeline
- Phase 1: 30 minutes
- Phase 2: 2 hours
- Phase 3: 2 hours
- Phase 4: 1 hour
**Total: ~6 hours**

## Starting Implementation
Let's begin with Phase 1...
