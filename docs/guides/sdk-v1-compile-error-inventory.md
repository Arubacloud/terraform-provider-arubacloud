# sdk-go v1.0.4 Compile Error Inventory

Generated after bumping `go.mod` from `sdk-go v0.1.24` → `v1.0.4` and running `go build -gcflags=all='-e' ./...`.

## Summary

| Category | Count | Scope | Fix strategy |
|---|---|---|---|
| `.Data undefined` | 660 | All 50 files | Replace `response.Data.*` access with wrapper promoted accessors (`wrapper.ID()`, `wrapper.URI()`, `wrapper.Name()`, `wrapper.State()`, `wrapper.Raw().Properties.*`) |
| `string→CallOption` (trailing `nil`) | 399 | All 50 files | Remove trailing `nil` argument — v1.0.4 CRUD methods no longer accept a trailing options parameter via the old `nil` pattern |
| `string→Ref` (parent ID params) | 214 | All 50 files | Replace explicit string parent IDs with builder-embedded `aruba.URI(...)` Refs; Create no longer takes separate `projectID` argument |
| `CheckResponse` type mismatch | 130 | All files using `CheckResponse` | Replace `CheckResponse[T](*sdktypes.Response[T])` with `CheckResponseErr(op, res, error)` (see issue #135) |
| `sdktypes.*` undefined | 107 | 25 resource files | Replace struct-literal request construction with fluent builder chains (see issues #136–#142) |
| `.IsError()` undefined | 15 | Files using `.IsError()` | Remove — errors surface via `(wrapper, error)` return; check `err != nil` instead |
| **Total** | **~1,525** | **50 files** | |

## Files affected

All 25 resource files and all 25 data source files in `internal/provider/`.

## Error category detail

### 1. `.Data undefined` (660 errors) — highest volume

Every `response.Data.Metadata.ID`, `response.Data.Metadata.URI`, `response.Data.Properties.*`, `response.Data.Status.State` access.

**Before**:
```go
data.Id  = types.StringValue(*response.Data.Metadata.ID)
data.Uri = types.StringValue(*response.Data.Metadata.URI)
state    := *response.Data.Status.State
```

**After**:
```go
data.Id  = types.StringValue(wrapper.ID())
data.Uri = types.StringValue(wrapper.URI())
state    := string(wrapper.State())
// Resource-specific properties:
raw := wrapper.Raw()
data.SizeGB = types.Int64Value(int64(*raw.Properties.SizeGB))
```

### 2. `string→CallOption` (399 errors) — trailing `nil` parameter

Every CRUD call ends with `..., nil)` which was the options parameter. v1.0.4 removes it from these positions.

**Before**:
```go
r.client.Client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
```

**After**:
```go
r.client.Client.FromNetwork().VPCs().Get(ctx, aruba.URI(data.Uri.ValueString()))
```

### 3. `string→Ref` (214 errors) — CRUD signature change

Parent IDs passed as explicit string parameters are no longer accepted. The builder embeds them.

**Before**:
```go
r.client.Client.FromNetwork().VPCs().Create(ctx, projectID, createRequest, nil)
```

**After**:
```go
r.client.Client.FromNetwork().VPCs().Create(ctx, aruba.NewVPC().Named(...).InProject(aruba.URI(...)))
```

### 4. `CheckResponse` type mismatch (130 errors)

`CheckResponse[T any](op, res string, resp *sdktypes.Response[T])` is called with the new wrapper types which don't implement `*sdktypes.Response[T]`.

Fix: Replace with `CheckResponseErr(op, res string, err error) *ProviderError` — see issue #135.

### 5. `sdktypes.*` undefined (107 errors)

Request struct types from `pkg/types` that were used for literal construction. All replaced by builder methods.

Examples:
- `sdktypes.CloudServerRequest{...}` → `aruba.NewCloudServer().Named(...)`
- `sdktypes.BlockStorageRequest{...}` → `aruba.NewBlockStorage().Named(...)`
- `sdktypes.SubnetDHCP{...}` → `aruba.NewSubnetDHCP().Enabled()...`
- `sdktypes.NodePoolProperties{...}` → `aruba.NewNodePool().Named()...`

### 6. `.IsError()` undefined (15 errors)

`response.IsError()` no longer exists on wrappers. Errors surface via the `error` return value of CRUD methods.

**Before**:
```go
if response.IsError() { ... }
```

**After**:
```go
if err != nil { ... }
```

## Confirmed architectural decisions (from this analysis)

1. **Module path unchanged**: `github.com/Arubacloud/sdk-go` (no v2 suffix) ✅
2. **All 50 files affected**: Every resource and data source needs updating
3. **Dominant change is `.Data` access**: 660 errors — the wrapper pattern replaces the `Response[T].Data` envelope entirely
4. **CRUD signature changes**: Confirmed across all 214 parent-ID usage sites
5. **`pkg/types` request types**: 107 undefined errors confirm complete elimination is required
6. **`provider_error.go`**: The 130 `CheckResponse` mismatches are all in resource files; fixing `provider_error.go` (issue #135) first eliminates them all
7. **`go mod tidy` succeeds**: No dependency conflicts; v1.0.4 is compatible with the rest of the dependency tree

## Implementation order

Based on this analysis, the fix order is:

```
#134 (auth rename — independent)
#135 (CheckResponse → CheckResponseErr — prerequisite for all resource files)
    ↓
#136 (compute)  #137 (networking)  #138 (ext-network)
#139 (storage)  #140 (database)    #141 (container)   #142 (security+schedule)
    ↓ (can parallelise)
#143 (cleanup)  #144 (tests)  #145 (docs+release)
```
