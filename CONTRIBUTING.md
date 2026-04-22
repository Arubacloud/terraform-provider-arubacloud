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
