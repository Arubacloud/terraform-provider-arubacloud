# Plan: Tech Debt Remediation (TD-001 → TD-029)

Tackle all 29 tech debt items across 5 phases. TD-003 (ImportState) is intentionally skipped per user decision.

---

## Phase 1 — Critical Fixes ✅ DONE

### ~~TD-002: Fix Provider Registry Address~~ ✅
- `main.go:31`: changed `registry.terraform.io/hashicorp/arubacloud` → `registry.terraform.io/arubacloud/arubacloud`

### ~~TD-001 (partial): Fix 3 Representative Data Sources~~ ✅
- `keypair_data_source.go`, `elasticip_data_source.go`, `backup_data_source.go`: real SDK calls, 404 handling, field mapping.
- Done in commit `2b11778`.

---

## Phase 2 — High Bugs: Shared Infrastructure

### ~~TD-004 + TD-019: Remove Dead Code from resource_wait.go~~ ✅
- Deleted `IsDependencyError`, `containsDependencyKeywords`, `RetryDeleteOperation`, unused `strings` import.

### ~~TD-005 + TD-018: Replace Reflection-Based Error Extraction with Typed Generics~~ ✅
- New `internal/provider/provider_error.go`:
  - `ProviderError` struct (Category, StatusCode, Title, Detail, Instance, Operation, Resource, Cause)
  - `CheckResponse[T any]` generic — replaces all reflection; returns `*ProviderError` or nil
  - `NewTransportError` — wraps network/SDK errors
  - `IsNotFound`, `ErrorIsSemantic`, `ErrorIsTransient`, `ErrorIsTechnical` — typed helpers via `errors.As`
  - Validation errors inlined into `Detail`; `Error()` is pure (no side effects)
- `error_helper.go`: `FormatAPIError` deleted; file reduced to `package provider`
- `resource_wait.go`: `ExtractSDKError` (reflection) deleted; `DeleteResourceWithRetry` now accepts `func() error` — caller wraps SDK call + `CheckResponse`/`NewTransportError`
- All 27 resource files + all data source files migrated to `CheckResponse` + `NewTransportError`
- `vpc_resource.go`: deleted `extractVPCDelError` + `extractErrorFunc` (TD-018); `reflect` import removed
- `go build ./...` and `go vet ./...` both pass.

### ~~TD-007: Fix Polling Error Swallowing~~ ✅
- `resource_wait.go` — `WaitForResourceActive`: added `consecutiveErrors` counter; after 3 consecutive checker errors returns a wrapped error with root cause instead of polling silently until timeout.

### TD-006: Fix CloudServer Update Ignoring Fields
- File: `cloudserver_resource.go` — Schema() + Update()
- Fields unsupported by the API update endpoint (VPC, subnets, security groups, elastic IP, zone, user data): add `RequiresReplace()` plan modifiers so Terraform destroys+recreates instead of silently ignoring changes

### TD-008: Fix CloudServer Read Configuration Drift
- File: `cloudserver_resource.go` — Read() L552–L607
- Map `network`, `settings.key_pair_uri_ref`, `settings.user_data`, `storage` from API response where available
- For write-only/input-only fields: add comment explaining state preservation

---

## Phase 3 — Medium Code Quality

### TD-013: Fix Wrong Error Message in Configure() (29 files)
- All `*_resource.go` + `*_data_source.go` files
- Replace `"Expected *http.Client, got: %T."` → `"Expected *ArubaCloudClient, got: %T."`

### TD-014: Restore Schema Validators (6 files)
- `elasticip_resource.go`: `stringvalidator.OneOf` on `location`, `billing_period`
- `securitygroup_resource.go`, `securityrule_resource.go`, `snapshot_resource.go`, `subnet_resource.go`, `vpc_resource.go`: restore validators on constrained string fields
- Add import `"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"`

### TD-015: Add RequiresReplace Plan Modifiers
- All resource Schema() methods for immutable fields: `location`, `project_id`, `zone`, `type`, `volume_id`
- Add `PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}`

### TD-016: Error Wrapping with %w
- All Go files in `internal/provider/`
- Replace `fmt.Errorf("...: %v", err)` / `fmt.Errorf("...: %s", err)` → `fmt.Errorf("...: %w", err)` where forwarding upstream errors

### TD-017: Extract Tag Helpers
- New file: `internal/provider/helpers.go`
- Add `TagsToList(ctx, []string) (types.List, diag.Diagnostics)` and `ListToTags(ctx, types.List) ([]string, diag.Diagnostics)`
- Replace inline tag conversion code in all resource/datasource files (~28 occurrences)

