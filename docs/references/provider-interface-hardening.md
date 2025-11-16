# Provider Interface Hardening Plan

## Current State Analysis

The `core.Provider` interface is minimal:
```go
type Provider interface {
    Name() string
    Start(ctx context.Context, out chan<- RunEvent) error
}
```

### Current Issues & Risks

1. **No Panic Recovery** ❌
   - Provider panic crashes entire application
   - No isolation between providers
   - One bad provider kills all monitoring

2. **No Input Validation** ⚠️
   - Providers can send invalid/malformed Runs
   - Empty IDs not consistently rejected
   - No validation of required fields
   - Invalid timestamps/durations not checked

3. **No Rate Limiting** ⚠️
   - Provider can flood channel with events
   - No backpressure mechanism
   - Can overwhelm UI/web clients

4. **No Timeout Protection** ⚠️
   - API calls can hang indefinitely
   - No per-request timeout enforcement
   - Stuck providers block resources

5. **No Health Tracking** ⚠️
   - Health info is tracked in main.go, not encapsulated
   - No standardized health check
   - Hard to diagnose provider issues

6. **No Graceful Degradation** ⚠️
   - Provider errors logged but not surfaced well
   - No circuit breaker pattern
   - No exponential backoff on repeated failures

7. **No Testing Contract** ⚠️
   - No standard test suite providers must pass
   - Each provider has different test coverage
   - No interface compliance tests

## Hardening Strategy

### Phase 1: Safety Wrapper (Critical - 2 hours)

Create a wrapper that:
- Catches and logs panics
- Validates all events before sending
- Adds per-provider timeout
- Implements circuit breaker

```go
// SafeProvider wraps any Provider with safety guarantees
type SafeProvider struct {
    inner     core.Provider
    validator *EventValidator
    breaker   *CircuitBreaker
    timeout   time.Duration
}
```

### Phase 2: Validation Layer (High Priority - 1 hour)

```go
// EventValidator ensures all events meet requirements
type EventValidator struct {
    rules []ValidationRule
}

type ValidationRule interface {
    Validate(event RunEvent) error
}

// Built-in rules:
// - RequiredFieldsRule: ID, Provider, Status must be set
// - TimestampRule: Timestamps must be reasonable
// - DurationRule: Duration must be non-negative
// - StatusRule: Status must be valid enum value
```

### Phase 3: Rate Limiting (Medium Priority - 1 hour)

```go
// RateLimiter prevents provider event floods
type RateLimiter struct {
    maxEventsPerSecond int
    burst              int
}
```

### Phase 4: Health Interface (Medium Priority - 1 hour)

```go
// Healthcheck extends Provider with health reporting
type Healthcheck interface {
    Provider
    Health() HealthStatus
}

type HealthStatus struct {
    State        HealthState  // Healthy, Degraded, Unhealthy
    LastSuccess  time.Time
    LastError    time.Time
    ErrorCount   int
    ErrorMessage string
    Metrics      map[string]interface{} // Custom metrics
}
```

### Phase 5: Testing Contract (High Priority - 2 hours)

```go
// ProviderTestSuite validates any Provider implementation
type ProviderTestSuite struct {
    t        *testing.T
    provider core.Provider
}

// Tests to enforce:
// - Start respects context cancellation
// - Events have valid structure
// - Name is not empty
// - No panics under various conditions
// - Concurrent safety
```

### Phase 6: Circuit Breaker (Medium Priority - 1 hour)

```go
// CircuitBreaker prevents cascading failures
type CircuitBreaker struct {
    maxFailures     int
    resetTimeout    time.Duration
    state           BreakerState // Closed, Open, HalfOpen
}
```

## Implementation Order

### Priority 1: Immediate Safety (Do First)
1. ✅ Panic recovery wrapper
2. ✅ Basic event validation
3. ✅ Provider test suite

### Priority 2: Reliability (Do Soon)
4. ✅ Circuit breaker
5. ✅ Health interface
6. ✅ Timeout protection

### Priority 3: Performance (Do Later)
7. Rate limiting
8. Backpressure handling
9. Advanced metrics

## Detailed Implementation

### 1. Panic Recovery Wrapper

