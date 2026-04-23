# TECH_DEBT.md — Known Issues, Bugs, and Refactoring Backlog

Items are ordered by severity. Each entry includes affected files/lines and a recommended fix.

---

## Critical

### TD-001 — 22/27 Data Sources Return Hardcoded Fake Data

**Affected files:** All `*_data_source.go` files **except** `kaas_data_source.go`, `kmip_data_source.go`, `key_data_source.go`.

**Issue:** The `Read()` method of 22 data sources never calls the SDK. Instead it populates the state with hardcoded example strings and URIs (e.g. `cloudserver_data_source.go:143` sets `data.Uri = "/v2/cloudservers/68398923fb2cb026400d4d31"`). Any `data.*` block referencing these data sources silently returns wrong values, making them completely unusable and dangerous for production use.

**Fix:** Implement each data source's `Read()` to call the appropriate SDK `Get()` or `List()` method using the input ID, handle 404 by removing the resource from state, and map the response fields into the model. Follow the pattern in `kaas_data_source.go`.

---

### TD-002 — Provider Registry Address Is Wrong

**Affected file:** `main.go:31`

**Issue:**
```go
Address: "registry.terraform.io/hashicorp/arubacloud",
```
This uses the `hashicorp` namespace, which is reserved for HashiCorp-owned providers. The correct address is `registry.terraform.io/arubacloud/arubacloud`. The mismatch between the binary address and the published registry source breaks provider installation when the user's `.tf` references `source = "arubacloud/arubacloud"`.

**Fix:** Change the address to `registry.terraform.io/arubacloud/arubacloud`.

---

## High — Bugs

### TD-003 — ImportState Broken for All Multi-Key Resources

**Affected resources (~12):** `backup`, `blockstorage`, `cloudserver`, `database`, `databasegrant`, `databasebackup`, `dbaasuser`, `elasticip`, `keypair`, `restore`, `schedulejob`, `securityrule`, `subnet`, `vpc`, and others.

**Issue:** Every resource uses `resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)`, which only sets the `id` field. However, their `Read()` methods require additional fields (`project_id`, `dbaas_id`, `vpc_id`, etc.) that are left as empty strings after import. The subsequent Read triggered by Terraform post-import will fail with "Missing Required Fields".

Example: `backup_resource.go:272-273` reads `projectID := data.ProjectID.ValueString()` and exits with an error if it is empty.

**Fix:** Implement a custom `ImportState()` per resource that parses a composite import ID (e.g. `project_id/resource_id` or `project_id/dbaas_id/resource_id`) using `strings.SplitN`, sets all required path roots via `resp.State.SetAttribute()`, and documents the expected format in the schema's `MarkdownDescription`.

---

### TD-004 — `IsDependencyError` Logic Is Broken (Dead Code Branch)

**Affected file:** `resource_wait.go:78-108`

**Issue:** The function returns `true` unconditionally for any non-404 status code. The `containsDependencyKeywords()` check at line 102 is dead code — even if it returns `false`, line 108 returns `true` anyway. The function name implies selective behaviour that does not exist. Additionally, neither `IsDependencyError` nor `containsDependencyKeywords` is called anywhere in the codebase; both are unreachable dead code.

**Fix:** Delete `IsDependencyError` (lines 75-109) and `containsDependencyKeywords` (lines 111-143). If keyword-based distinction is ever needed, add it where the retry decision is actually made in `DeleteResourceWithRetry`.

---

### TD-005 — `ExtractSDKError` Can Panic (Reflection `IsNil` on Non-Pointer Field)

**Affected files:** `resource_wait.go:319-335`, `vpc_resource.go:584-591` (duplicate)

**Issue:** The code calls `titleField.IsNil()` via reflection. If the `Title` field in the SDK error response is a plain `string` (not `*string`), `reflect.Value.IsNil()` panics with "reflect: call of reflect.Value.IsNil on string Value". The same applies to `detailField.IsNil()`.

```go
// resource_wait.go:323
if titleField.IsValid() && titleField.CanInterface() && !titleField.IsNil()  // panics if Title is string
```

**Fix:** Check `titleField.Kind() == reflect.Ptr` before calling `IsNil()`. Use:
```go
if titleField.IsValid() && titleField.Kind() == reflect.Ptr && !titleField.IsNil() {
```

---

### TD-006 — CloudServer Update Silently Ignores Network / Security / Zone Changes

