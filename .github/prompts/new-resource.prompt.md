---
mode: agent
description: Scaffold a new resource for the ArubaCloud Terraform provider
---

Read [`ai/ARCHITECTURE.md`](../../ai/ARCHITECTURE.md) and [`ai/CONVENTIONS.md`](../../ai/CONVENTIONS.md) before starting.

---

## New Resource Specification

**Resource name (snake_case):** <!-- e.g. snapshot -->
**Terraform type:** `arubacloud_{name}` <!-- e.g. arubacloud_snapshot -->
**SDK service accessor:** `client.Client.From{Service}()` <!-- e.g. FromStorage() -->
**SDK resource accessor:** `.{Resource}()` <!-- e.g. .Snapshots() -->

**Fields:**

| Attribute | Type | Required/Optional/Computed | Notes |
|-----------|------|---------------------------|-------|
| `id`      | string | Computed | Set from `*response.Data.Metadata.ID` |
| `uri`     | string | Computed | Set from `*response.Data.Metadata.URI` |
| `project_id` | string | Required | |
| <!-- add more rows --> | | | |

**Immutable fields (require replace):** <!-- list fields that cannot be updated in-place -->

---

## Implementation Steps

1. **Create `internal/provider/{name}_resource.go`** with:
   - `{Type}ResourceModel` struct (all fields as `types.*` with `tfsdk` tags)
   - `{Type}Resource` struct holding `client *ArubaCloudClient`
   - All required interface methods: `Metadata`, `Schema`, `Configure`, `Create`, `Read`, `Update`, `Delete`, `ImportState`

2. **Schema rules:**
   - Every attribute must have `MarkdownDescription`
   - Order: `id`, `uri` (computed) → required → optional
   - Add `stringplanmodifier.UseStateForUnknown()` on all computed fields
   - Add `RequiresReplace()` on immutable fields
   - Add `stringvalidator.OneOf(...)` on constrained string fields

3. **Create lifecycle:**
   - Read plan → build SDK request → call `Create()` → `CheckResponse[T]()` → extract ID → `WaitForResourceActive()` → re-read via `Get()` → set state

4. **Read lifecycle:**
   - Read state → call `Get()` → if 404 → `resp.State.RemoveResource(ctx)` → map fields → set state

5. **Delete lifecycle:**
   - Read state → call `Delete()` → `CheckResponse[T]()` → `WaitForResourceActive()` (wait for 404) or simple poll

6. **Register** the new resource in `provider.go` `Resources()` factory slice.

7. **Add an example** in `examples/resources/arubacloud_{name}/resource.tf`.

8. Run `make build` → `make test` → `make generate` (updates docs).

---

## Error Handling Pattern

```go
providerErr := CheckResponse[sdktypes.YourResponseType](response, err)
if providerErr != nil {
    resp.Diagnostics.AddError("Failed to create {Name}", providerErr.Error())
    return
}
```
