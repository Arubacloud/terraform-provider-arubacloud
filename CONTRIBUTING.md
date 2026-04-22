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

## Required: wait for deletion confirmation in every Delete handler

**A DELETE returning 2xx means the API has accepted the request, not that the resource is gone.**
ArubaCloud deletes infrastructure asynchronously. If Terraform removes a resource from state as soon as the 2xx arrives, dependent resources (e.g. subnets referencing a VPC that is still deleting) will fail on their own destroy with dependency errors.

Every Delete handler for a resource that creates real infrastructure **must** call `WaitForResourceDeleted` after the DELETE API call succeeds:

```go
// After the DELETE call succeeds:
if waitErr := WaitForResourceDeleted(ctx, func(ctx context.Context) (bool, error) {
    getResp, getErr := r.client.Client.From<Service>().<Resource>().Get(ctx, <ids...>, nil)
    if getErr != nil {
        return false, NewTransportError("get", "<Resource>", getErr)
    }
    if provErr := CheckResponse("get", "<Resource>", getResp); provErr != nil {
        if IsNotFound(provErr) {
            return true, nil  // confirmed gone
        }
        return false, provErr
    }
    return false, nil  // still exists, keep polling
}, "<Resource>", resourceID, r.client.ResourceTimeout); waitErr != nil {
    resp.Diagnostics.AddError("Error waiting for <Resource> deletion", waitErr.Error())
    return
}
```

The polling loop (defined in `internal/provider/resource_wait.go:WaitForResourceDeleted`) fires every 10 s, tolerates up to 3 consecutive GET errors, and returns `ErrWaitTimeout` if the resource is not gone within `resource_timeout` (default 10 min, configurable in the provider block).

Reference implementations: `internal/provider/vpc_resource.go:503`, `internal/provider/kaas_resource.go:1358`.

Resources where this pattern is currently missing are tracked in [`ai/TECH_DEBT.md`](ai/TECH_DEBT.md) as **TD-035**.

## Error handling conventions

- Use `CheckResponse` + `IsNotFound` (from `internal/provider/provider_error.go`) for response-level errors.
- Use `LogAndAppendAPIError` (from `internal/provider/error_logging.go`) to emit a `tflog.Error` before adding the user-facing diagnostic — makes failures visible under `TF_LOG=DEBUG`.
- 404 during Read → `resp.State.RemoveResource(ctx)`, never return an error.

Reference implementation: `internal/provider/databasegrant_resource.go`.

## Debugging API calls

Set `log_level = "DEBUG"` in the provider block and open Terraform's log pipeline on the same command line:

```bash
TF_LOG=DEBUG TF_LOG_PATH=./trace.log terraform plan
```

See [`docs/index.md#logging--troubleshooting`](docs/index.md#logging--troubleshooting) for the full two-filter model and common pitfalls.
