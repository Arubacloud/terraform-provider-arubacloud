# ARCHITECTURE.md — Architecture & Design Patterns

## Framework

Terraform provider built with [HashiCorp Terraform Plugin Framework v1](https://github.com/hashicorp/terraform-plugin-framework) (Protocol v6). SDK dependency: `github.com/Arubacloud/sdk-go v0.1.24`.

The `ArubaCloudProvider` struct implements three interfaces:
- `provider.Provider`
- `provider.ProviderWithFunctions`
- `provider.ProviderWithEphemeralResources`

---

## Provider Initialization

**Entry point**: `main.go` → `provider.New(version)` → `providerserver.Serve()`. Supports `-debug` flag for debugger attachment.

**Provider config model** (`provider.go:35-41`):
```go
type ArubaCloudProviderModel struct {
    ApiKey          types.String `tfsdk:"api_key"`
    ApiSecret       types.String `tfsdk:"api_secret"`
    ResourceTimeout types.String `tfsdk:"resource_timeout"`  // e.g. "5m", "10m"
    BaseURL         types.String `tfsdk:"base_url"`
    TokenIssuerURL  types.String `tfsdk:"token_issuer_url"`
}
```

**Configuration precedence** (`provider.go:76-160`):
1. Environment variables checked first: `ARUBACLOUD_API_KEY`, `ARUBACLOUD_API_SECRET`, `ARUBACLOUD_TOKEN_ISSUER_URL`
2. HCL provider block overrides env vars

**SDK client creation** (`provider.go:126-138`):
```go
options := aruba.DefaultOptions(apiKey, apiSecret).WithDefaultLogger()
// Optional:
options = options.WithBaseURL(baseURL)
options = options.WithTokenIssuerURL(tokenIssuerURL)
sdkClient, err := aruba.NewClient(options)
```

**Wrapped client** passed to all resources/datasources as provider data (`provider.go:179-184`):
```go
type ArubaCloudClient struct {
    ApiKey          string
    ApiSecret       string
    Client          aruba.Client
    ResourceTimeout time.Duration  // default 10m, parsed from HCL string
}
```

**Timeout**: default 10 minutes, parsed via `time.ParseDuration()`, falls back to default on error.

---

## Resource & DataSource Registration

Resources (27) and datasources (23) are registered via factory functions in `provider.go:186-254`. Each factory returns a new instance of the struct.

---

## Resource Pattern

Every resource is composed of 4 elements:

1. **Model struct** (`{Type}ResourceModel`) — Terraform state mapped via `tfsdk` tags
2. **Resource struct** (`{Type}Resource`) — holds `*ArubaCloudClient`
3. Implements: `resource.Resource` + `resource.ResourceWithImportState`
4. Methods: `Metadata()`, `Schema()`, `Configure()`, `Create()`, `Read()`, `Update()`, `Delete()`, `ImportState()`

Data sources only implement `datasource.DataSource` with `Read()`. All their schema attributes are `Computed: true`; query inputs are `Required: true`.

---

## SDK Client Usage

The SDK client exposes service namespaces via fluent accessors. All operations follow the same signature:

```go
// Pattern:
client.Client.From{Service}().{Resource}().{Operation}(ctx, projectID[, resourceID], request, nil)

// Services:
client.Client.FromCompute()    // CloudServers, KeyPairs
client.Client.FromNetwork()    // VPCs, Subnets, SecurityGroups, SecurityGroupRules, ElasticIPs, VPCPeerings, VPCPeeringRoutes, VPNTunnels, VPNRoutes
client.Client.FromStorage()    // Volumes (BlockStorage), Backups, Restores
client.Client.FromDatabase()   // DBaaS, Databases, Grants, Users, Backups
client.Client.FromContainer()  // KaaS, ContainerRegistry
```

**Response structure** (inferred from all handlers):
```go
type SDKResponse struct {
    IsError() bool              // method — check first
    StatusCode int
    Error *ErrorResponse        // present when IsError() == true
    Data *ResourceResponse      // present on success
}

type ResourceResponse struct {
    Metadata ResourceMetadataResponse   // .ID (*string), .URI (*string), .Name (*string), .Tags []string
    Properties *Properties
    Status     StatusResponse           // .State (*string)
}
```

The last parameter of all SDK calls is always `nil` (options, currently unused).

---

## Resource Lifecycle — Create

Canonical pattern (`cloudserver_resource.go:187-434`, `backup_resource.go:108-262`):

1. Read plan: `req.Plan.Get(ctx, &data)`
2. Extract nested objects and validate required IDs
3. Build SDK request struct (e.g., `sdktypes.CloudServerRequest`, `sdktypes.StorageBackupRequest`)
4. Call SDK `Create()`
5. Check `response.IsError()` → call `FormatAPIError()` → `resp.Diagnostics.AddError()`
6. Extract `*response.Data.Metadata.ID` → set `data.Id`
7. **Wait**: call `WaitForResourceActive()` with a checker closure that calls `Get()` and returns `*response.Data.Status.State`
8. On timeout: save partial state (with ID) so destroy can clean up
9. Re-read via `Get()` to populate URI and other server-assigned fields
10. Set state: `resp.State.Set(ctx, &data)`
11. Log: `tflog.Trace(ctx, "created a X resource", map[string]interface{}{...})`

---

## Resource Lifecycle — Read

Pattern (`cloudserver_resource.go:437-610`, `backup_resource.go:266-375`):

1. Read state: `req.State.Get(ctx, &data)`
2. Extract IDs, call SDK `Get()`
3. If **404**: `resp.State.RemoveResource(ctx)` and return
4. Preserve fields not returned by API from the existing state (prevents spurious diffs)
5. Rebuild nested objects merging API response + preserved state
6. Set state

---

## Resource Lifecycle — Update

Pattern (`cloudserver_resource.go:612-827`, `backup_resource.go:397-485`):

1. Read plan and state
2. Fetch current resource from API to get latest values
3. Build update request (keeping immutable fields from current API state)
4. Call SDK `Update()`
5. Preserve immutable fields from state (e.g. `data.Uri = state.Uri`)
6. Set state

Update operations do **not** call `WaitForResourceActive()` — assumed synchronous.

---

## Resource Lifecycle — Delete

All resources use `DeleteResourceWithRetry()` (`resource_wait.go:170-273`):

```go
err := DeleteResourceWithRetry(
    ctx,
    func() (interface{}, error) {
        return r.client.Client.From{Service}().{Resource}().Delete(ctx, projectID, resourceID, nil)
    },
    ExtractSDKError,
    "ResourceType",
    resourceID,
    r.client.ResourceTimeout,
)
```

**Retry strategy**:
- **404** = not retried (treated as already deleted = success)
- Any other error = retry with exponential backoff: `min(5*attempt seconds, 30s)`
- Continues until deadline (`ResourceTimeout`) is exceeded
- Logs each attempt

---

## Polling — WaitForResourceActive

`resource_wait.go:19-50` — polls every **5 seconds** until the resource leaves a transitional state.

**Transitional states** (not ready, `resource_wait.go:54-73`):
`"InCreation"`, `"Creating"`, `"Updating"`, `"Deleting"`, `"Pending"`, `"Provisioning"`

Any other state (e.g. `"Active"`, `"NotUsed"`, `"Running"`) = ready.

**Caller provides a checker closure**:
```go
checker := func(ctx context.Context) (string, error) {
    getResp, err := r.client.Client.From{Service}().{Resource}().Get(ctx, projectID, resourceID, nil)
    if err != nil {
        return "", err
    }
    if getResp != nil && getResp.Data != nil && getResp.Data.Status.State != nil {
        return *getResp.Data.Status.State, nil
    }
    return "Unknown", nil
}
WaitForResourceActive(ctx, checker, "ResourceType", resourceID, r.client.ResourceTimeout)
```

---

## Error Extraction — ExtractSDKError

`resource_wait.go:278-337` — uses **reflection** to access SDK response fields generically across all response types:

- Calls `response.IsError()` method via reflection
- Extracts `response.StatusCode` (int field)
- Extracts `response.Error.Title` and `response.Error.Detail` (*string fields)

This allows `DeleteResourceWithRetry` to work with any SDK response type without explicit casting.

---

## API Error Formatting — FormatAPIError

`error_helper.go:14-78` — formats API errors for user-facing diagnostics:

```go
FormatAPIError(ctx, response.Error, "Failed to create backup", map[string]interface{}{
    "project_id": projectID,
})
```

**Processing**:
1. Appends `Title` and `Detail` into a readable message
2. Extracts field-level validation errors from `Extensions["errors"]` array:
   - Each entry: `{fieldName: "...", errorMessage: "..."}`
   - Formats as: `"\n  - fieldName: errorMessage"`
3. Falls back to dumping all `Extensions` keys if not in validation format
4. Logs full JSON response via `tflog.Error()` for debugging

---

## State Preservation Patterns

Two patterns prevent Terraform from detecting spurious diffs:

1. **Immutable field preservation** (Update): fields the API doesn't accept in update requests are copied from state before saving — e.g. `data.Uri = state.Uri`, `data.VolumeID = state.VolumeID`

2. **Nested object rebuilding** (Read): nested objects are reconstructed merging API response fields with the original state values for fields the API doesn't return (e.g. user-provided URI references)

---

## Logging

All operations use `github.com/hashicorp/terraform-plugin-log/tflog`. Standard levels:

| Level | When |
|-------|------|
| `tflog.Trace()` | End of CRUD operations (with resource ID and name) |
| `tflog.Debug()` | Full request JSON, detailed operation info |
| `tflog.Info()` | Retry attempts, wait status |
| `tflog.Warn()` | Non-fatal issues (e.g. refresh failures) |
| `tflog.Error()` | Full API error response as JSON |

---

## Cross-Cutting Concerns

- **No middleware/interceptors**: SDK is called directly in each resource handler
- **No retries on Create or Read**: only Delete uses retry logic
- **No Update waiting**: Updates are assumed synchronous
- **Single provider-level timeout**: one `ResourceTimeout` applies to all resource operations
- **Dependency error detection** (`resource_wait.go:78-143`): scans error messages for keywords like `"dependency"`, `"in use"`, `"still exists"`, `"attached"`, `"linked"` to classify retry-able delete failures
