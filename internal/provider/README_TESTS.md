# Unit Test Suite

This directory contains comprehensive unit tests for the Terraform Provider ArubaCloud.

## Test Files

### [datasource_schema_test.go](datasource_schema_test.go)
Unit tests for all datasource schemas. Tests validate:
- Schema structure and required attributes
- Metadata and type names
- Flattened field structure (no nested `properties` objects)
- Provider instantiation

**Tests**: 6 test functions validating 13 datasources

### [resource_schema_test.go](resource_schema_test.go)
Unit tests for all resource schemas. Tests validate:
- Schema structure and required attributes
- Metadata and type names
- ImportState support for critical resources
- Required fields for key resources

**Tests**: 8 test functions validating 25 resources

## Running Unit Tests

### Run all unit tests
```bash
go test ./internal/provider -v -run="Test(DataSource|Resource|Provider)"
```

### Run specific test
```bash
go test ./internal/provider -v -run="TestDataSourceSchemas"
go test ./internal/provider -v -run="TestResourceMetadata"
```

### Run with coverage
```bash
go test ./internal/provider -cover -run="Test(DataSource|Resource|Provider)"
```

## Test Characteristics

- **Execution Time**: <0.1 seconds
- **API Required**: No
- **Dependencies**: None (uses provider framework only)
- **Coverage**: 38 datasources + resources

## Acceptance Tests

Acceptance tests are in separate `*_test.go` files and require:
- `TF_ACC=1` environment variable
- Valid API credentials
- Real API access

See [../../docs/TESTING.md](../../docs/TESTING.md) for complete testing guide.

## CI/CD Integration

These unit tests can be run on every commit/PR:

```yaml
- name: Run unit tests
  run: go test ./internal/provider -v -run="Test(DataSource|Resource|Provider)"
```

## Related Documentation

- [TEST_IMPROVEMENTS.md](../../TEST_IMPROVEMENTS.md) - Summary of test improvements
- [TESTING.md](../../docs/TESTING.md) - Complete testing guide
- [CODE_QUALITY_IMPROVEMENTS.md](../../CODE_QUALITY_IMPROVEMENTS.md) - Linting fixes
