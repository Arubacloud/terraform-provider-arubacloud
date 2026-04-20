---
mode: agent
description: Scaffold a new data source for the ArubaCloud Terraform provider
---

Read [`ai/ARCHITECTURE.md`](../../ai/ARCHITECTURE.md) and [`ai/CONVENTIONS.md`](../../ai/CONVENTIONS.md) before starting.

---

## New Data Source Specification

**Data source name (snake_case):** <!-- e.g. snapshot -->
**Terraform type:** `arubacloud_{name}` <!-- e.g. arubacloud_snapshot -->
**SDK service accessor:** `client.Client.From{Service}()` <!-- e.g. FromStorage() -->
**SDK resource accessor:** `.{Resource}()` <!-- e.g. .Snapshots() -->

**Query inputs (Required):**

| Attribute | Type | Notes |
|-----------|------|-------|
| `id`      | string | Resource ID to look up |
| `project_id` | string | |
| <!-- add more --> | | |

**Output fields (Computed):**

| Attribute | Type | Source |
|-----------|------|--------|
| `uri`     | string | `*response.Data.Metadata.URI` |
| `name`    | string | `*response.Data.Metadata.Name` |
| <!-- add more --> | | |

---

## Implementation Steps

1. **Create `internal/provider/{name}_data_source.go`** with:
   - `{Type}DataSourceModel` struct — query inputs as `Required: true`, outputs as `Computed: true`
   - `{Type}DataSource` struct holding `client *ArubaCloudClient`
   - Methods: `Metadata`, `Schema`, `Configure`, `Read`

2. **Schema rules:**
   - Every attribute must have `MarkdownDescription`
   - Query input attributes: `Required: true`
   - All output attributes: `Computed: true`
   - No `PlanModifiers` needed on data source attributes

3. **Read lifecycle:**
   ```go
   // 1. Read config
   req.Config.Get(ctx, &data)
   
   // 2. Call SDK Get
   response, err := client.Client.From{Service}().{Resource}().Get(ctx, projectID, resourceID, nil)
   
   // 3. Handle errors
   providerErr := CheckResponse[sdktypes.YourType](response, err)
   if providerErr != nil {
       if providerErr.IsNotFound() {
           resp.Diagnostics.AddError("Data source not found", providerErr.Error())
           return
       }
       resp.Diagnostics.AddError("Failed to read {Name}", providerErr.Error())
       return
   }
   
   // 4. Map fields
   data.Uri = types.StringValue(*response.Data.Metadata.URI)
   // ...
   
   // 5. Set state
   resp.State.Set(ctx, &data)
   ```

4. **Register** the new data source in `provider.go` `DataSources()` factory slice.

5. **Add an example** in `examples/data-sources/arubacloud_{name}/data-source.tf`.

6. Run `make build` → `make test` → `make generate` (updates docs).
