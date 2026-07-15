package aruba

import (
	"context"
	"errors"
	"net/http"

	"github.com/Arubacloud/sdk-go/pkg/async"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// refreshMixin — refresh callback + deletion wait
// --------------------------------------------------------------------------

// refreshMixin holds the adapter-supplied callback that re-fetches a resource
// from the server (a Get). Embedded by statusMixin and by Family-B resources
// that support polling but carry no lifecycle State.
type refreshMixin struct {
	refresh func(ctx context.Context) error
}

func (m *refreshMixin) setRefresh(fn func(context.Context) error) { m.refresh = fn }

// WaitUntilGone blocks until the resource no longer exists — that is, until a
// refresh (Get) returns HTTP 404. Use it after Delete to wait for teardown to
// complete before deleting the parent. Accepts the same WaitOptions as
// WaitUntilReady. Returns a descriptive error if the refresh callback was not
// set (resource not produced by an adapter).
func (m *refreshMixin) WaitUntilGone(ctx context.Context, opts ...WaitOption) error {
	if m.refresh == nil {
		return errors.New("WaitUntilGone: refresh callback not set; resource must be produced by an adapter (Create/Get/Update/List) to support polling")
	}
	cfg := applyWaitOptions(opts)
	call := func(ctx context.Context) (*types.Response[any], error) {
		err := m.refresh(ctx)
		if err == nil {
			return &types.Response[any]{}, nil // still exists — keep polling
		}
		var httpErr *HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			gone := any(struct{}{})
			return &types.Response[any]{Data: &gone}, nil // gone
		}
		return nil, err // transient — retry
	}
	check := func(resp *types.Response[any]) (bool, error) {
		return resp != nil && resp.Data != nil, nil
	}
	_, err := async.WaitFor[any](ctx, cfg.retries, cfg.baseDelay, cfg.timeout, call, check).Await(ctx)
	return err
}
