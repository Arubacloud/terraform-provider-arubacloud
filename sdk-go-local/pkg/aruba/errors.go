package aruba

import (
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// HTTPError is returned by Delete methods when the server responds with a non-2xx status.
// Callers can inspect StatusCode, Body, and ErrResp without unwrapping a resource wrapper.
type HTTPError struct {
	StatusCode int
	Body       []byte
	ErrResp    *types.ErrorResponse
}

func (e *HTTPError) Error() string {
	if e.ErrResp != nil {
		title := ""
		if e.ErrResp.Title != nil {
			title = *e.ErrResp.Title
		}
		detail := ""
		if e.ErrResp.Detail != nil {
			detail = *e.ErrResp.Detail
		}
		if detail != "" {
			return fmt.Sprintf("HTTP %d: %s — %s", e.StatusCode, title, detail)
		}
		if title != "" {
			return fmt.Sprintf("HTTP %d: %s", e.StatusCode, title)
		}
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}
