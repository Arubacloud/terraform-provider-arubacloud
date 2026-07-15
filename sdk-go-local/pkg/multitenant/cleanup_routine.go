package multitenant

import (
	"context"
	"time"
)

// StartCleanupRoutine starts a background goroutine that periodically calls CleanUp on the Multitenant instance.
//
// Parameters:
//   - ctx: parent context for cancellation control
//   - m: the Multitenant instance to clean up
//   - tickInterval: duration between cleanup runs (default 1 hour if zero or negative)
//   - fromDuration: threshold duration passed to CleanUp() to remove inactive clients
//
// Returns:
//   - context.CancelFunc: a function to cancel the cleanup routine and stop the goroutine
//
// The routine runs in a separate goroutine and will:
//   - Execute CleanUp(fromDuration) on each tick
//   - Exit gracefully when ctx.Done() fires or the cancel function is called
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	cleanupCancel := StartCleanupRoutine(ctx, mt, 5*time.Minute, 24*time.Hour)
//	defer cleanupCancel()
func StartCleanupRoutine(ctx context.Context, m Multitenant, tickInterval time.Duration, fromDuration time.Duration) context.CancelFunc {
	if tickInterval <= 0 {
		tickInterval = 1 * time.Hour // default 1 hour
	}

	if fromDuration <= 0 {
		fromDuration = 24 * time.Hour // default 24 hours
	}

	// Create a child context that we can cancel
	cleanupCtx, cancel := context.WithCancel(ctx)

	// Launch the cleanup goroutine
	go func() {
		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()

		for {
			select {
			case <-cleanupCtx.Done():
				return
			case <-ticker.C:
				// Execute cleanup
				m.CleanUp(fromDuration)
			}
		}
	}()

	return cancel
}
