# Web Server Testing Guide

This directory contains comprehensive tests for the ceye web server implementation.

## Test Files

### `server_test.go` - Unit Tests
Tests individual components of the web server:

- **TestServerStaticFiles**: Verifies all static assets (HTML, CSS, JS) are served correctly
  - Index page loads with correct title "ceye"
  - CSS loads with expected styling
  - JavaScript loads with required functions

- **TestWebSocketConnection**: Tests WebSocket connection establishment
  - Connection upgrade succeeds
  - Initial snapshot is sent immediately
  - Message format is correct

- **TestBroadcastUpdate**: Tests broadcasting to multiple clients
  - Multiple WebSocket clients can connect
  - All clients receive broadcast updates
  - Messages are delivered in order

- **TestMessageFormat**: Validates JSON message structure
  - All required fields are present
  - Run totals are calculated correctly
  - Provider health is included

- **TestUpdateStatus**: Tests provider status updates
  - Status changes are persisted
  - Health metrics are tracked
  - Updates are reflected in snapshots

- **TestTotalsCalculation**: Tests run status totals
  - Running, queued, success, and failed counts
  - Edge cases (empty, all same status)
  - Correct classification of completed runs

- **TestJSONSerialization**: Validates JSON encoding/decoding
  - Messages serialize correctly
  - All fields round-trip properly

### `integration_test.go` - Integration Tests
Tests end-to-end workflows:

- **TestFullWorkflow**: Complete user journey
  1. Load index page via HTTP
  2. Connect WebSocket
  3. Receive initial snapshot
  4. Add runs and receive broadcast
  5. Update run status
  6. Update provider health
  7. Multiple providers working together

- **TestMultipleClients**: Concurrent client connections
  - Multiple clients connect simultaneously
  - All receive broadcasts correctly
  - No race conditions or deadlocks

- **TestStaticAssets**: HTTP asset serving
  - All static files accessible via HTTP
  - Correct Content-Type headers
  - Expected content in responses

- **TestWebSocketReconnection**: Reconnection handling
  - Client can disconnect and reconnect
  - State is maintained
  - New connection receives fresh snapshot

- **TestConcurrentUpdates**: Thread safety
  - Concurrent status updates don't cause crashes
  - No deadlocks under load
  - Mutex protection works correctly

### `web/app.test.html` - Frontend JavaScript Tests
Browser-based tests for the JavaScript client:

- **formatStatus**: Status code to display name conversion
- **formatDuration**: Nanosecond duration formatting
- **formatTimestamp**: Relative time display
- **escapeHtml**: XSS prevention
- **filters**: Client-side filtering logic
- **updateStats**: Stats card updates
- **updateProviderFilter**: Provider dropdown population
- **updateRunsTable**: Run table rendering with filtering
- **updateProviderHealth**: Provider health display
- **render**: Full UI update flow
- **updateConnectionStatus**: Connection indicator

## Running Tests

### Backend Tests
```bash
# Run all server tests
go test -v ./internal/server/...

# Run with race detection
go test -race ./internal/server/...

# Run specific test
go test -v ./internal/server/ -run TestFullWorkflow

# Run with coverage
go test -cover ./internal/server/...
```

### Frontend Tests
1. Start the web server:
   ```bash
   ./bin/ci-dash --demo --web --port 8080
   ```

2. Open in browser:
   ```
   http://localhost:8080/app.test.html
   ```

3. View test results in the browser UI

## Test Coverage

Current test coverage includes:

✅ **HTTP Server**
- Static file serving
- MIME type detection
- 404 handling

✅ **WebSocket**
- Connection establishment
- Initial snapshot delivery
- Message broadcasting
- Reconnection handling
- Concurrent connections

✅ **Message Format**
- JSON serialization
- Field validation
- Data types

✅ **Run Management**
- Run storage
- Status updates
- Provider filtering
- Totals calculation

✅ **Provider Health**
- Status tracking
- Error counting
- Timestamp recording

✅ **Frontend**
- Data formatting
- Filtering logic
- DOM updates
- XSS prevention

✅ **Concurrency**
- Thread safety
- Race condition prevention
- Deadlock avoidance

## Adding New Tests

### Backend Test Template
```go
func TestNewFeature(t *testing.T) {
    store := core.NewStore()
    srv := New(store, []string{"test"}, 8080)
    
    // Test setup
    
    // Perform actions
    
    // Assertions
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

### Frontend Test Template
```javascript
test("new feature works", () => {
    // Setup
    
    // Perform action
    
    // Assert
    assertEquals(actual, expected, "feature works");
});
```

## Continuous Integration

Tests should be run:
- Before each commit
- In CI/CD pipeline
- Before releases
- When reviewing PRs

## Known Limitations

- Frontend tests require manual browser execution
- WebSocket tests use local connections only
- No load testing included (add separately if needed)
- No accessibility testing (consider adding)

## Future Improvements

- [ ] Add E2E tests with real browsers (Playwright/Selenium)
- [ ] Add performance benchmarks
- [ ] Add accessibility tests (WCAG compliance)
- [ ] Add visual regression tests
- [ ] Add API contract tests
- [ ] Add load/stress tests
