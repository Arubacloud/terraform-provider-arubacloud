// Package restclient provides HTTP interceptor helper functions for use with
// the Aruba Cloud SDK middleware chain.
//
// # Note on package name
//
// This package is located at pkg/util/middleware/ but is named "restclient"
// (see TD-025 in ai/TECH_DEBT.md for the planned rename). Import it as:
//
//	import restclient "github.com/Arubacloud/sdk-go/pkg/util/middleware"
//
// # Provided helpers
//
// Each function returns an interceptor.InterceptFunc suitable for use with
// aruba.NewOptions().WithCustomMiddleware(interceptor):
//
//   - [WithCustomHeaders] — attaches a static map of request headers.
//   - [WithUserAgent]     — sets the User-Agent header.
//   - [WithContentType]   — sets the Content-Type header.
//   - [WithAccept]        — sets the Accept header.
//
// These helpers are intended for cases where the built-in SDK middleware is
// insufficient. Most callers do not need to import this package directly.
package restclient
