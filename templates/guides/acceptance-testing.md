---
page_title: "Acceptance Testing - ArubaCloud Provider Development"
subcategory: "Development"
description: |-
  How to run acceptance tests for the ArubaCloud Terraform provider, including required environment variables and CI setup.
---

# Acceptance Testing

This guide is for contributors who want to run or extend the provider's acceptance test suite.

Acceptance tests exercise the real ArubaCloud API — they provision, read, update, and destroy actual cloud resources. They are separate from unit tests and are **not** run on every pull request to avoid consuming API quota.

## Prerequisites

- Go ≥ 1.24
- Terraform CLI in `PATH` (acceptance tests download and drive Terraform internally)
- A valid ArubaCloud account with API credentials
- An existing project to scope resource creation (`ARUBACLOUD_PROJECT_ID`)

## Running Locally

```bash
export TF_ACC=1
export ARUBACLOUD_API_KEY=<your-api-key>
export ARUBACLOUD_API_SECRET=<your-api-secret>
export ARUBACLOUD_PROJECT_ID=<an-existing-project-id>

# Run all acceptance tests
go test -v -timeout=120m ./internal/provider/... -run '^TestAcc'

# Run a single resource test
go test -v -timeout=30m ./internal/provider/... -run '^TestAccKeypairResource$'
```

A convenience wrapper is available that loads credentials from `examples/test/compute/terraform.tfvars`:

```bash
./run-acceptance-tests.sh TestAccKeypairResource
```

## CI — Manual Trigger

The [Acceptance Tests workflow](https://github.com/Arubacloud/terraform-provider-arubacloud/actions/workflows/acceptance.yml) can be triggered manually from any branch via **Actions → Acceptance Tests → Run workflow**.

| Input | Default | Description |
|---|---|---|
| `test_filter` | _(all `^TestAcc`)_ | Go `-run` filter, e.g. `TestAccVpcResource` |
| `timeout` | `120m` | Test timeout passed to `go test` |

The workflow also runs automatically on every push to `main`.

## Environment Variables

### Always Required

| Variable | Purpose |
|---|---|
| `TF_ACC` | Set to `"1"` to activate acceptance tests |
| `ARUBACLOUD_API_KEY` | API authentication |
| `ARUBACLOUD_API_SECRET` | API authentication |

### Resource Tests

Resource tests exercise the full Create → Read → Update → Delete lifecycle and only need:

| Variable | Purpose |
|---|---|
| `ARUBACLOUD_PROJECT_ID` | Scopes all resource creation to an existing project |

### Data Source Tests

Data source tests look up an **existing** resource by ID. A data source test skips gracefully when its variable is missing — set only the variables for the resources you want to test.

| Variable | Data source(s) tested |
|---|---|
| `ARUBACLOUD_PROJECT_ID` | All data sources |
| `ARUBACLOUD_VPC_ID` | `arubacloud_vpc`, `arubacloud_subnet`, `arubacloud_securitygroup`, `arubacloud_securityrule`, `arubacloud_vpcpeering`, `arubacloud_vpcpeeringroute` |
| `ARUBACLOUD_CLOUDSERVER_ID` | `arubacloud_cloudserver` |
| `ARUBACLOUD_KEYPAIR_ID` | `arubacloud_keypair` |
| `ARUBACLOUD_BLOCKSTORAGE_ID` | `arubacloud_blockstorage` |
| `ARUBACLOUD_SNAPSHOT_ID` | `arubacloud_snapshot` |
| `ARUBACLOUD_ELASTICIP_ID` | `arubacloud_elasticip` |
| `ARUBACLOUD_BACKUP_ID` | `arubacloud_backup`, `arubacloud_restore` |
| `ARUBACLOUD_RESTORE_ID` | `arubacloud_restore` |
| `ARUBACLOUD_DBAAS_ID` | `arubacloud_dbaas`, `arubacloud_dbaasuser`, `arubacloud_database`, `arubacloud_databasegrant` |
| `ARUBACLOUD_DATABASE_ID` | `arubacloud_database`, `arubacloud_databasegrant` |
| `ARUBACLOUD_DATABASE_BACKUP_ID` | `arubacloud_databasebackup` |
| `ARUBACLOUD_DBAAS_USERNAME` | `arubacloud_dbaasuser` |
| `ARUBACLOUD_DBAAS_USER_ID` | `arubacloud_databasegrant` |
| `ARUBACLOUD_KAAS_ID` | `arubacloud_kaas` |
| `ARUBACLOUD_CONTAINERREGISTRY_ID` | `arubacloud_containerregistry` |
| `ARUBACLOUD_SECURITYGROUP_ID` | `arubacloud_securitygroup`, `arubacloud_securityrule` |
| `ARUBACLOUD_SECURITYRULE_ID` | `arubacloud_securityrule` |
| `ARUBACLOUD_KMS_ID` | `arubacloud_kms` |
| `ARUBACLOUD_SCHEDULEJOB_ID` | `arubacloud_schedulejob` |
| `ARUBACLOUD_VPNTUNNEL_ID` | `arubacloud_vpntunnel`, `arubacloud_vpnroute` |
| `ARUBACLOUD_VPNROUTE_ID` | `arubacloud_vpnroute` |
| `ARUBACLOUD_VPCPEERING_ID` | `arubacloud_vpcpeering`, `arubacloud_vpcpeeringroute` |
| `ARUBACLOUD_VPCPEERINGROUTE_ID` | `arubacloud_vpcpeeringroute` |

## GitHub Actions Secrets

To run acceptance tests in CI, add the following [repository secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets):

- `ARUBACLOUD_API_KEY`
- `ARUBACLOUD_API_SECRET`

The `ARUBACLOUD_PROJECT_ID` and resource-specific IDs are not stored as secrets (they are not sensitive) — they can be added as [repository variables](https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables) and referenced as `${{ vars.ARUBACLOUD_PROJECT_ID }}` in the workflow if needed.

## Adding New Acceptance Tests

Each resource has a corresponding `*_resource_test.go` file. The standard pattern is:

```go
func TestAccExampleResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {Config: testAccExampleResourceConfig("name")},
            {ResourceName: "arubacloud_example.test", ImportState: true, ImportStateVerify: true},
            {Config: testAccExampleResourceConfig("name-updated")},
        },
    })
}
```

`testAccPreCheck` (defined in `provider_test.go`) validates that `ARUBACLOUD_API_KEY` and `ARUBACLOUD_API_SECRET` are set before each test case runs.
