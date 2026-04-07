# DEVEX.md — Developer Experience

## Make Commands

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

## Acceptance Tests

Require environment variables `ARUBACLOUD_API_KEY` and `ARUBACLOUD_API_SECRET`, or use `run-acceptance-tests.sh` to load them from `terraform.tfvars`.

## Local Terraform Development

Set `TF_CLI_CONFIG_FILE="terraform.tfrc"` with a provider override pointing to the built binary.

## Docs Generation

Always run `make generate` after any schema changes. It runs `tfplugindocs` then post-processes output with `scripts/format-docs.sh` (bash) / `scripts/format-docs.ps1` (PowerShell) to separate Arguments from Attributes sections.
