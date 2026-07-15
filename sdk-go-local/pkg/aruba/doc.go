// Package aruba is the public entry point for the Aruba Cloud Go SDK.
//
// # Overview
//
// Importing only this package covers ~99.9% of real-world use cases:
//
//	import "github.com/Arubacloud/sdk-go/pkg/aruba"
//
// Construct a client with your OAuth2 credentials:
//
//	client, err := aruba.NewClient(aruba.DefaultOptions(clientID, clientSecret))
//
// Then reach any resource through the ten service-group accessors:
// [Client.FromCompute], [Client.FromNetwork], [Client.FromStorage],
// [Client.FromProject], [Client.FromDatabase], [Client.FromContainer],
// [Client.FromAudit], [Client.FromMetric], [Client.FromSchedule],
// [Client.FromSecurity].
//
// # Wrapper layer
//
// Every resource follows the triplet pattern defined in each resource_<name>.go:
//   - Wrapper  — chainable fluent builder constructed via NewXxx().
//   - Low-level client interface — contract the adapter depends on (testable in isolation).
//   - Adapter  — bridges the wrapper to internal/clients/<domain>.
//
// Wrappers are divided into two wire-shape families:
//   - Family A: Metadata{Properties{…}} envelope, regional, embeds statusMixin.
//     Most resources (CloudServer, VPC, KaaS, DBaaS, BlockStorage, Job, KMS, …).
//   - Family B: flat request body, no metadata envelope, no tags, no statusMixin.
//     Resources: Database, Key, Kmip, User, Grant.
//
// # Single-import principle
//
// All typed enum constants (Region, Zone, BillingPeriod, CloudServerFlavor, …)
// are re-exported from [pkg/aruba/aliases.go] so callers never need to import
// pkg/types for day-to-day use. Wrappers expose Raw(), RawJSON(), and RawYAML()
// for serialisation without a second import.
//
// The residual cases that require pkg/types or pkg/async — non-promoted wire
// fields, structured validation errors, and concurrent background polling — are
// documented in docs/website/docs/working-at-low-level.md.
//
// # Error accumulation
//
// Setter-time errors are deferred into an internal errMixin and surfaced as a
// single joined error when Create or Update is called. This means a builder
// chain never panics or short-circuits; call wrapper.Err() to inspect
// accumulated errors before submitting.
//
// # Async polling
//
// Resources that embed statusMixin gain WaitUntilActive, WaitUntilReady, and
// WaitUntilStates(ctx, targets, opts…) for free. The underlying polling is
// driven by pkg/async.WaitFor with defaults DefaultRetries=60,
// DefaultBaseDelay=10s, DefaultTimeout=600s (overridable via WaitOption helpers).
//
// See ai/ARCHITECTURE.md and ai/CONVENTIONS.md for the full design reference.
package aruba