**Affected file:** `cloudserver_resource.go:696-710` (Update method)

**Issue:** The update request only includes `Name`, `Tags`, `Location`, `FlavorName`, `BootVolume`, and `KeyPair`. Fields like VPC, subnets, security groups, elastic IP, zone, and user data are never included. If a user changes these in their `.tf` config, Terraform plans a diff, the Update runs without error, but the API is never told about the change. The state is then saved with the new plan values (lines 768-824), creating a false sense of success.

**Fix:** Either add the missing fields to the update request (if the API supports it), or add `RequiresReplace()` plan modifiers to those fields so Terraform destroys and recreates the resource instead.

---

### TD-007 — Polling Swallows Checker Errors Indefinitely

**Affected file:** `resource_wait.go:36-38`

**Issue:**
```go
state, err := checker(ctx)
if err != nil {
    tflog.Warn(ctx, fmt.Sprintf("Error checking %s %s status: %v", ...))
    continue
}
```
When the status checker returns an error (e.g. auth failure, network error), the loop logs a warning and retries until the full `ResourceTimeout` (default 10 minutes) expires. The final error reported is a generic "timeout" that hides the real cause.

**Fix:** Implement a consecutive-error limit (e.g. 3 retries). After exceeding it, return an error that wraps the last checker error:
```go
return fmt.Errorf("status check failed for %s %s after %d consecutive errors: %w", resourceType, resourceID, maxErrors, lastErr)
```

---

### TD-008 — CloudServer Read Does Not Detect Configuration Drift

**Affected file:** `cloudserver_resource.go:552-607` (Read method)

**Issue:** The Read method copies `network`, `settings.key_pair_uri_ref`, `settings.user_data`, and `storage` directly from the previous Terraform state instead of reading them from the API response. Changes made outside Terraform (drift) are never detected for these fields.

**Fix:** Map all available fields from the API response. For fields the API does not return (write-only or input-only), document this explicitly and preserve state only for those specific fields with a comment explaining why.

---

## High — Testing

### TD-009 — No `CheckDestroy` in Any Acceptance Test

**Affected files:** All `*_resource_test.go` files.

**Issue:** No test verifies that resources are actually deleted after the test completes. Leaked cloud resources accumulate cost and can cause quota issues.

**Fix:** Add a `CheckDestroy` function to each `resource.TestCase` that calls the API and asserts a 404 response.

---

### TD-010 — No Error-Case Acceptance Tests

**Affected files:** All `*_resource_test.go` files.

**Issue:** Every `TestAcc*` only tests the happy path. Missing coverage: invalid configuration (wrong field values), missing required fields, 404 during Read (resource removed outside Terraform), concurrent modification.

**Fix:** Add `resource.TestStep` entries that use `ExpectError` with a regex matching the expected diagnostic, or add separate test functions for error scenarios.

---

### TD-011 — Data Source Tests Are Meaningless (See TD-001)

**Affected files:** All `*_data_source_test.go` files.

**Issue:** Since data sources return hardcoded values (TD-001), their tests only verify that hardcoded strings are non-null. They test nothing real.

**Fix:** Fix TD-001 first, then rewrite the tests to verify real API responses.

---

### TD-012 — No Unit Tests for Helper Functions

**Affected files:** `resource_wait.go`, `error_helper.go`

**Issue:** Complex logic like `WaitForResourceActive`, `DeleteResourceWithRetry`, `ExtractSDKError`, and `FormatAPIError` has zero unit test coverage. Bugs in these helpers affect all 27 resources.

**Fix:** Add `resource_wait_test.go` and `error_helper_test.go` with table-driven tests. Use mock checker functions and mock SDK responses.

---

## Medium — Code Quality

### TD-013 — 29 `Configure()` Methods Have Wrong Error Message

**Affected files:** ~29 files including `backup_data_source.go:92`, `blockstorage_data_source.go:118`, `blockstorage_resource.go:116`, `elasticip_resource.go:108`, `securitygroup_resource.go:98`, `securityrule_resource.go:249`, `snapshot_resource.go:98`, `subnet_resource.go:179`, and others.

**Issue:**
```go
fmt.Sprintf("Expected *http.Client, got: %T.", req.ProviderData)
```
The expected type is `*ArubaCloudClient`, not `*http.Client`. This is a copy-paste artifact from scaffolding. When this error fires, it actively misleads debugging.

