// Package standard provides a concrete, default implementation of the
// `interceptor.Interceptor` and `interceptor.Interceptable` interfaces.
package standard

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/ports/interceptor"
)

// Interceptor is the concrete type that holds and executes a chain of
// interceptor.InterceptFunc functions.
//
// interceptFuncs is expected to be frozen after construction. See Bind for the
// concurrency contract.
type Interceptor struct {
	interceptFuncs []interceptor.InterceptFunc
}

var _ interceptor.Interceptable = (*Interceptor)(nil)
var _ interceptor.Interceptor = (*Interceptor)(nil)

// NewInterceptor creates and returns a pointer to a new, empty Interceptor
// instance.
//
// The returned interceptor has no functions bound to it initially.
func NewInterceptor() *Interceptor {
	return &Interceptor{}
}

// NewInterceptorWithFuncs creates and returns a pointer to a new Interceptor
// instance, initialized with the provided intercept functions.
//
// It returns an error if any of the provided functions are nil.
func NewInterceptorWithFuncs(interceptFuncs ...interceptor.InterceptFunc) (*Interceptor, error) {
	if err := validateInterceptFuncs(interceptFuncs...); err != nil {
		return nil, fmt.Errorf("%w: %w", interceptor.ErrInvalidInterceptFunc, err)
	}

	return &Interceptor{interceptFuncs: interceptFuncs}, nil
}

// Bind implements interceptor.Interceptable. It is intended for
// construction/setup only and must not be called concurrently with Intercept.
func (i *Interceptor) Bind(interceptFuncs ...interceptor.InterceptFunc) error {
	if err := validateInterceptFuncs(interceptFuncs...); err != nil {
		return err
	}

	i.interceptFuncs = append(i.interceptFuncs, interceptFuncs...)

	return nil
}

// Intercept implements interceptor.Interceptor. Concurrent calls to Intercept
// are safe as long as Bind is not called concurrently.
func (i *Interceptor) Intercept(ctx context.Context, r *http.Request) error {
	if r == nil {
		return fmt.Errorf("%w: nil http requests are not allowed to be intercepted", interceptor.ErrInvalidHTTPRequest)
	}

	for _, interceptFunc := range i.interceptFuncs {
		if err := interceptFunc(ctx, r); err != nil {
			return fmt.Errorf("%w: %w", interceptor.ErrInterceptFuncFailed, err)
		}
	}

	return nil
}

// validateInterceptFuncs is an unexported helper function that checks if any of
// the provided intercept functions are nil.
func validateInterceptFuncs(interceptFuncs ...interceptor.InterceptFunc) error {
	for _, interceptFunc := range interceptFuncs {
		if interceptFunc == nil {
			return fmt.Errorf("%w: nil intercept function are not allowed to be bound", interceptor.ErrInvalidInterceptFunc)
		}
	}

	return nil
}
