package aruba

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// listPageFetch returns a function that GETs an absolute pagination URL and returns
// the decoded response. Used by each adapter to wire List.Next/Prev/First/Last.
func listPageFetch[L any](rest *restclient.Client, opts []CallOption) func(ctx context.Context, absURL string) (*types.Response[L], error) {
	return func(ctx context.Context, absURL string) (*types.Response[L], error) {
		if rest == nil {
			return nil, fmt.Errorf("pagination unavailable: REST client is not initialised")
		}
		co := applyCallOptions(opts)
		rp := co.toRequestParameters()
		// The server-supplied pagination URL is authoritative; do not re-append
		// the original query params (limit, offset, filter, api-version, …) as
		// they are already baked into the link and duplicating them can produce
		// conflicting or repeated query keys. Only forward request-level headers
		// (e.g. Accept) that are not part of the URL.
		httpResp, err := rest.DoRequestAbs(ctx, http.MethodGet, absURL, nil, nil, rp.ToHeaders())
		if err != nil {
			return nil, err
		}
		defer httpResp.Body.Close()
		return types.ParseResponseBody[L](httpResp, rest.Logger())
	}
}
