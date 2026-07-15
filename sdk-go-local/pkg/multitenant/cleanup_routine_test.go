package multitenant

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
)

// mockMultitenant is a mock implementation of Multitenant for testing
type mockMultitenant struct {
	cleanupCalls atomic.Int32
	lastFromDur  time.Duration
	lock         sync.Mutex
}

func (m *mockMultitenant) New(tenant string) error {
	return nil
}

func (m *mockMultitenant) NewFromOptions(tenant string, options *aruba.Options) error {
	return nil
}

func (m *mockMultitenant) Add(tenant string, client aruba.Client) {
}

func (m *mockMultitenant) Get(tenant string) (aruba.Client, bool) {
	return nil, false
}

func (m *mockMultitenant) MustGet(tenant string) aruba.Client {
	return nil
}

func (m *mockMultitenant) GetOrNil(tenant string) aruba.Client {
	return nil
}

func (m *mockMultitenant) CleanUp(from time.Duration) {
	m.cleanupCalls.Add(1)
	m.lock.Lock()
	defer m.lock.Unlock()
	m.lastFromDur = from
}

func TestStartCleanupRoutine_ExecutesOnTick(t *testing.T) {
	mock := &mockMultitenant{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickInterval := 50 * time.Millisecond
	fromDuration := 24 * time.Hour

	cleanupCancel := StartCleanupRoutine(ctx, mock, tickInterval, fromDuration)
	defer cleanupCancel()

	// Wait for at least 2 ticks to occur
	time.Sleep(150 * time.Millisecond)

	calls := mock.cleanupCalls.Load()
	if calls == 0 {
		t.Fatal("expected at least 1 cleanup call, got 0")
	}

	// Verify the fromDuration was passed correctly
	mock.lock.Lock()
	if mock.lastFromDur != fromDuration {
		t.Errorf("expected fromDuration %v, got %v", fromDuration, mock.lastFromDur)
	}
	mock.lock.Unlock()
}

func TestStartCleanupRoutine_StopsOnContextCancel(t *testing.T) {
	mock := &mockMultitenant{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickInterval := 50 * time.Millisecond
	cleanupCancel := StartCleanupRoutine(ctx, mock, tickInterval, 24*time.Hour)

	// Wait for some ticks
	time.Sleep(150 * time.Millisecond)
	initialCalls := mock.cleanupCalls.Load()

	// Cancel the routine
	cleanupCancel()

	// Let any in-flight cleanup complete
	time.Sleep(50 * time.Millisecond)

	// Continue waiting and verify no new cleanup calls occur
	time.Sleep(200 * time.Millisecond)
	finalCalls := mock.cleanupCalls.Load()

	if finalCalls > initialCalls+1 {
		t.Errorf("expected cleanup to stop after cancel, but got new calls: initial=%d, final=%d", initialCalls, finalCalls)
	}
}

func TestStartCleanupRoutine_UsesDefaultInterval(t *testing.T) {
	mock := &mockMultitenant{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Pass zero interval, should use default 1 hour
	cleanupCancel := StartCleanupRoutine(ctx, mock, 0, 24*time.Hour)
	defer cleanupCancel()

	// Give it a tiny window to start
	time.Sleep(10 * time.Millisecond)

	// The routine should be running but not ticking yet (1 hour interval)
	// Just verify no panic and routine is alive by canceling it
	cleanupCancel()
	time.Sleep(10 * time.Millisecond)
	// If we get here without panic, test passed
}

func TestStartCleanupRoutine_StopsOnParentContextCancel(t *testing.T) {
	mock := &mockMultitenant{}
	ctx, cancel := context.WithCancel(context.Background())

	tickInterval := 50 * time.Millisecond
	cleanupCancel := StartCleanupRoutine(ctx, mock, tickInterval, 24*time.Hour)
	defer cleanupCancel()

	// Wait for some ticks
	time.Sleep(150 * time.Millisecond)
	initialCalls := mock.cleanupCalls.Load()

	// Cancel parent context
	cancel()
	time.Sleep(100 * time.Millisecond)

	// Verify cleanup calls stopped
	finalCalls := mock.cleanupCalls.Load()
	if finalCalls > initialCalls+1 {
		t.Errorf("expected cleanup to stop after parent context cancel: initial=%d, final=%d", initialCalls, finalCalls)
	}
}

func TestStartCleanupRoutine_MultipleRoutines(t *testing.T) {
	mock1 := &mockMultitenant{}
	mock2 := &mockMultitenant{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickInterval := 50 * time.Millisecond

	cancel1 := StartCleanupRoutine(ctx, mock1, tickInterval, 24*time.Hour)
	cancel2 := StartCleanupRoutine(ctx, mock2, tickInterval, 12*time.Hour)
	defer cancel1()
	defer cancel2()

	time.Sleep(150 * time.Millisecond)

	calls1 := mock1.cleanupCalls.Load()
	calls2 := mock2.cleanupCalls.Load()

	if calls1 == 0 {
		t.Fatal("expected cleanup calls on mock1, got 0")
	}
	if calls2 == 0 {
		t.Fatal("expected cleanup calls on mock2, got 0")
	}

	// Verify different durations were passed
	mock1.lock.Lock()
	dur1 := mock1.lastFromDur
	mock1.lock.Unlock()

	mock2.lock.Lock()
	dur2 := mock2.lastFromDur
	mock2.lock.Unlock()

	if dur1 != 24*time.Hour {
		t.Errorf("expected mock1 duration 24h, got %v", dur1)
	}
	if dur2 != 12*time.Hour {
		t.Errorf("expected mock2 duration 12h, got %v", dur2)
	}
}
