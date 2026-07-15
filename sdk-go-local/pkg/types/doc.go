// Package types defines all wire-level data-transfer objects (DTOs) used by the
// Aruba Cloud SDK — both outbound request bodies and inbound response bodies.
//
// # Naming convention
//
// Every exported struct in this package carries a suffix that declares its wire role:
//
//   - *Request  — serialised into an outbound HTTP request body only.
//   - *Response — deserialised from an inbound HTTP response body only.
//   - *Common   — appears on both the request AND the response side of the wire
//     (i.e. the same JSON shape is sent in a request body and received back in a
//     response body). Use *Common for shared nested shapes such as billing plans,
//     linked-resource references, or VPN settings structs.
//
// The following categories are intentionally out of scope for the three-way rule and
// carry no Request/Response/Common suffix:
//
//   - Scalar enum / constant types (State, Region, Zone, BillingPeriod, RuleProtocol,
//     KeyAlgorithm, …) — their role is determined by the parent field, not the type.
//   - The generic transport envelope Response[T] and its embedded ListResponse.
//   - HTTP-utility types (RequestParameters, AcceptHeader).
//   - Error types (ErrorResponse, ValidationError, MetadataValidationError).
//   - The Resource interface.
//   - Custom (un)marshaler helpers (SubnetCIDROrRef).
//
// # File organisation
//
// Each file groups the DTOs for a single resource domain:
//
//	<domain>.<resource>.go   e.g. compute.cloudserver.go, network.vpc.go
//
// Cross-resource base types (metadata envelopes, status, billing, linked-resource
// references) live in resource.go.
package types
