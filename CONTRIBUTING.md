# Contributing

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Go | ≥ 1.24 | `go.mod` pins the exact minimum |
| Terraform | any recent | Required only for `make generate` (doc regeneration) |
| golangci-lint | v2+ | `make lint` auto-installs if missing |

**Windows note:** `make generate` needs Terraform in PATH. Git Bash typically does not have it; run from **WSL** instead. `make fmt` and `make lint` also behave more reliably from WSL on Windows.

## Workflow

```bash
# 1. Make your changes

# 2. Format (enforced by linter)
make fmt

# 3. Run unit tests
make test

# 4. Run linter (auto-installs golangci-lint v2 if missing)
make lint

# 5. Regenerate docs after any schema or template change
make generate

# 6. Or run everything in one shot (matches CI):
make ci-test
```

`make ci-test` runs the same sequence as the CI pipeline: build → lint → generate → `go mod tidy` → `git diff --exit-code` → unit tests.

## What CI checks

Three jobs run on every PR:

| Job | What it does | Common failure causes |
|---|---|---|
| `build` | `go build .` + `golangci-lint` | Compile error, gofmt issue, lint warning |
| `generate` | `make generate` then `git diff --exit-code` | Uncommitted doc changes after schema edit |
| `test` | `go test ./...` with coverage | Failing unit test |

A fourth job (`acceptance`) runs automatically on every push to `main` and can be triggered manually from any branch via **Actions → Acceptance Tests → Run workflow**. See [Acceptance Tests](#acceptance-tests) below.

## Acceptance Tests

Acceptance tests exercise the real ArubaCloud API — they create, read, update, and delete actual cloud resources. They are skipped unless `TF_ACC=1` is set.

### Running locally

```bash
export TF_ACC=1
export ARUBACLOUD_API_KEY=<your-api-key>
export ARUBACLOUD_API_SECRET=<your-api-secret>
export ARUBACLOUD_PROJECT_ID=<an-existing-project-id>

# Run all acceptance tests (slow — provisions real infra)
go test -v -timeout=120m ./internal/provider/... -run '^TestAcc'

# Run a single resource
go test -v -timeout=30m ./internal/provider/... -run '^TestAccKeypairResource$'

# Convenience wrapper (loads credentials from examples/test/compute/terraform.tfvars)
./run-acceptance-tests.sh TestAccKeypairResource
```

### CI — manual trigger

Go to **Actions → Acceptance Tests → Run workflow** and optionally fill in:

| Input | Default | Example |
|---|---|---|
| `test_filter` | _(all `^TestAcc`)_ | `TestAccVpcResource` |
| `timeout` | `120m` | `30m` |

### Required environment variables

The following variables must be set as [repository secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets) for CI, or as shell exports when running locally.

#### Always required

| Variable | Purpose |
|---|---|
| `TF_ACC` | Must be `"1"` to activate acceptance tests |
| `ARUBACLOUD_API_KEY` | API authentication |
| `ARUBACLOUD_API_SECRET` | API authentication |

#### Resource tests

Resource tests (Create / Update / Delete lifecycle) only need the three variables above plus:

| Variable | Purpose |
|---|---|
| `ARUBACLOUD_PROJECT_ID` | Project scope for every resource creation |

#### Data source tests

Data source tests look up an **existing** resource by ID. Each data source test skips (rather than fails) when its variable is absent — set only the ones you need.

| Variable | Data source(s) |
|---|---|
| `ARUBACLOUD_PROJECT_ID` | All data sources |
| `ARUBACLOUD_VPC_ID` | vpc, subnet, securitygroup, securityrule, vpcpeering, vpcpeeringroute |
| `ARUBACLOUD_CLOUDSERVER_ID` | cloudserver |
| `ARUBACLOUD_KEYPAIR_ID` | keypair |
| `ARUBACLOUD_BLOCKSTORAGE_ID` | blockstorage |
| `ARUBACLOUD_SNAPSHOT_ID` | snapshot |
| `ARUBACLOUD_ELASTICIP_ID` | elasticip |
| `ARUBACLOUD_BACKUP_ID` | backup, restore |
| `ARUBACLOUD_RESTORE_ID` | restore |
| `ARUBACLOUD_DBAAS_ID` | dbaas, dbaasuser, database, databasegrant |
| `ARUBACLOUD_DATABASE_ID` | database, databasegrant |
| `ARUBACLOUD_DATABASE_BACKUP_ID` | databasebackup |
| `ARUBACLOUD_DBAAS_USERNAME` | dbaasuser |
| `ARUBACLOUD_DBAAS_USER_ID` | databasegrant |
| `ARUBACLOUD_KAAS_ID` | kaas |
| `ARUBACLOUD_CONTAINERREGISTRY_ID` | containerregistry |
| `ARUBACLOUD_SECURITYGROUP_ID` | securitygroup, securityrule |
| `ARUBACLOUD_SECURITYRULE_ID` | securityrule |
| `ARUBACLOUD_KMS_ID` | kms |
| `ARUBACLOUD_SCHEDULEJOB_ID` | schedulejob |
| `ARUBACLOUD_VPNTUNNEL_ID` | vpntunnel, vpnroute |
| `ARUBACLOUD_VPNROUTE_ID` | vpnroute |
| `ARUBACLOUD_VPCPEERING_ID` | vpcpeering, vpcpeeringroute |
| `ARUBACLOUD_VPCPEERINGROUTE_ID` | vpcpeeringroute |

## When to regenerate docs

Run `make generate` whenever you change:
- Provider schema (`internal/provider/provider.go`)
- Any resource or data source schema
- A file under `templates/`

The `generate` CI job will fail with `git diff` output if committed docs don't match the regenerated output.

## Adding a new resource

1. Create `internal/provider/<name>_resource.go` and `internal/provider/<name>_resource_test.go`.
2. Register the resource in `internal/provider/provider.go`.
3. Add example HCL under `examples/resources/<name>/resource.tf`.
4. Add a template `templates/resources/<name>.md.tmpl` (or let tfplugindocs generate a default).
5. Run `make generate`.

The `ai/` directory has detailed guidance for each task type — see [`CLAUDE.md`](CLAUDE.md) for the index.

## Error handling conventions

New resource/data-source handlers should follow the existing pattern for API errors:

- Use `CheckResponse` + `IsNotFound` (from `internal/provider/provider_error.go`) for response-level errors.
- Use `LogAndAppendAPIError` (from `internal/provider/error_logging.go`) to emit a `tflog.Error` before adding the user-facing diagnostic — this ensures failures are visible under `TF_LOG=DEBUG` without requiring `log_level = "DEBUG"`.
- 404 → `resp.State.RemoveResource(ctx)` (never return an error for missing resources during Read).

Reference implementation: `internal/provider/databasegrant_resource.go`.

## Debugging API calls

See [Logging & Troubleshooting](docs/index.md#logging--troubleshooting) in the provider docs, or the [Debugging & Logging](README.md#debugging--logging) section in the README.

Short version — set **both** filters and use a single-line invocation:

```bash
TF_LOG=DEBUG TF_LOG_PATH=./trace.log terraform plan
```

With `log_level = "DEBUG"` in the provider block (or `ARUBACLOUD_LOG_LEVEL=DEBUG`).
