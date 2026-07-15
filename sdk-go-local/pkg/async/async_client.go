package async

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// Default parameters for WaitFor
const (
	DefaultRetries   = 60
	DefaultBaseDelay = 10 * time.Second
	DefaultTimeout   = 600 * time.Second
)

// Result wraps the response from a WaitFor/Async operation and an error, if any.
type Result[T any] struct {
	Response *types.Response[T]
	Error    error
}

// AsyncClient represents a "future" for an async operation.
// It exposes a channel to receive the Result when ready.
type AsyncClient[T any] struct {
	resultCh chan Result[T]
	result   *Result[T] // cached result
	mu       sync.Mutex // protects access to result
}

// Await blocks until the asynchronous operation completes or the provided context is done.
// It returns the response from the async operation, or an error if the operation failed
// or the context was cancelled/expired.
//
// Only a single call to Await should be made per AsyncClient instance; subsequent calls
// will block until the same result is returned.
func (f *AsyncClient[T]) Await(ctx context.Context) (*types.Response[T], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.result != nil {
		return f.result.Response, f.result.Error
	}
	select {
	case res := <-f.resultCh:
		f.result = &res

		return res.Response, res.Error
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// DefaultWaitFor is a helper for WaitFor using default parameters:
// 60 retries, 10 seconds delay between attempts, 600 seconds (10 min) total timeout.
// Simplifies usage when you want standard retry behavior.
func DefaultWaitFor[T any](
	ctx context.Context,
	call func(ctx context.Context) (*types.Response[T], error),
	check func(*types.Response[T]) (bool, error),
) *AsyncClient[T] {
	if call == nil {
		asyncClient := &AsyncClient[T]{resultCh: make(chan Result[T], 1)}
		asyncClient.resultCh <- Result[T]{Response: nil, Error: fmt.Errorf("call function cannot be nil")}
		return asyncClient
	}
	if check == nil {
		return WaitFor(ctx, DefaultRetries, DefaultBaseDelay, DefaultTimeout, call, nil)
	}

	return WaitFor(ctx, DefaultRetries, DefaultBaseDelay, DefaultTimeout, call, check)
}

// WaitFor executes an async call repeatedly with retries, a fixed delay, and a timeout.
// The `call` function performs the API call.
// The `check` function decides if the result is acceptable (done).
func WaitFor[T any](
	ctx context.Context,
	retries int,
	baseDelay time.Duration,
	timeout time.Duration,
	call func(ctx context.Context) (*types.Response[T], error),
	check func(*types.Response[T]) (bool, error),
) *AsyncClient[T] {
	asyncClient := &AsyncClient[T]{resultCh: make(chan Result[T], 1)}

	// Validate that call and check are not nil
	if call == nil {
		asyncClient.resultCh <- Result[T]{Response: nil, Error: fmt.Errorf("call function cannot be nil")}
		return asyncClient
	}
	if check == nil {
		// default check: consider any non-nil response as "done"
		check = func(resp *types.Response[T]) (bool, error) {
			checkResource := resp != nil && resp.Data != nil
			if !checkResource {
				return false, fmt.Errorf("response nil")
			}

			return checkResource, nil
		}
	}

	go func() {
		var lastErr error

		ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		for range retries {

			// Check if timeout expired before attempting
			select {
			case <-ctxTimeout.Done():
				asyncClient.resultCh <- Result[T]{Response: nil, Error: ctxTimeout.Err()}
				return
			default:
			}

			// Perform the async call
			resp, err := call(ctxTimeout)

			if err != nil {
				// Record the last error, will return if retries exhausted
				lastErr = err
			} else {
				// Check if the response satisfies the done condition
				// ignore errors from check() and continue retrying
				result, err := check(resp)

				// if the check itself errors, record and continue
				if err != nil {
					lastErr = err
					continue
				}

				if result {
					asyncClient.resultCh <- Result[T]{Response: resp, Error: nil}
					return
				}
			}

			// Wait before next attempt or exit if context done
			select {
			case <-ctxTimeout.Done():
				asyncClient.resultCh <- Result[T]{Response: nil, Error: ctxTimeout.Err()}
				return
			case <-time.After(baseDelay):
				// no exponential backoff, constant delay
			}
		}
		// All retries exhausted, return last error
		asyncClient.resultCh <- Result[T]{Response: nil, Error: fmt.Errorf("after %d retries: %w", retries, lastErr)}
	}()

	return asyncClient
}
