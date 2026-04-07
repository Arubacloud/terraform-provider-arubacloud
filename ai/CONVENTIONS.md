# CONVENTIONS.md — Code Conventions

## Naming Conventions

| Concept | Convention | Example |
|---------|-----------|---------|
| Resource file | `{type}_resource.go` | `backup_resource.go` |
| DataSource file | `{type}_data_source.go` | `backup_data_source.go` |
| Test file | `{type}_resource_test.go` | `backup_resource_test.go` |
| Resource struct | `{Type}Resource` | `BackupResource` |
| DataSource struct | `{Type}DataSource` | `BackupDataSource` |
| Resource model | `{Type}ResourceModel` | `BackupResourceModel` |
| DataSource model | `{Type}DataSourceModel` | `BackupDataSourceModel` |
| Nested object model | `{Parent}{Child}Model` | `CloudServerNetworkModel` |
| Model struct fields | PascalCase | `ProjectID`, `RetentionDays` |
| Local variables | camelCase | `projectID`, `backupID` |
| `tfsdk` tags | snake_case | `tfsdk:"project_id"` |

---

## Model Structs

Fields use Terraform Framework types with `tfsdk` snake_case tags matching schema attribute names:

```go
type BackupResourceModel struct {
    Id            types.String `tfsdk:"id"`
    Uri           types.String `tfsdk:"uri"`
    Name          types.String `tfsdk:"name"`
    Location      types.String `tfsdk:"location"`
    Tags          types.List   `tfsdk:"tags"`
    ProjectID     types.String `tfsdk:"project_id"`
    Type          types.String `tfsdk:"type"`
    VolumeID      types.String `tfsdk:"volume_id"`
    RetentionDays types.Int64  `tfsdk:"retention_days"`
    BillingPeriod types.String `tfsdk:"billing_period"`
}
```

Nested objects get their own dedicated struct:

```go
type CloudServerNetworkModel struct {
    VpcUriRef            types.String `tfsdk:"vpc_uri_ref"`
    ElasticIpUriRef      types.String `tfsdk:"elastic_ip_uri_ref"`
    SubnetUriRefs        types.List   `tfsdk:"subnet_uri_refs"`
    SecurityGroupUriRefs types.List   `tfsdk:"securitygroup_uri_refs"`
}
```

---

## Schema Declarations

Every attribute must have `MarkdownDescription`. Attribute order convention: computed-only fields first (`id`, `uri`), then required, then optional.

```go
resp.Schema = schema.Schema{
    MarkdownDescription: "Backup resource",
    Attributes: map[string]schema.Attribute{
        "id": schema.StringAttribute{
            MarkdownDescription: "Backup identifier",
            Computed:            true,
        },
        "uri": schema.StringAttribute{
            MarkdownDescription: "Backup URI",
            Computed:            true,
            PlanModifiers: []planmodifier.String{
                stringplanmodifier.UseStateForUnknown(),
            },
        },
        "name": schema.StringAttribute{
            MarkdownDescription: "Backup name",
            Required:            true,
        },
        "retention_days": schema.Int64Attribute{
            MarkdownDescription: "Retention days",
            Optional:            true,
        },
        "tags": schema.ListAttribute{
            ElementType:         types.StringType,
            MarkdownDescription: "List of tags",
            Optional:            true,
        },
    },
}
```

**Plan modifiers**: use `stringplanmodifier.UseStateForUnknown()` on all `Computed: true` fields that are assigned by the API and should not change on subsequent plans.

**Nested attributes** use `schema.SingleNestedAttribute{}`:

```go
"storage": schema.SingleNestedAttribute{
    MarkdownDescription: "Storage configuration",
    Required:            true,
    Attributes: map[string]schema.Attribute{
        "size_gb": schema.Int64Attribute{
            MarkdownDescription: "Storage size in GB",
            Required:            true,
        },
    },
},
```

---

## Type Conversions

### Framework types → Go native

```go
// String
s := data.Name.ValueString()

// Int64
n := data.RetentionDays.ValueInt64()
nInt := int(n)
ptr := &nInt  // SDK fields are often *int

// Bool
b := data.Enabled.ValueBool()

// List of strings
var tags []string
if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
    diags := data.Tags.ElementsAs(ctx, &tags, false)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

// Nested object
var networkModel CloudServerNetworkModel
diags := data.Network.As(ctx, &networkModel, basetypes.ObjectAsOptions{})
resp.Diagnostics.Append(diags...)
if resp.Diagnostics.HasError() {
    return
}
```

### Go native → Framework types

