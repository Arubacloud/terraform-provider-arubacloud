// Package async provides low-level polling primitives for long-running operations.
//
// # When to use this package
//
// Most callers should use the wrapper-level wait helpers instead:
//
//	result.WaitUntilReady(ctx)                     // waits for any healthy state
//	result.WaitUntilActive(ctx)                    // waits for StateActive
//	result.WaitUntilStates(ctx, targets, opts...)  // waits for a specific state set
//
// Reach for pkg/async directly when you need to:
//   - Start multiple resource waits concurrently (fan-out futures).
//   - Poll an arbitrary condition — not a resource lifecycle state.
//   - Control retries, base delay, and timeout independently per call.
//
// # Core types
//
// [WaitFor] launches a polling goroutine and returns an [AsyncClient] future.
// Call [AsyncClient.Await] to block for the result; subsequent calls return the
// cached value immediately (safe to call multiple times).
//
// [DefaultWaitFor] is a convenience wrapper that uses the package defaults:
//   - DefaultRetries   = 60  (poll attempts)
//   - DefaultBaseDelay = 10s (fixed delay between attempts)
//   - DefaultTimeout   = 600s (wall-clock deadline)
//
// # Polling semantics
//
// WaitFor retries the call function at most `retries` times with a fixed
// `baseDelay` between attempts (no exponential backoff — intentional for
// predictability). A separate `timeout` context deadline terminates the loop
// regardless of remaining retries.
//
// The check function receives the full *types.Response[T] and returns:
//   - (true, nil)   — success, stop polling and deliver the result.
//   - (true, error) — terminal failure, stop polling and deliver the error.
//   - (false, nil)  — keep polling.
//
// If check is nil, any non-nil response.Data is treated as success.
package async
