# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make              # Full pipeline: fmt → lint → test → build → generate
make build        # Build provider binary
make install      # Build and install provider
make fmt          # Format code with gofmt
make lint         # Run golangci-lint (auto-installs v2 if missing)
make test         # Unit tests (fast, no external deps)
make testacc      # Acceptance tests (requires TF_ACC=1 and real API credentials)
make testacc-run TEST=TestAccCloudServer  # Run specific acceptance test by name
make testcov      # Generate HTML coverage report
make generate     # Regenerate docs
make ci-test      # Full CI pipeline locally (build, lint, generate, test, mod tidy)
```

Acceptance tests require environment variables `ARUBACLOUD_API_KEY` and `ARUBACLOUD_API_SECRET`, or use `run-acceptance-tests.sh` to load them from `terraform.tfvars`.

For local Terraform development, set `TF_CLI_CONFIG_FILE="terraform.tfrc"` with a provider override pointing to the built binary.

## Architecture

This is a Terraform provider built with [HashiCorp Terraform Plugin Framework v1](https://github.com/hashicorp/terraform-plugin-framework) (Protocol v6). All provider logic lives in `internal/provider/`.

### Provider entry point

`main.go` → `internal/provider/provider.go`

`ArubaCloudProvider` wraps an `ArubaCloudClient` (containing `sdk-go` client + `ResourceTimeout`). Provider config attributes: `api_key`, `api_secret`, `resource_timeout` (default 10m), `base_url`, `token_issuer_url`.

### Resource/DataSource pattern

Every resource follows the same structure:
1. **Model struct** (`CloudServerResourceModel`) — Terraform state mapped to Go types using `types.String`, `types.Int64`, `types.Bool`, `types.Object`, `types.List`
2. **Resource struct** (`CloudServerResource`) — holds `*ArubaCloudClient`
3. Implements `resource.Resource` + `resource.ResourceWithImportState`
4. Methods: `Metadata()`, `Schema()`, `Create()`, `Read()`, `Update()`, `Delete()`, `ImportState()`

Data sources follow the same pattern but only implement `datasource.DataSource` with a `Read()` method.

Nested objects use dedicated model types + `schema.SingleNestedObjectAttribute()`. Computed-only fields use `stringplanmodifier.UseStateForUnknown()`.

### Shared utilities

- `resource_wait.go` — `WaitForResourceActive()` polls every 5s until the resource leaves a transitional state (InCreation, Creating, Updating, Deleting, Pending, Provisioning). Respects `ResourceTimeout`.
- `error_helper.go` — `FormatAPIError()` extracts structured validation errors from the API response's `Extensions` field and logs full JSON for debugging.

### Resource coverage

24 resources across: Compute (CloudServer, Keypair, ElasticIP, Project), Storage (BlockStorage, Snapshot, Backup, Restore), Networking (VPC, Subnet, SecurityGroup, SecurityRule, VPCPeering, VPCPeeringRoute, VPNTunnel, VPNRoute), Container (ContainerRegistry, KaaS), Database (DBaaS, Database, DatabaseGrant, DatabaseBackup, DBaaSUser), Security (KMS). Two resources (Key, KMIP) are disabled due to SDK limitations.

### Testing

- Unit tests: schema validation and logic, no API calls — run with `make test`
- Acceptance tests: full CRUD against real infrastructure, in `*_test.go` files using `resource.TestCase` with `ProtoV6ProviderFactories`, assertions via `statecheck.ExpectKnownValue()` + `tfjsonpath`

### Docs generation

`make generate` runs `tfplugindocs` then post-processes output with `scripts/format-docs.sh` (bash) / `scripts/format-docs.ps1` (PowerShell) to separate Arguments from Attributes sections. Always run this after schema changes.