```go
package providers

import (
    "context"
    "fmt"
    "log"
    "runtime/debug"
    "time"
    
    "github.com/joelklabo/ceye/internal/core"
)

// SafeProvider wraps a Provider with panic recovery
type SafeProvider struct {
    inner     core.Provider
    validator *EventValidator
}

func NewSafeProvider(p core.Provider) *SafeProvider {
    return &SafeProvider{
        inner:     p,
        validator: NewEventValidator(),
    }
}

func (s *SafeProvider) Name() string {
    return s.inner.Name()
}

func (s *SafeProvider) Start(ctx context.Context, out chan<- core.RunEvent) (err error) {
    defer func() {
        if r := recover(); r != nil {
            stack := debug.Stack()
            log.Printf("PANIC in provider %s: %v\n%s", s.Name(), r, stack)
            
            // Send error event before returning
            select {
            case out <- core.RunEvent{
                Provider:  s.Name(),
                Timestamp: time.Now(),
                Err:       fmt.Errorf("provider panic: %v", r),
            }:
            case <-ctx.Done():
            }
            
            err = fmt.Errorf("provider panic: %v", r)
        }
    }()
    
    // Wrap the channel to intercept and validate events
    validatedCh := make(chan core.RunEvent)
    go s.validateAndForward(ctx, validatedCh, out)
    
    return s.inner.Start(ctx, validatedCh)
}

func (s *SafeProvider) validateAndForward(ctx context.Context, in <-chan core.RunEvent, out chan<- core.RunEvent) {
    for {
        select {
        case <-ctx.Done():
            return
        case event := <-in:
            if err := s.validator.Validate(event); err != nil {
                log.Printf("Invalid event from %s: %v", s.Name(), err)
                // Forward error event
                select {
                case out <- core.RunEvent{
                    Provider:  s.Name(),
                    Timestamp: time.Now(),
                    Err:       fmt.Errorf("validation failed: %w", err),
                }:
                case <-ctx.Done():
                    return
                }
                continue
            }
            
            // Forward valid event
            select {
            case out <- event:
            case <-ctx.Done():
                return
            }
        }
    }
}
```

### 2. Event Validator

```go
package providers

import (
    "fmt"
    "time"
    
    "github.com/joelklabo/ceye/internal/core"
)

type EventValidator struct {
    rules []ValidationRule
}

type ValidationRule interface {
    Validate(event core.RunEvent) error
}

func NewEventValidator() *EventValidator {
    return &EventValidator{
        rules: []ValidationRule{
            &RequiredFieldsRule{},
            &TimestampRule{},
            &StatusRule{},
            &DurationRule{},
        },
    }
}

func (v *EventValidator) Validate(event core.RunEvent) error {
    for _, rule := range v.rules {
        if err := rule.Validate(event); err != nil {
            return err
        }
    }
    return nil
}

// RequiredFieldsRule ensures critical fields are set
type RequiredFieldsRule struct{}

func (r *RequiredFieldsRule) Validate(event core.RunEvent) error {
    if event.Provider == "" {
        return fmt.Errorf("provider name is required")
    }
    
    for i, run := range event.Runs {
        if run.ID == "" {
            return fmt.Errorf("run[%d]: ID is required", i)
        }
        if run.Provider == "" {
            return fmt.Errorf("run[%d]: Provider is required", i)
        }
        if run.Status == "" {
            return fmt.Errorf("run[%d]: Status is required", i)
        }
    }
    
    return nil
}

// TimestampRule validates timestamps are reasonable
type TimestampRule struct{}

func (r *TimestampRule) Validate(event core.RunEvent) error {
    now := time.Now()
    future := now.Add(24 * time.Hour)
    past := now.Add(-365 * 24 * time.Hour)
    
    for i, run := range event.Runs {
        if !run.UpdatedAt.IsZero() {
            if run.UpdatedAt.After(future) {
                return fmt.Errorf("run[%d]: UpdatedAt is in the future", i)
            }
            if run.UpdatedAt.Before(past) {
                return fmt.Errorf("run[%d]: UpdatedAt is too far in the past", i)
            }
        }
        
        if !run.StartedAt.IsZero() {
            if run.StartedAt.After(future) {
                return fmt.Errorf("run[%d]: StartedAt is in the future", i)
            }
            if run.StartedAt.Before(past) {
                return fmt.Errorf("run[%d]: StartedAt is too far in the past", i)
            }
        }
    }
    
    return nil
}

// StatusRule validates status values
type StatusRule struct{}

func (r *StatusRule) Validate(event core.RunEvent) error {
    validStatuses := map[core.RunStatus]bool{
        core.RunStatusUnknown:    true,
        core.RunStatusQueued:     true,
        core.RunStatusInProgress: true,
        core.RunStatusCompleted:  true,
        core.RunStatusFailed:     true,
        core.RunStatusCancelled:  true,
    }
    
    for i, run := range event.Runs {
        if !validStatuses[run.Status] {
            return fmt.Errorf("run[%d]: invalid status %q", i, run.Status)
        }
    }
    
    return nil
}

// DurationRule validates durations are non-negative
type DurationRule struct{}

func (r *DurationRule) Validate(event core.RunEvent) error {
    for i, run := range event.Runs {
        if run.Duration < 0 {
            return fmt.Errorf("run[%d]: duration cannot be negative", i)
        }
    }
    return nil
}
```

