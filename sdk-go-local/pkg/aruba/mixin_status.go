package aruba

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/async"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// statusMixin — resource lifecycle state
// --------------------------------------------------------------------------

// WaitOption configures WaitUntilActive / WaitUntilStates behaviour.
type WaitOption func(*waitOptions)

type waitOptions struct {
	retries   int
	baseDelay time.Duration
	timeout   time.Duration
}

func defaultWaitOptions() waitOptions {
	return waitOptions{
		retries:   async.DefaultRetries,
		baseDelay: async.DefaultBaseDelay,
		timeout:   async.DefaultTimeout,
	}
}

// WithRetries sets the maximum number of polling attempts (default: 60).
func WithRetries(n int) WaitOption { return func(o *waitOptions) { o.retries = n } }

// WithBaseDelay sets the fixed delay between polling attempts (default: 10s).
func WithBaseDelay(d time.Duration) WaitOption { return func(o *waitOptions) { o.baseDelay = d } }

// WithTimeout sets the overall deadline for the polling loop (default: 600s).
func WithTimeout(d time.Duration) WaitOption { return func(o *waitOptions) { o.timeout = d } }

func applyWaitOptions(opts []WaitOption) waitOptions {
	out := defaultWaitOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}
	return out
}

type statusMixin struct {
	refreshMixin
	status *types.ResourceStatusResponse
}

func (m *statusMixin) setStatus(s *types.ResourceStatusResponse) { m.status = s }

// State returns the current lifecycle state, or the zero State ("").
func (m *statusMixin) State() types.State {
	if m.status == nil || m.status.State == nil {
		return ""
	}
	return *m.status.State
}

// IsDisabled returns true when the server has disabled this resource.
func (m *statusMixin) IsDisabled() bool {
	if m.status == nil || m.status.DisableStatusInfoResponse == nil {
		return false
	}
	return m.status.DisableStatusInfoResponse.IsDisabled
}

// DisableReasons returns the reasons for disabling, or nil.
func (m *statusMixin) DisableReasons() []string {
	if m.status == nil || m.status.DisableStatusInfoResponse == nil {
		return nil
	}
	return m.status.DisableStatusInfoResponse.Reasons
}

// FailureReason returns the failure reason string, or "".
func (m *statusMixin) FailureReason() string {
	if m.status == nil || m.status.FailureReason == nil {
		return ""
	}
	return *m.status.FailureReason
}

// PreviousState returns the previous lifecycle state, or the zero State ("").
func (m *statusMixin) PreviousState() types.State {
	if m.status == nil || m.status.PreviousStatusResponse == nil || m.status.PreviousStatusResponse.State == nil {
		return ""
	}
	return *m.status.PreviousStatusResponse.State
}

// WaitUntilActive blocks until the resource reaches the "Active" state.
// Equivalent to WaitUntilStates(ctx, []State{StateActive}, opts...).
func (m *statusMixin) WaitUntilActive(ctx context.Context, opts ...WaitOption) error {
	return m.WaitUntilStates(ctx, []types.State{types.StateActive}, opts...)
}

// WaitUntilReady blocks until the resource reaches any healthy settled state.
// Use this when a caller does not care which steady state the resource lands in
// — only that it is no longer transitioning. Succeeds on Active, Running,
// Stopped, NotUsed, Reserved, InUse, or Used.
func (m *statusMixin) WaitUntilReady(ctx context.Context, opts ...WaitOption) error {
	return m.WaitUntilStates(ctx, []types.State{
		types.StateActive,
		types.StateRunning,
		types.StateStopped,
		types.StateNotUsed,
		types.StateReserved,
		types.StateInUse,
		types.StateUsed,
	}, opts...)
}

// WaitUntilStates blocks until the resource reaches any of the given target states.
//
// The check applies four rules in order:
//  1. state ∈ targets → success.
//  2. state.IsFailure() → terminal error.
//  3. state == "" || state.IsTransitory() → keep polling.
//  4. otherwise (settled, non-target) → terminal error.
//
// Rule 4 makes wait semantics context-dependent: a resource that settles in
// "Reserved" succeeds for a waiter that lists Reserved as a target and fails
// fast for one that does not. Returns a descriptive error if the refresh
// callback was not set (resource not produced by an adapter).
func (m *statusMixin) WaitUntilStates(ctx context.Context, targets []types.State, opts ...WaitOption) error {
	if m.refresh == nil {
		return errors.New("WaitUntilStates: refresh callback not set; resource must be produced by an adapter (Create/Get/Update/List) to support polling")
	}
	cfg := applyWaitOptions(opts)
	call := func(ctx context.Context) (*types.Response[any], error) {
		if err := m.refresh(ctx); err != nil {
			return nil, err
		}
		return &types.Response[any]{}, nil
	}
	var terminalErr error
	check := func(_ *types.Response[any]) (bool, error) {
		state := m.State()
		for _, t := range targets {
			if state == t {
				return true, nil
			}
		}
		if state.IsFailure() {
			terminalErr = fmt.Errorf("resource entered failure state %q (targets %v)", state, targets)
			return true, terminalErr
		}
		if state == "" || state.IsTransitory() {
			return false, nil
		}
		// settled, non-target, non-failure
		terminalErr = fmt.Errorf("resource settled in state %q which is not a wait target %v", state, targets)
		return true, terminalErr
	}
	_, err := async.WaitFor[any](ctx, cfg.retries, cfg.baseDelay, cfg.timeout, call, check).Await(ctx)
	if terminalErr != nil {
		return terminalErr
	}
	return err
}