```go
// String
data.Id = types.StringValue(*response.Data.Metadata.ID)

// Int64
data.RetentionDays = types.Int64Value(int64(*backup.Properties.RetentionDays))

// Bool
data.Bootable = types.BoolValue(*volume.Properties.Bootable)

// Null string
data.Uri = types.StringNull()

// List from Go slice
tagValues := make([]attr.Value, len(tags))
for i, tag := range tags {
    tagValues[i] = types.StringValue(tag)
}
tagsList, diags := types.ListValue(types.StringType, tagValues)
resp.Diagnostics.Append(diags...)

// Empty list
emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
resp.Diagnostics.Append(diags...)
```

---

## Optional Field Building

Always guard optional fields with null/unknown checks before setting SDK request fields:

```go
if !data.RetentionDays.IsNull() && !data.RetentionDays.IsUnknown() {
    retentionDays := int(data.RetentionDays.ValueInt64())
    createRequest.Properties.RetentionDays = &retentionDays
}

if !data.BillingPeriod.IsNull() && !data.BillingPeriod.IsUnknown() {
    billingPeriod := data.BillingPeriod.ValueString()
    createRequest.Properties.BillingPeriod = &billingPeriod
}
```

---

## State Management

```go
// Read plan
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
    return
}

// Read state
resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
    return
}

// Write state
resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

// Remove resource (404 or destroyed)
resp.State.RemoveResource(ctx)
```

---

## Error Handling

**Generic error**:
```go
resp.Diagnostics.AddError(
    "Missing Required Fields",
    "Project ID and Volume ID are required to create a backup",
)
```

**Attribute-level error** (provider configuration):
```go
resp.Diagnostics.AddAttributeError(
    path.Root("api_key"),
    "Unknown ArubaCloud API Key",
    "The provider cannot create the ArubaCloud API client as the API key is unknown.",
)
```

**API error**:
```go
if response != nil && response.IsError() && response.Error != nil {
    errorMsg := FormatAPIError(ctx, response.Error, "Failed to create backup", map[string]interface{}{
        "project_id": projectID,
    })
    resp.Diagnostics.AddError("API Error", errorMsg)
    return
}
```

**404 handling in Read** (removes resource from state):
```go
if response != nil && response.IsError() && response.Error != nil {
    if response.StatusCode == 404 {
        resp.State.RemoveResource(ctx)
        return
    }
    // handle other errors...
}
```

---

## Configure Method

Identical pattern in every resource/datasource:

```go
func (r *BackupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    client, ok := req.ProviderData.(*ArubaCloudClient)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *ArubaCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }
    r.client = client
}
```

---

## ImportState Method

Identical in every resource — uses the resource `id` field:

```go
func (r *BackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

---

## Preserving Immutable Fields in Update

Copy immutable fields from state before saving to avoid plan diffs:

```go
data.Id = state.Id
data.ProjectID = state.ProjectID
data.Uri = state.Uri         // computed by API, not in update response
data.VolumeID = state.VolumeID
data.Type = state.Type
```

---

## Logging

Log at the **end** of each CRUD operation (after state is saved):

```go
tflog.Trace(ctx, "created a Backup resource", map[string]interface{}{
    "backup_id":   data.Id.ValueString(),
    "backup_name": data.Name.ValueString(),
})
```

Use `tflog.Debug()` for full SDK request JSON, `tflog.Info()` for wait/retry status, `tflog.Warn()` for non-fatal issues.

---

## Test Conventions

**Function naming**: `TestAcc{TypeName}Resource` or `TestAcc{TypeName}DataSource`

**Test structure**:
```go
func TestAccBackupResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccBackupResourceConfig("test-backup", "de-1", "full", 30),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "arubacloud_backup.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact("test-backup"),
                    ),
                    statecheck.ExpectKnownValue(
                        "arubacloud_backup.test",
                        tfjsonpath.New("id"),
                        knownvalue.NotNull(),
                    ),
                },
            },
            // ImportState step:
            {
                ResourceName:      "arubacloud_backup.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

**HCL config helper pattern** — use `fmt.Sprintf` with indexed format verbs:

```go
func testAccBackupResourceConfig(name, location, backupType string, retentionDays int) string {
    return fmt.Sprintf(`
resource "arubacloud_backup" "test" {
  name           = %[1]q
  location       = %[2]q
  type           = %[3]q
  retention_days = %[4]d
}
`, name, location, backupType, retentionDays)
}
```

- Use `%[N]q` for quoted strings, `%[N]d` for integers
- Resource address: `arubacloud_{type}.test`
- Test instance name is always `test`

**StateCheck values**:
```go
knownvalue.StringExact("value")   // exact string match
knownvalue.Int64Exact(42)         // exact int64 match
knownvalue.Bool(true)             // boolean match
knownvalue.NotNull()              // non-null check
knownvalue.ListSizeExact(3)       // list length check
tfjsonpath.New("field")           // top-level attribute
tfjsonpath.New("network").AtMapKey("vpc_uri_ref")  // nested attribute
```