**Fix:** Replace `*http.Client` with `*ArubaCloudClient` in all affected files. Consider extracting `Configure()` into a shared helper to prevent this class of copy-paste bug permanently.

---

### TD-014 — All Schema Validators Removed

**Affected files:** `elasticip_resource.go:55,69`, `snapshot_resource.go:67,72`, `securityrule_resource.go:172,204,209,227`, `securitygroup_resource.go:68`, `subnet_resource.go:92,113`, `vpc_resource.go:64` (12 locations total).

**Issue:** Comments read `// Validators removed for v1.16.1 compatibility`. Fields like `location`, `billing_period`, `direction`, `protocol`, `type` accept any string. Errors surface only at API call time with cryptic messages.

**Fix:** Restore validators using `stringvalidator.OneOf(...)` from `github.com/hashicorp/terraform-plugin-framework-validators`. Verify the compatibility concern with the current `terraform-plugin-framework v1.16.1` — the validators package is separate and unaffected by the core version.

---

### TD-015 — No `RequiresReplace` Plan Modifiers on Immutable Fields

**Affected files:** All resource `Schema()` methods.

**Issue:** Fields that are immutable after creation (e.g. `location`, `project_id`, `zone`, `type`, `volume_id`) have no `stringplanmodifier.RequiresReplace()`. Terraform will attempt in-place updates that either silently do nothing (TD-006) or fail at the API level. The correct pattern is to signal at plan time that recreation is required.

**Fix:** Add `PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}` to all schema attributes that are set once at creation and cannot be changed.

---

### TD-016 — No `%w` Error Wrapping Anywhere

**Affected files:** All Go files in `internal/provider/`.

**Issue:** All `fmt.Errorf()` calls use `%s` or `%v` to format errors. This discards the error chain and prevents callers from using `errors.Is()` or `errors.As()`. Example: `resource_wait.go:29` formats `ctx.Err()` as a string instead of wrapping it.

**Fix:** Replace `fmt.Errorf("...: %v", err)` with `fmt.Errorf("...: %w", err)` for all cases where an upstream error is being forwarded.

---

### TD-017 — Massive Tag-Handling Code Duplication

**Affected files:** All `*_resource.go` and `*_data_source.go` files (28+ instances).

**Issue:** The pattern for converting `[]string` tags to `types.List` and back is repeated identically in every resource's Create/Read/Update. Any change or bug fix must be applied in 28+ places.

**Fix:** Extract into shared helpers in a new `helpers.go` file:
```go
func TagsToList(ctx context.Context, tags []string) (types.List, diag.Diagnostics)
func ListToTags(ctx context.Context, list types.List) ([]string, diag.Diagnostics)
```

---

### TD-018 — Duplicate Reflection Error Extractor in `vpc_resource.go`

**Affected file:** `vpc_resource.go:504-598` (`extractVPCDelError`)

**Issue:** `extractVPCDelError` is nearly identical to `ExtractSDKError` in `resource_wait.go`. The VPC resource duplicates this reflection logic rather than reusing the shared function. Any bug fix (e.g. TD-005) must be applied in both places.

**Fix:** Delete `extractVPCDelError` and the associated `extractErrorFunc` variable. Call `ExtractSDKError` directly in the VPC delete method.

---

### TD-019 — Dead Code in `resource_wait.go`

**Affected file:** `resource_wait.go`

**Issue:**
- `RetryDeleteOperation` (lines 352-360): trivial wrapper around `DeleteResourceWithRetry` with no additional logic. Adds an indirection layer with zero benefit.
- `IsDependencyError` (lines 75-109): never called anywhere. (See TD-004.)
- `containsDependencyKeywords` (lines 111-143): only called by `IsDependencyError`. (See TD-004.)

**Fix:** Delete all three functions. If `RetryDeleteOperation` is part of a public API surface used by tests, inline the single call.

---

### TD-020 — `parseTimeout` Silently Swallows Invalid Configuration

**Affected file:** `provider.go:164-176`

**Issue:** If the user sets `resource_timeout = "banana"` in their provider block, `time.ParseDuration` fails and the function silently returns the 10-minute default. The user gets no feedback that their configuration was ignored.

**Fix:** Accept a `*diag.Diagnostics` parameter and emit a `diags.AddWarning(...)` on parse failure before returning the default.

---

### TD-021 — Inconsistent Empty Tags Representation

**Affected file:** `keypair_resource.go:345`

