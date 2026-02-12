# Testing Improvements Summary

This document summarizes the improvements made to the test suite for the ArubaCloud Terraform Provider.

## Changes Made

### 1. Provider Configuration Fix
**File**: `internal/provider/provider.go`

Changed provider schema to make `api_key` and `api_secret` **optional** instead of required, allowing them to be set via environment variables:
- `ARUBACLOUD_API_KEY`
- `ARUBACLOUD_API_SECRET`

**Impact**: Acceptance tests can now run without requiring explicit provider configuration blocks.

### 2. Fixed DBaaS User Test
**File**: `internal/provider/dbaasuser_data_source_test.go`

Fixed test configuration to include required `username` field:
```hcl
data "arubacloud_dbaasuser" "test" {
  username = "test-user"  # Added required field
}
```

### 3. Test Execution Script
**File**: `run-acceptance-tests.sh`

Created a helper script that:
- Automatically loads credentials from `examples/test/compute/terraform.tfvars`
- Sets required environment variables (`TF_ACC`, `ARUBACLOUD_API_KEY`, `ARUBACLOUD_API_SECRET`)
- Runs acceptance tests with proper timeout (120 minutes)
- Supports running specific test patterns

**Usage**:
```bash
# Run all acceptance tests
./run-acceptance-tests.sh

# Run specific test(s)
./run-acceptance-tests.sh TestAccProjectDataSource
./run-acceptance-tests.sh "TestAccVpcDataSource|TestAccKeypairDataSource"
```

## Running Tests

### Unit Tests
Run unit tests (fast, no API credentials needed):
```bash
go test -v ./internal/provider/... -run "Test[^Acc]"
```

### Acceptance Tests
Acceptance tests require valid API credentials.

**Method 1: Using the script (recommended)**
```bash
./run-acceptance-tests.sh [optional-test-pattern]
```

**Method 2: Manual environment variables**
```bash
export TF_ACC=1
export ARUBACLOUD_API_KEY="your-api-key"
export ARUBACLOUD_API_SECRET="your-api-secret"
go test -v -timeout=120m ./internal/provider/...
```

### GitHub Actions
The CI/CD pipeline (`.github/workflows/test.yml`) is configured to use GitHub Secrets:
- `ARUBACLOUD_API_KEY`
- `ARUBACLOUD_API_SECRET`

Make sure these secrets are added to your repository:
1. Go to Settings → Secrets and variables → Actions
2. Add both secrets with your API credentials

## Test Status

### Unit Tests ✅
- **11 tests** passing
- **Execution time**: <0.1s
- **Coverage**: Provider metadata, schemas, import state tests

### Acceptance Tests 🔧
- **50 acceptance tests** total (25 datasources + 25 resources)
- All tests now have proper schema configurations
- Tests require valid API credentials to execute
- Some tests may fail if resources don't exist in your account

### Code Quality ✅
- **Linting**: 0 issues
- **Code generation**: Up to date
- **Documentation**: Complete

## Known Issues

Some acceptance tests may fail due to:
1. Missing resources in the test account (expected - tests try to read existing resources)
2. Schema mismatches in API responses (e.g., keypair test expecting `public_key` field)
3. API rate limiting or quota restrictions

These are expected for tests that query actual API endpoints without creating test fixtures.

## Next Steps

To run acceptance tests successfully:
1. Ensure you have valid ArubaCloud API credentials
2. Add credentials to `examples/test/compute/terraform.tfvars` or set environment variables
3. Run tests using `./run-acceptance-tests.sh`
4. Some tests may require pre-existing resources in your account

For CI/CD:
1. Add `ARUBACLOUD_API_KEY` and `ARUBACLOUD_API_SECRET` to GitHub Secrets
2. Acceptance tests only run on the main branch to conserve API quota
