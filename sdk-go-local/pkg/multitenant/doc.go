// Package multitenant provides a per-tenant aruba.Client registry with
// automatic idle-eviction.
//
// # Overview
//
// Create a manager and register clients by tenant ID:
//
//	mt := multitenant.New()
//	mt.NewFromOptions("tenant-a", aruba.DefaultOptions(idA, secretA))
//	mt.NewFromOptions("tenant-b", aruba.DefaultOptions(idB, secretB))
//
//	client, ok := mt.Get("tenant-a")
//
// # Constructors
//
//   - [New] — empty manager; tenants must be added via NewFromOptions or Add.
//   - [NewWithTemplate] — every New("tenant") call deep-copies the template
//     Options; slices are deep-copied, *http.Client / logger / middleware are
//     shallow-copied as shared singletons.
//
// # Access methods
//
// All three access methods ([Multitenant.Get], [Multitenant.MustGet],
// [Multitenant.GetOrNil]) update the entry's lastUsage timestamp on each call.
// This timestamp drives idle-eviction via [Multitenant.CleanUp].
//
// # Idle eviction
//
// Call [StartCleanupRoutine] to start a background goroutine that periodically
// removes entries idle longer than `fromDuration`:
//
//	cancel := multitenant.StartCleanupRoutine(ctx, mt, 5*time.Minute, 24*time.Hour)
//	defer cancel()
//
// Default recommended values: tick every hour, evict after 24 hours.
//
// Note: the lastUsage field is written under a read lock (TD-003 in
// ai/TECH_DEBT.md) — use a single-goroutine access pattern or apply the fix
// before production use under high concurrency.
package multitenant