**Issue:** `keypair_resource.go` sets empty tags to `types.ListNull(types.StringType)` while every other resource uses `types.ListValue(types.StringType, []attr.Value{})`. This inconsistency causes spurious plan diffs when Terraform compares a null list against an empty list.

**Fix:** Standardize on `types.ListValue(types.StringType, []attr.Value{})` for empty tags in `keypair_resource.go`. Or better, fix TD-017 first and use the shared helper everywhere.

---

## Medium — Project / Build

### TD-022 — `go.mod` Uses Bare Module Path

**Affected file:** `go.mod:1`

**Issue:** `module terraform-provider-arubacloud` is a bare name, not a fully-qualified module path. The Go convention is `github.com/Arubacloud/terraform-provider-arubacloud`. This affects `go get` usability and all internal import paths.

**Fix:** Update the module declaration and all `import "terraform-provider-arubacloud/..."` paths throughout the codebase. This is a broad find-replace but straightforward.

---

### TD-023 — Wrong Copyright Headers

**Affected files:** `main.go:1`, `.golangci.yml:1`

**Issue:** Both files carry `// Copyright (c) HashiCorp, Inc.` — a leftover from the HashiCorp provider template. `tools/tools.go` correctly says "Aruba S.p.A.".

**Fix:** Update to reflect the actual copyright holder.

---

### TD-024 — Build Artifacts Committed to Git

**Affected files:** `coverage.html`, `coverage.out` (root directory)

**Issue:** These are test output artifacts that should not be tracked in version control.

**Fix:** Add `coverage.html` and `coverage.out` to `.gitignore` and remove them from the repository.

---

### TD-025 — Test Temporary Files Leaked into `examples/test/`

**Affected paths:** `examples/test/` subdirectories

**Issue:** `.terraform/` directories, `.tfstate` files, `.tfvars` files, and `.terraform.lock.hcl` files are present under `examples/test/`. The root `.gitignore` patterns are root-relative and do not cover subdirectories.

**Fix:** Add glob patterns to `.gitignore`:
```
**/.terraform/
**/*.tfstate
**/*.tfstate.backup
**/terraform.tfvars
```

---

### TD-026 — Release Workflow Uses Unpinned Action Tags

**Affected file:** `.github/workflows/release.yml`

**Issue:** Actions are referenced as `@v4` / `@v5` tags instead of pinned SHA hashes. This is inconsistent with `test.yml` which does pin SHAs. Unpinned tags are a supply-chain attack vector.

**Fix:** Pin each action to a specific commit SHA, matching the pattern used in `test.yml`.

---

### TD-027 — `govet` Linter Not Enabled

**Affected file:** `.golangci.yml`

**Issue:** `govet` is the baseline Go static analysis tool and is conspicuously absent from the enabled linters list. It catches common bugs like misused `sync.Mutex`, incorrect `Printf` format strings, and structtag issues.

**Fix:** Add `govet` to the `linters.enable` list in `.golangci.yml`.

---

## Low

### TD-028 — Wrong `MarkdownDescription` on ElasticIP Resource

**Affected file:** `elasticip_resource.go:46`

**Issue:** `MarkdownDescription: "Project resource"` was copy-pasted from `project_resource.go`.

**Fix:** Change to `"Elastic IP resource"`.

---

### TD-029 — Commented-Out KMIP/Key Resources Without Cleanup

**Affected file:** `provider.go:214-216, 246-248`

**Issue:** `NewKMIPResource`, `NewKeyResource`, and their data source equivalents are commented out with `// Temporarily disabled due to issues`. The full implementation files still exist. This creates confusion about what is supported.

**Fix:** Either complete the implementation and re-enable them, or delete the implementation files and remove the commented lines from `provider.go`.

---

### TD-030 — `int64` to `int` Silent Truncation

**Affected files:** `backup_resource.go:187`, `blockstorage_resource.go:172`, `subnet_resource.go:288`

**Issue:** `int(data.X.ValueInt64())` silently truncates on 32-bit platforms. While unlikely in practice (providers run on 64-bit), it violates correctness.

**Fix:** Use explicit `int64` in SDK calls where possible. Where `int` is required by the SDK, add a bounds check or a `//nolint:gosec` comment with an explanation.

---

### TD-031 — Nil Pointer Risk in CloudServer Read/Update for Optional API Fields

**Affected file:** `cloudserver_resource.go`

