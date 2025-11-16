package providers

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// SafeProvider wraps a Provider with panic recovery and validation
type SafeProvider struct {
	inner     core.Provider
	validator *EventValidator
}

// NewSafeProvider wraps a provider with safety guarantees
func NewSafeProvider(p core.Provider) *SafeProvider {
	return &SafeProvider{
		inner:     p,
		validator: NewEventValidator(),
	}
}

// Name returns the wrapped provider's name
func (s *SafeProvider) Name() string {
	return s.inner.Name()
}

// Start wraps the provider's Start method with panic recovery and validation
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
	validatedCh := make(chan core.RunEvent, 10)
	go s.validateAndForward(ctx, validatedCh, out)

	return s.inner.Start(ctx, validatedCh)
}

func (s *SafeProvider) validateAndForward(ctx context.Context, in <-chan core.RunEvent, out chan<- core.RunEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-in:
			if !ok {
				return
			}

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

// Unwrap returns the inner provider (useful for testing)
func (s *SafeProvider) Unwrap() core.Provider {
	return s.inner
}
