package restclient

import (
	"context"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/ports/interceptor"
)

// TODO: review the placement of this file and the utility of these functions.

// WithCustomHeaders returns a interceptor.InterceptFunc that adds custom headers
func WithCustomHeaders(headers map[string]string) interceptor.InterceptFunc {
	return func(ctx context.Context, r *http.Request) error {
		for k, v := range headers {
			r.Header.Set(k, v)
		}
		return nil
	}
}

// WithUserAgent returns a interceptor.InterceptFunc that sets the User-Agent header
func WithUserAgent(userAgent string) interceptor.InterceptFunc {
	return func(ctx context.Context, r *http.Request) error {
		r.Header.Set("User-Agent", userAgent)
		return nil
	}
}

// WithContentType returns a interceptor.InterceptFunc that sets the Content-Type header
func WithContentType(contentType string) interceptor.InterceptFunc {
	return func(ctx context.Context, r *http.Request) error {
		r.Header.Set("Content-Type", contentType)
		return nil
	}
}

// WithAccept returns a interceptor.InterceptFunc that sets the Accept header
func WithAccept(accept string) interceptor.InterceptFunc {
	return func(ctx context.Context, r *http.Request) error {
		r.Header.Set("Accept", accept)
		return nil
	}
}
