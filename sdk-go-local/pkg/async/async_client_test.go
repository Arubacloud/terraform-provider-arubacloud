package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

const (
	testRetries = 5
	testDelay   = 100 * time.Millisecond
	testTimeout = 1 * time.Second
)

type DummyData struct {
	Name string
}

func fakeCallSuccess(ctx context.Context) (*types.Response[DummyData], error) {
	return &types.Response[DummyData]{
		Data: &DummyData{Name: "ok"},
	}, nil
}

func fakeCallError(ctx context.Context) (*types.Response[DummyData], error) {
	return nil, errors.New("network error")
}

func checkAlwaysTrue(resp *types.Response[DummyData]) (bool, error) {
	return true, nil
}

func checkWaitForData(resp *types.Response[DummyData]) (bool, error) {
	if resp.Data != nil {
		return true, nil
	}
	return false, nil
}

func TestWaitFor_SucceedsImmediately(t *testing.T) {
	fut := WaitFor(
		t.Context(),
		testRetries,   // retries
		testDelay,     // delay
		2*time.Second, // timeout
		fakeCallSuccess,
		checkAlwaysTrue,
	)
	resp, err := fut.Await(t.Context())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp.Data.Name != "ok" {
		t.Fatalf("Expected data name 'ok', got: %s", resp.Data.Name)
	}
}

func TestWaitFor_SucceedsAfterRetries(t *testing.T) {
	attempt := 0

	fakeCallDelayed := func(ctx context.Context) (*types.Response[DummyData], error) {
		attempt++
		if attempt < 3 {
			// Not ready yet
			return &types.Response[DummyData]{Data: nil}, nil
		}
		return &types.Response[DummyData]{Data: &DummyData{Name: "ready"}}, nil
	}

	fut := WaitFor(
		t.Context(),
		testRetries,
		testDelay,
		2*time.Second,
		fakeCallDelayed,
		checkWaitForData,
	)

	resp, err := fut.Await(t.Context())

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if resp.Data.Name != "ready" {
		t.Fatalf("Unexpected data: %v", resp.Data)
	}
}

func TestWaitFor_FailsAfterRetries(t *testing.T) {
	fut := WaitFor(
		t.Context(),
		2,
		50*time.Millisecond,
		testTimeout,
		fakeCallError,
		checkAlwaysTrue,
	)

	_, err := fut.Await(t.Context())

	if err == nil {
		t.Fatalf("Expected error after retries, got nil")
	}
}

func TestWaitFor_CheckAlwaysFails(t *testing.T) {
	fut := WaitFor(
		t.Context(),
		testRetries,
		2*time.Millisecond,
		testTimeout,
		fakeCallSuccess,
		func(resp *types.Response[DummyData]) (bool, error) {
			return false, errors.New("check failed")
		},
	)

	_, err := fut.Await(t.Context())

	if err == nil {
		t.Fatalf("Expected error after retries, got nil")
	}
}

func TestWaitFor_TimesOut(t *testing.T) {
	fut := WaitFor(
		t.Context(),
		10,
		testDelay,
		250*time.Millisecond, // short timeout
		fakeCallError,
		checkWaitForData,
	)

	_, err := fut.Await(t.Context())

	if err == nil {
		t.Fatalf("Expected timeout error, got nil")
	}
}

func TestWaitFor_NilCall(t *testing.T) {
	fut := WaitFor(
		t.Context(),
		testRetries,
		testDelay,
		testTimeout,
		nil, // nil call
		checkAlwaysTrue,
	)

	_, err := fut.Await(t.Context())

	if err == nil {
		t.Fatalf("Expected error for nil call, got nil")
	}
}

func TestWaitFor_NilCheckUsesDefault(t *testing.T) {
	fut := WaitFor(
		t.Context(),
		testRetries,
		testDelay,
		testTimeout,
		fakeCallSuccess,
		nil, // nil check
	)

	resp, err := fut.Await(t.Context())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp.Data.Name != "ok" {
		t.Fatalf("Expected data name 'ok', got: %s", resp.Data.Name)
	}
}

func TestDefaultWaitFor_ReturnsErrorIfNoCall(t *testing.T) {
	fut := DefaultWaitFor[DummyData](
		t.Context(),
		nil, nil,
	)

	resp, err := fut.Await(t.Context())

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if resp != nil {
		t.Fatalf("Expected data name 'ok', got: %s", resp.Data.Name)
	}
}

func TestAsyncClient_MultipleAwait(t *testing.T) {
	fut := WaitFor(
		t.Context(),
		testRetries,
		testDelay,
		testTimeout,
		fakeCallSuccess,
		checkAlwaysTrue,
	)

	// First Await
	resp1, err1 := fut.Await(t.Context())
	// Second Await
	resp2, err2 := fut.Await(t.Context())
	if err1 != nil || err2 != nil {
		t.Fatalf("Expected no error, got: %v and %v", err1, err2)
	}
	if resp1.Data.Name != "ok" || resp2.Data.Name != "ok" {
		t.Fatalf("Expected data name 'ok', got: %s and %s", resp1.Data.Name, resp2.Data.Name)
	}
	if resp1 != resp2 {
		t.Fatalf("Expected same response instance, got different")
	}
}

func TestAsyncClient_IndependentClients(t *testing.T) {
	fut := WaitFor(
		t.Context(),
		testRetries,
		testDelay,
		testTimeout,
		fakeCallSuccess,
		checkAlwaysTrue,
	)

	fut2 := WaitFor(
		t.Context(),
		testRetries,
		testDelay,
		testTimeout,
		fakeCallError,
		checkAlwaysTrue,
	)

	// First Await
	resp1, err1 := fut.Await(t.Context())
	resp2, err2 := fut2.Await(t.Context())

	if err1 != nil {
		t.Fatalf("First Await expected no error, got: %v", err1)
	}
	if resp1.Data.Name != "ok" {
		t.Fatalf("First Await expected data name 'ok', got: %s", resp1.Data.Name)
	}

	// Second Await
	if err2 == nil {
		t.Fatalf("Second Await expected no error, got: %v", err2)
	}
	if resp2 != nil {
		t.Fatalf("Second Await expected data name 'ok', got: %s", resp2.Data.Name)
	}
}
