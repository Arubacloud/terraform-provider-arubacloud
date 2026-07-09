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
- Terraform CLI **or** OpenTofu CLI in `PATH` (acceptance tests drive whichever binary you point them at)
- A valid ArubaCloud account with API credentials
- An existing project to scope resource creation (`ARUBACLOUD_PROJECT_ID`)

## Running Locally

```bash
export TF_ACC=1
export ARUBACLOUD_CLIENT_ID=<your-api-key>
export ARUBACLOUD_CLIENT_SECRET=<your-api-secret>
export ARUBACLOUD_PROJECT_ID=<an-existing-project-id>

# Run all acceptance tests against Terraform (default)
go test -v -timeout=120m ./internal/provider/... -run '^TestAcc'

# Run all acceptance tests against OpenTofu
TF_ACC_TERRAFORM_PATH=$(which tofu) go test -v -timeout=120m ./internal/provider/... -run '^TestAcc'

# Run a single resource test
go test -v -timeout=30m ./internal/provider/... -run '^TestAccKeypairResource$'
```

`TF_ACC_TERRAFORM_PATH` tells the test harness which binary to invoke. Point it at `terraform` or `tofu` interchangeably — the provider is compatible with both.

A convenience wrapper is available that accepts credentials and fixture IDs as flags:

```bash
# Minimal: run a single test
./run-acceptance-tests.sh \
  --client-id "$ARUBACLOUD_CLIENT_ID" \
  --client-secret "$ARUBACLOUD_CLIENT_SECRET" \
  --project-id "$ARUBACLOUD_PROJECT_ID" \
  -t '^TestAccKeypairResource$'

# With an optional fixture for DBaaS-dependent tests
./run-acceptance-tests.sh \
  --client-id "$ARUBACLOUD_CLIENT_ID" \
  --client-secret "$ARUBACLOUD_CLIENT_SECRET" \
  --project-id "$ARUBACLOUD_PROJECT_ID" \
  --dbaas-id "$ARUBACLOUD_DBAAS_ID" \
  -t '^TestAccDatabaseDataSource$'

# Full flag reference
./run-acceptance-tests.sh --help
```

## CI — Manual Trigger

Two workflows are available, both triggerable manually from any branch:

- **Terraform**: [Acceptance Tests](https://github.com/Arubacloud/terraform-provider-arubacloud/actions/workflows/acceptance.yml) — runs tests with the `terraform` binary.
- **OpenTofu**: [Acceptance Tests (OpenTofu)](https://github.com/Arubacloud/terraform-provider-arubacloud/actions/workflows/acceptance-opentofu.yml) — runs the same tests with the `tofu` binary.

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
| `ARUBACLOUD_CLIENT_ID` | OAuth2 Client ID for API authentication |
| `ARUBACLOUD_CLIENT_SECRET` | OAuth2 Client Secret for API authentication |

### Resource Tests

Resource tests exercise the full Create → Read → Update → Delete lifecycle. Most only need:

| Variable | Purpose |
|---|---|
| `ARUBACLOUD_PROJECT_ID` | Scopes all resource creation to an existing project |

Some resource tests have additional prerequisites:

| Variable | Required by |
|---|---|
| `ARUBACLOUD_OS_IMAGE_ID` | `TestAccBlockStorageResource_Bootable`, `TestAccCloudserverResource` — OS image used to create a bootable disk |
| `ARUBACLOUD_DBAAS_ID` | `TestAccDatabaseResource`, `TestAccDatabasebackupResource`, `TestAccDatabasegrantResource`, `TestAccDbaasuserResource` — existing DBaaS cluster used as a prerequisite to avoid zone-capacity conflicts when multiple DBaaS tests run sequentially |
| `ARUBACLOUD_BACKUP_ID` | `TestAccRestoreResource` — existing block storage backup to restore from |
| `ARUBACLOUD_VPNTUNNEL_ID` | `TestAccVpnrouteResource` — existing VPN tunnel to attach routes to |

### Data Source Tests

All data source tests create their own infrastructure inline and destroy it on completion — only `ARUBACLOUD_PROJECT_ID` is required for most. Two additional variables are needed for tests that provision a cloud server as a prerequisite, and two VPN tests are exceptions that read pre-existing fixtures because VPN tunnels are too complex to provision inline.

| Variable | Required by | Notes |
|---|---|---|
| `ARUBACLOUD_PROJECT_ID` | All data sources | |
| `ARUBACLOUD_OS_IMAGE_ID` | `arubacloud_cloudserver`, `arubacloud_schedulejob` | OS image slug for bootable disk (e.g. `ubuntu-22.04`) |
| `ARUBACLOUD_VPNTUNNEL_ID` | `arubacloud_vpntunnel`, `arubacloud_vpnroute` | Pre-existing VPN tunnel — inline provisioning not feasible |
| `ARUBACLOUD_VPNROUTE_ID` | `arubacloud_vpnroute` | Pre-existing VPN route within the above tunnel |

## GitHub Actions Secrets

To run acceptance tests in CI, add the following [repository secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets):

- `ARUBACLOUD_CLIENT_ID`
- `ARUBACLOUD_CLIENT_SECRET`

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

`testAccPreCheck` (defined in `provider_test.go`) validates that `ARUBACLOUD_CLIENT_ID` and `ARUBACLOUD_CLIENT_SECRET` are set before each acceptance test case runs.
