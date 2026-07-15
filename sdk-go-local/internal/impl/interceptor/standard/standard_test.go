package standard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Arubacloud/sdk-go/internal/ports/interceptor"
)

func TestInterceptor_Bind(t *testing.T) {
	t.Run("should refuse nil interceptor functions", func(t *testing.T) {
		// Given a fresh standard interceptor
		instance := NewInterceptor()

		// And poisoned list containing valid and nil interceptor functions
		interceptFuncs := createPoisonedInterceptFuncs(t)

		// When we try to bind a nil intercept function
		err := instance.Bind(interceptFuncs...)

		// Then an invalid intercept function error is reported
		require.ErrorIs(t, err, interceptor.ErrInvalidInterceptFunc)

		// And the message informs that a nil intercept function was given as parameter
		require.ErrorContains(t, err, "nil intercept function are not allowed to be bound")

		// And the interceptor should not contain any intercept function
		require.Empty(t, instance.interceptFuncs)
	})

	t.Run("should accept valid interceptor functions", func(t *testing.T) {
		// Given a fresh standard interceptor
		instance := NewInterceptor()

		// And a list containing valid interceptor functions
		interceptFuncs := createValidInterceptFuncs(t)

		// When we try to bind the interceptor functions
		err := instance.Bind(interceptFuncs...)

		// Then no error should be reported
		require.NoError(t, err)

		// And the interceptor should contain the same functions of the list in the same order
		for i, interceptFunc := range interceptFuncs {
			require.True(t, reflect.ValueOf(interceptFunc).Pointer() == reflect.ValueOf(instance.interceptFuncs[i]).Pointer())
		}
	})
}

func TestInterceptor_Intercept(t *testing.T) {
	t.Run("should fail to intercept nil http requests", func(t *testing.T) {
		// Given a fresh standard interceptor
		instance := NewInterceptor()

		// When we try to intercept a nil http request
		err := instance.Intercept(t.Context(), nil)

		// Then an invalid http request error is reported
		require.ErrorIs(t, err, interceptor.ErrInvalidHTTPRequest)

		// And the message informs that the http request is nil
		require.ErrorContains(t, err, "nil http requests are not allowed to be intercepted")
	})

	t.Run("should not fail when no intercept function is bound", func(t *testing.T) {
		// Given a fresh standard interceptor
		instance := NewInterceptor()

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to intercept the request
		err := instance.Intercept(t.Context(), r)

		// Then no error should be reported
		require.NoError(t, err)
	})

	t.Run("should run intercept functions in order", func(t *testing.T) {
		// Given a standard interceptor with a set of valid intercept functions
		var tracer strings.Builder
		instance, _ := NewInterceptorWithFuncs(createValidInterceptFuncsWithTracer(t, &tracer)...)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to intercept the request
		err := instance.Intercept(t.Context(), r)

		// Then no error should be reported
		require.NoError(t, err)

		// And intercept functions should run in order
		require.Equal(t, "func_0func_1func_2", tracer.String())
	})

	t.Run("should fail when one of intercept functions fail", func(t *testing.T) {
		// Given a standard interceptor with a set of valid intercept functions which we know that one will fail
		instance, _ := NewInterceptorWithFuncs(createValidInterceptFuncsWithErrors(t)...)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to intercept the request
		err := instance.Intercept(t.Context(), r)

		// Then an intercept function failure error is reported
		require.ErrorIs(t, err, interceptor.ErrInterceptFuncFailed)

		// And the message informs the specific intercept function problem
		require.ErrorContains(t, err, "intercept function 1 failed")
	})
}

// Helpers
func createValidInterceptFuncs(t *testing.T) []interceptor.InterceptFunc {
	t.Helper()

	return []interceptor.InterceptFunc{
		func(ctx context.Context, r *http.Request) error { fmt.Println("func_0"); return nil },
		func(ctx context.Context, r *http.Request) error { fmt.Println("func_1"); return nil },
		func(ctx context.Context, r *http.Request) error { fmt.Println("func_2"); return nil },
	}
}

func createValidInterceptFuncsWithErrors(t *testing.T) []interceptor.InterceptFunc {
	t.Helper()

	return []interceptor.InterceptFunc{
		func(ctx context.Context, r *http.Request) error {
			fmt.Println("func_0")
			return nil
		},
		func(ctx context.Context, r *http.Request) error {
			fmt.Println("func_1")
			return errors.New("intercept function 1 failed")
		},
		func(ctx context.Context, r *http.Request) error {
			fmt.Println("func_2")
			return nil
		},
	}
}

func createValidInterceptFuncsWithTracer(t *testing.T, tracer io.Writer) []interceptor.InterceptFunc {
	t.Helper()

	return []interceptor.InterceptFunc{
		func(ctx context.Context, r *http.Request) error {
			tracer := tracer
			fmt.Fprintf(tracer, "func_0")
			return nil
		},
		func(ctx context.Context, r *http.Request) error {
			tracer := tracer
			fmt.Fprintf(tracer, "func_1")
			return nil
		},
		func(ctx context.Context, r *http.Request) error {
			tracer := tracer
			fmt.Fprintf(tracer, "func_2")
			return nil
		},
	}
}

func createPoisonedInterceptFuncs(t *testing.T) []interceptor.InterceptFunc {
	t.Helper()

	return []interceptor.InterceptFunc{
		func(ctx context.Context, r *http.Request) error { fmt.Println("func_0,"); return nil },
		nil,
		func(ctx context.Context, r *http.Request) error { fmt.Println("func_2,"); return nil },
	}
}
