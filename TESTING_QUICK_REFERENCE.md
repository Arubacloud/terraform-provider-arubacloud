# Quick Reference - Testing & Quality

## Test Commands

```bash
# Build provider
make build

# Run unit tests
make test

# Run acceptance tests (may create real resources)
make testacc

# Run specific test
make testacc-run TEST=TestAccBackupResource

# Generate coverage report
make testcov

# Format code
make fmt

# Run linter
make lint

# Generate documentation
make docs
```

## Test File Locations

- **Resource tests**: `internal/provider/*_resource_test.go`
- **Data source tests**: `internal/provider/*_data_source_test.go`
- **Provider setup**: `internal/provider/provider_test.go`

## Creating New Tests

### Option 1: Generate stubs automatically
```bash
bash scripts/generate-test-stubs.sh
```

### Option 2: Copy existing test
```bash
cp internal/provider/backup_resource_test.go \
   internal/provider/mynewresource_resource_test.go
```
Then update:
- Function names
- Resource names
- Schema attributes
- Config functions

## Test Structure

```go
func TestAccMyResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create
            {
                Config: testAccMyResourceConfig("name"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "arubacloud_myresource.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact("name"),
                    ),
                },
            },
            // Import
            {
                ResourceName:      "arubacloud_myresource.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
            // Update
            {
                Config: testAccMyResourceConfig("new-name"),
                // Add checks...
            },
        },
    })
}
```

## Common State Checks

```go
// String value
statecheck.ExpectKnownValue(
    "arubacloud_resource.test",
    tfjsonpath.New("name"),
    knownvalue.StringExact("expected"),
)

// Integer value
statecheck.ExpectKnownValue(
    "arubacloud_resource.test",
    tfjsonpath.New("count"),
    knownvalue.Int64Exact(42),
)

// Boolean value
statecheck.ExpectKnownValue(
    "arubacloud_resource.test",
    tfjsonpath.New("enabled"),
    knownvalue.Bool(true),
)

// Nested attribute
statecheck.ExpectKnownValue(
    "arubacloud_resource.test",
    tfjsonpath.New("properties").AtMapKey("size"),
    knownvalue.Int64Exact(100),
)

// List size
statecheck.ExpectKnownValue(
    "arubacloud_resource.test",
    tfjsonpath.New("tags"),
    knownvalue.ListSizeExact(3),
)

// Not null
statecheck.ExpectKnownValue(
    "arubacloud_resource.test",
    tfjsonpath.New("id"),
    knownvalue.NotNull(),
)
```

## Environment Variables

```bash
# Enable acceptance tests
export TF_ACC=1

# Enable debug logging
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform.log

# Provider credentials (if needed)
export ARUBACLOUD_API_KEY=your-key
export ARUBACLOUD_API_SECRET=your-secret
```

## CI/CD Status

Tests run automatically on:
- Every push to any branch
- Every pull request
- Acceptance tests only on `main` branch

Check status: `.github/workflows/test.yml`

## Documentation

- **Full Testing Guide**: [docs/TESTING.md](docs/TESTING.md)
- **Code Quality Summary**: [CODE_QUALITY_IMPROVEMENTS.md](CODE_QUALITY_IMPROVEMENTS.md)
- **Script Documentation**: [scripts/README.md](scripts/README.md)

## Files Created

### Test Files
- `internal/provider/backup_resource_test.go`
- `internal/provider/backup_data_source_test.go`
- `internal/provider/blockstorage_resource_test.go`
- `internal/provider/blockstorage_data_source_test.go`

### Documentation
- `docs/TESTING.md` - Comprehensive testing guide
- `CODE_QUALITY_IMPROVEMENTS.md` - Summary of improvements

### Scripts
- `scripts/format-docs.sh` - Format generated documentation
- `scripts/format-docs.ps1` - PowerShell version
- `scripts/generate-test-stubs.sh` - Generate test templates

### CI/CD
- Enhanced `.github/workflows/test.yml` with test jobs

## Quick Tips

1. **Tests not running?** Make sure `TF_ACC=1` is set for acceptance tests
2. **Test failing?** Check the resource schema for required fields
3. **Need more examples?** Look at existing test files
4. **Want to debug?** Use `TF_LOG=DEBUG` for verbose output
5. **Coverage too low?** Run `make testcov` to see what's missing
