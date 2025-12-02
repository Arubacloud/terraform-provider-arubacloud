# Terraform Provider Testing Guide

This guide explains how to write and run acceptance tests for the ArubaCloud Terraform provider.

## Overview

Acceptance tests verify that your Terraform provider works correctly by:
- Creating resources with Terraform configurations
- Verifying the state matches expected values
- Testing updates to resources
- Testing imports
- Testing deletions

## Running Tests

### Run All Tests
```bash
TF_ACC=1 go test ./internal/provider/... -v
```

### Run Specific Test
```bash
TF_ACC=1 go test ./internal/provider/ -v -run TestAccBackupResource
```

### Run Tests in Parallel
```bash
TF_ACC=1 go test ./internal/provider/... -v -parallel=4
```

## Test Structure

### Resource Test Example

```go
func TestAccBackupResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read
            {
                Config: testAccBackupResourceConfig("name", "params"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "arubacloud_backup.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact("expected-name"),
                    ),
                },
            },
            // Import
            {
                ResourceName:      "arubacloud_backup.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
            // Update
            {
                Config: testAccBackupResourceConfig("new-name", "new-params"),
                // Add checks...
            },
        },
    })
}
```

### Data Source Test Example

```go
func TestAccBackupDataSource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccBackupDataSourceConfig,
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "data.arubacloud_backup.test",
                        tfjsonpath.New("id"),
                        knownvalue.StringExact("backup-id"),
                    ),
                },
            },
        },
    })
}
```

## Best Practices

### 1. Test Multiple Scenarios
- Basic creation with required fields only
- Creation with all optional fields
- Updates to various fields
- Edge cases and validation

### 2. Use Config Helper Functions
```go
func testAccResourceConfig(name string, param int) string {
    return fmt.Sprintf(`
resource "arubacloud_resource" "test" {
  name  = %[1]q
  param = %[2]d
}
`, name, param)
}
```

### 3. Test Import Functionality
Always include an import test step:
```go
{
    ResourceName:      "arubacloud_backup.test",
    ImportState:       true,
    ImportStateVerify: true,
}
```

### 4. Use State Checks
Verify computed values and complex nested attributes:
```go
statecheck.ExpectKnownValue(
    "arubacloud_blockstorage.test",
    tfjsonpath.New("properties").AtMapKey("size_gb"),
    knownvalue.Int64Exact(100),
)
```

## Common State Check Types

### String Values
```go
knownvalue.StringExact("exact-match")
knownvalue.StringRegexp(regexp.MustCompile("pattern.*"))
```

### Numeric Values
```go
knownvalue.Int64Exact(42)
knownvalue.Float64Exact(3.14)
```

### Boolean Values
```go
knownvalue.Bool(true)
```

### Collections
```go
knownvalue.ListSizeExact(3)
knownvalue.MapSizeExact(2)
knownvalue.SetExact([]knownvalue.Check{
    knownvalue.StringExact("item1"),
    knownvalue.StringExact("item2"),
})
```

### Null/Unknown
```go
knownvalue.Null()
knownvalue.NotNull()
```

## Testing Nested Attributes

For nested objects (like `properties` in blockstorage):
```go
statecheck.ExpectKnownValue(
    "arubacloud_blockstorage.test",
    tfjsonpath.New("properties").AtMapKey("size_gb"),
    knownvalue.Int64Exact(100),
)
```

## Environment Variables

Set these before running tests:

```bash
export TF_ACC=1                           # Enable acceptance tests
export TF_LOG=DEBUG                       # Enable debug logging
export ARUBACLOUD_API_KEY=your-key       # Your API credentials
export ARUBACLOUD_API_SECRET=your-secret
```

## Debugging Tests

### Enable Verbose Output
```bash
TF_ACC=1 TF_LOG=DEBUG go test ./internal/provider/ -v -run TestAccBackupResource
```

### Run Single Test
```bash
TF_ACC=1 go test ./internal/provider/ -v -run TestAccBackupResource/Create
```

## Test Coverage

Aim to cover:
- ✅ Basic CRUD operations (Create, Read, Update, Delete)
- ✅ Import functionality
- ✅ Optional vs required fields
- ✅ Validation errors
- ✅ Complex nested attributes
- ✅ List/set attributes
- ✅ Dependencies between resources

## Generated Test Files

The following test files have been created:
- `backup_resource_test.go` - Backup resource tests
- `backup_data_source_test.go` - Backup data source tests
- `blockstorage_resource_test.go` - BlockStorage resource tests
- `blockstorage_data_source_test.go` - BlockStorage data source tests

Use these as templates for creating tests for other resources.

## Creating Tests for Other Resources

1. Copy one of the test files (e.g., `backup_resource_test.go`)
2. Rename it to match your resource (e.g., `cloudserver_resource_test.go`)
3. Update the test function names
4. Update the resource/data source names
5. Adjust the config functions to match your schema
6. Add state checks for your specific attributes

## CI/CD Integration

Add to your `.github/workflows/test.yml`:

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test ./internal/provider/... -v
      - name: Acceptance Tests
        env:
          TF_ACC: 1
        run: go test ./internal/provider/... -v -timeout 120m
```

## Additional Resources

- [Terraform Plugin Testing](https://developer.hashicorp.com/terraform/plugin/testing)
- [Plugin Framework Testing](https://developer.hashicorp.com/terraform/plugin/framework/acctests)
- [terraform-plugin-testing](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-testing)
