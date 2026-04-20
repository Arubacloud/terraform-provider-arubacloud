# GitHub Copilot Instructions

This is a **Terraform provider** for ArubaCloud built with [HashiCorp Terraform Plugin Framework v1](https://github.com/hashicorp/terraform-plugin-framework) (Protocol v6).

For deeper reference, load the relevant file from `ai/` based on your task:

| Task | File |
|------|------|
| Builds, tests, docs generation | [`ai/DEVEX.md`](../ai/DEVEX.md) |
| Repo layout / what lives where | [`ai/REPO.md`](../ai/REPO.md) |
| Implementing or reviewing features | [`ai/ARCHITECTURE.md`](../ai/ARCHITECTURE.md) |
| Matching existing code style | [`ai/CONVENTIONS.md`](../ai/CONVENTIONS.md) |
| Tech debt / refactoring backlog | [`ai/TECH_DEBT.md`](../ai/TECH_DEBT.md) |

---

## Quick Reference

### Stack

- **Language**: Go 1.22+
- **Framework**: `terraform-plugin-framework` v1 (Protocol v6)
- **SDK**: `github.com/Arubacloud/sdk-go`
- **All provider logic**: `internal/provider/`
- **Entry point**: `main.go` → `internal/provider/provider.go`

### Key Shared Utilities

- `internal/provider/resource_wait.go` — polling helper `WaitForResourceActive()`
- `internal/provider/provider_error.go` — typed error handling: `CheckResponse[T]`, `NewTransportError`, `ProviderError`
- `internal/provider/error_helper.go` — legacy `FormatAPIError()` (being phased out)

---

## Naming Conventions

| Concept | Pattern | Example |
|---------|---------|---------|
| Resource file | `{type}_resource.go` | `backup_resource.go` |
| DataSource file | `{type}_data_source.go` | `backup_data_source.go` |
| Resource struct | `{Type}Resource` | `BackupResource` |
| DataSource struct | `{Type}DataSource` | `BackupDataSource` |
| Resource model | `{Type}ResourceModel` | `BackupResourceModel` |
| DataSource model | `{Type}DataSourceModel` | `BackupDataSourceModel` |
| Nested object model | `{Parent}{Child}Model` | `CloudServerNetworkModel` |
| Model fields | PascalCase | `ProjectID`, `RetentionDays` |
| Local variables | camelCase | `projectID`, `backupID` |
| `tfsdk` tags | snake_case | `tfsdk:"project_id"` |

---

## Resource Pattern (all resources follow this structure)

```go
// 1. Model struct
type BackupResourceModel struct {
    Id        types.String `tfsdk:"id"`
    // ...all fields
}

// 2. Resource struct
type BackupResource struct {
    client *ArubaCloudClient
}

// 3. Implements: resource.Resource + resource.ResourceWithImportState
// 4. Methods: Metadata, Schema, Configure, Create, Read, Update, Delete, ImportState
```

**Create lifecycle**: Read plan → Build SDK request → Call `Create()` → Check `response.IsError()` → Extract ID → `WaitForResourceActive()` → Re-read via `Get()` → Set state.

**Read lifecycle**: Read state → Call `Get()` → If 404 call `resp.State.RemoveResource(ctx)` → Map fields → Set state.

**Error handling**: Use `CheckResponse[T]()` from `provider_error.go`; fall back to `FormatAPIError()` only in legacy code.

---

## SDK Usage Pattern

```go
// Service accessors:
client.Client.FromCompute()    // CloudServers, KeyPairs
client.Client.FromNetwork()    // VPCs, Subnets, SecurityGroups, ElasticIPs, VPNs
client.Client.FromStorage()    // Volumes, Backups, Restores
client.Client.FromDatabase()   // DBaaS, Databases, Grants, Users
client.Client.FromContainer()  // KaaS, ContainerRegistry

// Call signature:
client.Client.From{Service}().{Resource}().{Operation}(ctx, projectID[, resourceID], request, nil)
```

---

## Schema Rules

- Every attribute **must** have `MarkdownDescription`.
- Attribute order: `id`, `uri` (computed-only) → required → optional.
- Use `stringplanmodifier.UseStateForUnknown()` on all `Computed: true` fields assigned by the API.
- Use `stringvalidator.OneOf(...)` on fields with a fixed set of valid values.

---

## Make Commands

```bash
make              # fmt → lint → test → build → generate
make build        # Build provider binary
make test         # Unit tests (fast)
make testacc      # Acceptance tests (requires TF_ACC=1 + real credentials)
make generate     # Regenerate docs (run after any schema change)
make lint         # Run golangci-lint
make testcov      # HTML coverage report
```

---

## Prompts

Reusable prompt files for common tasks live in `.github/prompts/`:

| File | Purpose |
|------|---------|
| [`issue-solve.prompt.md`](prompts/issue-solve.prompt.md) | Diagnose and fix a reported Terraform provider issue |
| [`new-resource.prompt.md`](prompts/new-resource.prompt.md) | Scaffold a new resource |
| [`new-datasource.prompt.md`](prompts/new-datasource.prompt.md) | Scaffold a new data source |
| [`plan-techDebtRemediation.prompt.md`](prompts/plan-techDebtRemediation.prompt.md) | Plan tech debt remediation |