**Issue:**
- Read (line 525): if `server.Metadata.LocationResponse` is nil, `data.Location` is left as a zero `types.String{}` (neither null nor valid). Should be `types.StringNull()`.
- Update (line 703): accesses `current.Metadata.LocationResponse.Value` without a nil check, while Read (line 525) guards against this being nil. This is a potential nil pointer dereference.

**Fix:** Add nil guards in both Read and Update for `LocationResponse` (and any other pointer fields). In Read, set fields to `types.StringNull()` in the `else` branch when the API returns nil.

---

### TD-032 — `strings.Split` URI Parsing Always Succeeds on Empty String

**Affected file:** `backup_resource.go:328-331`

**Issue:**
```go
parts := strings.Split(backup.Properties.Origin.URI, "/")
if len(parts) > 0 {  // always true, even for ""
    data.VolumeID = types.StringValue(parts[len(parts)-1])
}
```
`strings.Split("", "/")` returns `[""]`, so `parts[len(parts)-1]` is `""`. An empty `VolumeID` is silently stored.

**Fix:** Change the guard to `len(parts) > 1 && parts[len(parts)-1] != ""`, or verify the URI is non-empty before splitting.

---

### TD-033 — Missing Community Files

**Issue:** No `CONTRIBUTING.md`, `SECURITY.md`, or `.editorconfig`.

**Fix:** Add:
- `CONTRIBUTING.md`: build steps, test instructions, PR guidelines
- `SECURITY.md`: vulnerability reporting process
- `.editorconfig`: consistent indentation (tabs for Go, 2 spaces for HCL)

---

### TD-034 — GoReleaser Config Uses Deprecated `version: 1`

**Affected file:** `.goreleaser.yml:1`

**Issue:** GoReleaser v2 uses `version: 2` and has breaking changes in schema. The current config on `version: 1` will eventually stop working as GoReleaser drops v1 support.

**Fix:** Migrate to `version: 2` following the GoReleaser migration guide.

---

## Observability

### TD-035 — SDK HTTP Logging Is Permanently Disabled; No Provider Log-Level Control

**Affected file:** `internal/provider/provider.go:127-128`

**Issue:** The provider hard-codes `options.WithDefaultLogger()` when building the SDK client. Despite the name, `WithDefaultLogger()` maps to `LoggerNoLog` in SDK v0.1.24 (`pkg/aruba/options.go:524-528`). This means the SDK never emits any HTTP request/response detail — method, URL, headers, body, response status — regardless of the `TF_LOG` setting. Operators troubleshooting failed provisioning have zero visibility into what the SDK is actually sending to the API.

The SDK already implements full per-call `Debugf` logging in `internal/restclient/client.go` (method + URL, query params, headers with `Authorization: Bearer [REDACTED]`, request body, response status/headers/body). It also exposes `WithCustomLogger(logger.Logger)` (`options.go:689-693`) for injecting a custom logger that satisfies a 4-method interface (`Debugf/Infof/Warnf/Errorf`). This capability is simply unused.

**Fix:** Implement a thin `sdkLogAdapter` struct in `internal/provider/sdk_logger.go` that:
1. Satisfies the SDK's `logger.Logger` interface (4 methods — no import of the internal package needed due to Go structural typing).
2. Holds a `LogLevel` field (Off/Error/Warn/Info/Debug/Trace) that gates each method before forwarding via `tflog.SubsystemDebug/Info/Warn/Error(ctx, "arubacloud-sdk", msg)`.
3. Registers a `tflog` subsystem (`ctx = tflog.NewSubsystem(ctx, "arubacloud-sdk")`) so output is filterable independently from provider-level logs via `TF_LOG_PROVIDER_ARUBACLOUD_SDK`.

Expose the level as a new optional provider attribute `log_level` (+ env var `ARUBACLOUD_LOG_LEVEL`) with default `INFO`. Valid values (case-insensitive): `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `OFF`. Invalid values emit a Warning diagnostic and fall back to `INFO`.

Replace `options.WithDefaultLogger()` with `options.WithCustomLogger(newSDKLogAdapter(ctx, logLevel))`.

Add `ArubaCloudProviderModel.LogLevel types.String \`tfsdk:"log_level"\`` and corresponding schema attribute with a `MarkdownDescription` that explains the two-filter model (`log_level` vs `TF_LOG`).

Add unit tests for `ParseLogLevel` and level-gating in `internal/provider/sdk_logger_test.go`. Regenerate docs via `make generate`.