### TD-020: Fix parseTimeout Silent Failure
- File: `provider.go` — `parseTimeout`
- Change signature to accept `diag.Diagnostics`; add `diags.AddWarning(...)` when parse fails, then return default

### TD-021: Fix Inconsistent Empty Tags in keypair_resource.go
- File: `keypair_resource.go` L343
- Change `types.ListNull(types.StringType)` → `types.ListValue(types.StringType, []attr.Value{})` (or use TD-017 helper)

### TD-028: Fix Wrong MarkdownDescription on ElasticIP
- File: `elasticip_resource.go` L46
- Change `"Project resource"` → `"Elastic IP resource"`

---

## Phase 4 — Testing

### TD-012: Unit Tests for Helper Functions
- New files: `resource_wait_test.go`, `provider_error_test.go`
- Table-driven tests for `WaitForResourceActive`, `DeleteResourceWithRetry`, `CheckResponse`, `NewTransportError`, `IsNotFound`
- Mock checker functions

### TD-009: Add CheckDestroy to Acceptance Tests
- All `*_resource_test.go` files
- Add `CheckDestroy` function calling the API and asserting 404

### TD-010: Add Error-Case Acceptance Tests
- Add `resource.TestStep` entries with `ExpectError` for invalid config, missing fields, etc.

### TD-011: Rewrite Meaningless Data Source Tests
- After TD-001 fixes, rewrite `keypair_data_source_test.go`, `elasticip_data_source_test.go`, `backup_data_source_test.go` to test real API responses

---

## Phase 5 — Project / Build

### TD-022: Fix go.mod Module Path
- `go.mod` line 1: `module terraform-provider-arubacloud` → `module github.com/Arubacloud/terraform-provider-arubacloud`
- `main.go`: update import path accordingly

### TD-023: Fix Copyright Headers
- `main.go` L1: `// Copyright (c) HashiCorp, Inc.` → `// Copyright (c) Aruba S.p.A.`
- `.golangci.yml` L1: same

### TD-024 + TD-025: Fix .gitignore
- Add `coverage.html`, `coverage.out`
- Add subdirectory globs: `**/.terraform/`, `**/*.tfstate`, `**/*.tfstate.backup`, `**/terraform.tfvars`

### TD-026: Pin GitHub Actions in release.yml
- Pin all 4 actions to SHA hashes matching the versions already used

### TD-027: Add govet to .golangci.yml
- Add `govet` to `linters.enable` list

### TD-029: Resolve Commented-Out KMIP/Key
- Assess whether implementation is complete; if not, delete implementation files and remove commented lines from `provider.go`

---

## Relevant Files

- `main.go` — TD-002 ✅, TD-022, TD-023
- `internal/provider/provider_error.go` (new) — TD-005 ✅, TD-018 ✅
- `internal/provider/error_helper.go` — TD-005 ✅ (gutted)
- `internal/provider/resource_wait.go` — TD-004 ✅, TD-005 ✅, TD-007, TD-019 ✅
- `internal/provider/vpc_resource.go` — TD-018 ✅
- `internal/provider/cloudserver_resource.go` — TD-006, TD-008
- `internal/provider/helpers.go` (new) — TD-017
- `internal/provider/provider.go` — TD-020, TD-029
- `internal/provider/keypair_resource.go` — TD-021
- `internal/provider/elasticip_resource.go` — TD-028
- All `*_resource.go` + `*_data_source.go` — TD-005 ✅, TD-013, TD-015, TD-016
- `elasticip_resource.go`, `securitygroup_resource.go`, `securityrule_resource.go`, `snapshot_resource.go`, `subnet_resource.go`, `vpc_resource.go` — TD-014
- `.gitignore` — TD-024, TD-025
- `.golangci.yml` — TD-023, TD-027
- `.github/workflows/release.yml` — TD-026
- `go.mod` — TD-022

---

## Verification

1. `go build ./...` — passes ✅
2. `go vet ./...` — passes ✅
3. `golangci-lint run` — must pass with no new errors after TD-013/TD-014/TD-015/TD-016
4. `go test ./internal/provider/... -run TestUnit` — unit tests for TD-012
5. Manual: acceptance tests for the 3 fixed data sources (TD-001)
6. Manual: verify `terraform plan` produces `RequiresReplace` annotations for immutable fields (TD-015)

---

## Decisions

- TD-003 (ImportState): Skipped intentionally
- TD-001: Only 3 representative data sources fixed; remainder tracked as follow-up
- TD-005/TD-018: Adopted operator-style `ProviderError` + `CheckResponse[T]` pattern instead of patching the reflection code
- TD-022 module rename touches only `go.mod` + `main.go` (only 1 internal import)
- TD-017 helpers.go should be created before TD-021 to consolidate work
