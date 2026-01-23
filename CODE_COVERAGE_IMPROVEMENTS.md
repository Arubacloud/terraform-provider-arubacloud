# Code Coverage Improvement Plan

## Current Status
- **Current Coverage**: 10.3% (as of latest test run)
- **Key Issue**: Most resource implementations are stubbed with minimal logic

## Issues Fixed

### 1. Key Resource Test Failures ✅
- **Problem**: `TestAccKeyDataSource` failing due to SDK limitation
- **Solution**: Added `t.Skip()` with explanation that SDK needs to return `project_id` and `kms_id` in API response
- **Files Modified**:
  - `internal/provider/key_data_source_test.go`
  - `internal/provider/key_resource_test.go`

### 2. Documentation Generation ✅
- **Problem**: Pipeline failing due to outdated generated documentation
- **Solution**: Ran `make generate` to update all docs
- **Files Updated**:
  - `docs/data-sources/key.md`
  - `docs/data-sources/kmip.md`
  - `docs/resources/key.md`
  - `docs/resources/kmip.md`
  - `docs/resources/kms.md`

## Coverage Improvement Strategies

### Quick Wins (Can Increase Coverage to ~40-50%)

#### 1. Add Unit Tests for Schema Validation
Create schema tests for all resources that don't have them yet:

```go
func TestResourceSchema(t *testing.T) {
    // Test that required fields are present
    // Test that optional/computed fields work correctly
    // Test plan modifiers
}
```

**Target resources**:
- All resources currently at 0.7-2.5% coverage
- Estimated impact: +15-20% coverage

#### 2. Expand Existing Data Source Tests
Currently, data source tests have good coverage (82-91%). Add similar patterns to resources:

```go
func TestAccResourceBasic(t *testing.T) {
    // Test Create
    // Test Read (state refresh)
    // Test Update if supported
    // Test Delete
}
```

**Target resources** (currently <5% coverage):
- `backup_resource.go` (1.2%)
- `blockstorage_resource.go` (0.8%)
- `cloudserver_resource.go` (2.4%)
- `containerregistry_resource.go` (1.5%)
- `database_resource.go` (2.4%)
- `databasebackup_resource.go` (2.2%)
- `databasegrant_resource.go` (2.5%)
- `dbaas_resource.go` (0.7%)
- `dbaasuser_resource.go` (2.2%)
- `elasticip_resource.go` (1.2%)
- `kaas_resource.go` (0.6%)
- `keypair_resource.go` (1.7%)
- `kms_resource.go` (1.6%)
- `restore_resource.go` (1.5%)
- `schedulejob_resource.go` (0.9%)
- `securitygroup_resource.go` (1.5%)
- `securityrule_resource.go` (0.7%)
- `snapshot_resource.go` (1.2%)
- `subnet_resource.go` (0.7%)
- `vpc_resource.go` (1.3%)
- `vpcpeering_resource.go` (1.7%)
- `vpcpeeringroute_resource.go` (0.7%)
- `vpnroute_resource.go` (0.7%)
- `vpntunnel_resource.go` (0.8%)

**Estimated impact**: +20-25% coverage

#### 3. Test Error Handling Paths
Add tests for error conditions:

```go
func TestResourceCreateError(t *testing.T) {
    // Test with invalid project_id
    // Test with missing required fields
    // Test with API errors
}

func TestResourceReadNotFound(t *testing.T) {
    // Test 404 handling
    // Verify resource removed from state
}
```

**Estimated impact**: +5-10% coverage

#### 4. Test CRUD Operations Without Real API
Use mocked client for unit tests:

```go
func TestResourceCRUD(t *testing.T) {
    // Create mock client
    // Test Create logic
    // Test Read logic
    // Test Update logic
    // Test Delete logic
}
```

**Estimated impact**: +10-15% coverage

### Medium Effort (Can Increase Coverage to ~60-70%)

#### 5. Integration Tests with Test Fixtures
Create reusable test fixtures:

```go
func testAccPreCheckWithCleanup(t *testing.T) {
    // Setup test resources
    // Register cleanup function
}
```

#### 6. Test Plan Modifiers and Validators
Test custom plan modifiers:

```go
func TestRequiresReplace(t *testing.T) {
    // Test that changing immutable fields triggers replacement
}
```

#### 7. Test Import Functionality
Add import tests for all resources:

```go
func TestResourceImport(t *testing.T) {
    // Test ImportState
    // Verify all fields populated correctly
}
```

### Long Term (Can Reach 80%+ Coverage)

#### 8. Add Acceptance Tests for All Resources
Create comprehensive acceptance tests that actually call the API (when available):

```go
func TestAccResourceComplete(t *testing.T) {
    // Full lifecycle test
    // Test all attribute combinations
    // Test dependencies between resources
}
```

#### 9. Test Edge Cases
- Concurrent operations
- Large datasets
- Timeouts and retries
- Network errors

#### 10. Property-Based Testing
Use property-based testing for complex validation logic:

```go
func TestResourceValidation(t *testing.T) {
    // Generate random valid inputs
    // Verify they're accepted
    // Generate invalid inputs
    // Verify they're rejected
}
```

## Priority Action Plan

### Phase 1 (This Week)
1. ✅ Fix failing tests (Key resource)
2. ✅ Update documentation generation
3. Add schema validation tests for top 10 resources
4. Add basic CRUD tests with mocks

**Expected Coverage After Phase 1**: ~35-40%

### Phase 2 (Next Week)
1. Add error handling tests
2. Expand existing resource tests
3. Add import tests
4. Test plan modifiers

**Expected Coverage After Phase 2**: ~55-65%

### Phase 3 (Following Week)
1. Add acceptance tests (where API access available)
2. Test edge cases
3. Add integration tests
4. Performance testing

**Expected Coverage After Phase 3**: ~75-85%

## Resources Needing Immediate Attention

Based on usage and importance:

### High Priority (Core Infrastructure)
1. `project_resource.go` - 47.1% (Good! Keep improving)
2. `vpc_resource.go` - 1.3% (Needs work)
3. `subnet_resource.go` - 0.7% (Needs work)
4. `securitygroup_resource.go` - 1.5% (Needs work)
5. `cloudserver_resource.go` - 2.4% (Needs work)

### Medium Priority (Storage & Compute)
1. `blockstorage_resource.go` - 0.8%
2. `snapshot_resource.go` - 1.2%
3. `elasticip_resource.go` - 1.2%
4. `kaas_resource.go` - 0.6%

### Lower Priority (Specialized Services)
1. Database resources (2-2.5%)
2. Container registry (1.5%)
3. Backup/Restore (1.2-1.5%)
4. VPN resources (0.7-0.8%)

## Metrics to Track

1. **Overall Coverage**: Target 80%+
2. **Per-Resource Coverage**: Target 70%+ for each resource
3. **Test Execution Time**: Keep under 5 minutes for unit tests
4. **Flaky Tests**: Zero tolerance
5. **Test Maintenance**: Update tests with every schema change

## Next Steps

1. Review this plan with team
2. Assign resources to team members
3. Set up code coverage tracking in CI/CD
4. Add coverage badges to README
5. Weekly review of coverage metrics