### 3. Provider Test Suite

```go
package providers_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/joelklabo/ceye/internal/core"
)

// ProviderTestSuite tests any Provider implementation
type ProviderTestSuite struct {
    t        *testing.T
    provider core.Provider
    timeout  time.Duration
}

func NewProviderTestSuite(t *testing.T, provider core.Provider) *ProviderTestSuite {
    return &ProviderTestSuite{
        t:        t,
        provider: provider,
        timeout:  5 * time.Second,
    }
}

func (s *ProviderTestSuite) Run() {
    s.t.Run("Name", s.testName)
    s.t.Run("ContextCancellation", s.testContextCancellation)
    s.t.Run("EventStructure", s.testEventStructure)
    s.t.Run("NoPanic", s.testNoPanic)
    s.t.Run("ConcurrentSafety", s.testConcurrentSafety)
}

func (s *ProviderTestSuite) testName() {
    name := s.provider.Name()
    if name == "" {
        s.t.Error("provider Name() returned empty string")
    }
}

func (s *ProviderTestSuite) testContextCancellation() {
    ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
    defer cancel()
    
    eventCh := make(chan core.RunEvent, 10)
    errCh := make(chan error, 1)
    
    go func() {
        errCh <- s.provider.Start(ctx, eventCh)
    }()
    
    cancel()
    
    select {
    case err := <-errCh:
        if err != context.Canceled && err != context.DeadlineExceeded {
            s.t.Errorf("expected context error, got %v", err)
        }
    case <-time.After(2 * time.Second):
        s.t.Error("provider did not respect context cancellation within 2s")
    }
}

func (s *ProviderTestSuite) testEventStructure() {
    ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
    defer cancel()
    
    eventCh := make(chan core.RunEvent, 10)
    
    go s.provider.Start(ctx, eventCh)
    
    select {
    case event := <-eventCh:
        if event.Provider == "" {
            s.t.Error("event missing Provider field")
        }
        
        for i, run := range event.Runs {
            if run.ID == "" {
                s.t.Errorf("run[%d] missing ID", i)
            }
            if run.Provider == "" {
                s.t.Errorf("run[%d] missing Provider", i)
            }
            if run.Status == "" {
                s.t.Errorf("run[%d] missing Status", i)
            }
        }
    case <-time.After(s.timeout):
        // No event received - acceptable for some providers
    }
}

func (s *ProviderTestSuite) testNoPanic() {
    defer func() {
        if r := recover(); r != nil {
            s.t.Errorf("provider panicked: %v", r)
        }
    }()
    
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    eventCh := make(chan core.RunEvent, 10)
    s.provider.Start(ctx, eventCh)
}

func (s *ProviderTestSuite) testConcurrentSafety() {
    ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
    defer cancel()
    
    // Start provider multiple times concurrently
    for i := 0; i < 3; i++ {
        go func() {
            eventCh := make(chan core.RunEvent, 10)
            s.provider.Start(ctx, eventCh)
        }()
    }
    
    time.Sleep(500 * time.Millisecond)
    // If we get here without panic, test passes
}
```

## Migration Path

1. **Week 1**: Implement SafeProvider wrapper
2. **Week 1**: Add EventValidator
3. **Week 1**: Create ProviderTestSuite
4. **Week 2**: Add tests to all existing providers
5. **Week 2**: Implement CircuitBreaker
6. **Week 3**: Add Health interface
7. **Week 3**: Documentation and examples

## Success Metrics

- ✅ Zero provider panics crash the application
- ✅ All events validated before processing
- ✅ All providers pass standard test suite
- ✅ Circuit breaker triggers on repeated failures
- ✅ Health status visible in UI
- ✅ 95%+ uptime even with provider failures

## Testing Plan

1. **Fault Injection Tests**
   - Provider that panics
   - Provider that sends invalid data
   - Provider that hangs
   - Provider with memory leaks

2. **Chaos Testing**
   - Kill random providers
   - Inject random errors
   - Network failures
   - Rate limit exhaustion

3. **Load Testing**
   - 100+ providers
   - 10,000+ runs
   - Event flood scenarios
