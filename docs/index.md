---
page_title: "Provider: Aruba Cloud"
description: |-
  The Aruba Cloud terraform provider is used to interact with the resources supported by Aruba Cloud.
---

# Aruba Cloud Provider

The Aruba Cloud terraform provider is used to interact with the resources supported by [Aruba Cloud](https://arubacloud.com). 
You need to configure the provider with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```terraform
terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = ">= 0.0.1"
    }
  }
}

provider "arubacloud" {
  api_key    = "YOUR_API_KEY"
  api_secret = "YOUR_API_SECRET"
}
```


## Argument Reference

The following arguments are supported:

- `api_key` - (Required, string) ArubaCloud API key. Can also be specified with the `ARUBACLOUD_API_KEY` environment variable.
- `api_secret` - (Required, string) ArubaCloud API secret. Can also be specified with the `ARUBACLOUD_API_SECRET` environment variable.
- `resource_timeout` - (Optional, string) Timeout for waiting for resources to become active after creation (e.g. `"5m"`, `"10m"`). Default: `"10m"`.
- `base_url` - (Optional, string) Override the ArubaCloud API base URL. Advanced use only.
- `token_issuer_url` - (Optional, string) Override the ArubaCloud token issuer URL. Advanced use only.
- `log_level` - (Optional, string) SDK log level for HTTP request/response tracing. Accepted values (case-insensitive): `OFF`, `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE`. Default: `OFF`. Can also be set via the `ARUBACLOUD_LOG_LEVEL` environment variable; the HCL attribute takes precedence.

## Logging & Troubleshooting

The provider exposes two independent log filters:

| Filter | Controls |
|---|---|
| `log_level` / `ARUBACLOUD_LOG_LEVEL` | What the SDK HTTP client forwards to Terraform logs |
| `TF_LOG` / `TF_LOG_PROVIDER` | What Terraform actually writes to stderr |

A message is visible only when **both** filters permit it. SDK messages are tagged with the `arubacloud-sdk` subsystem and can be targeted specifically with `TF_LOG_PROVIDER_ARUBACLOUD_SDK`.

### Enable full HTTP tracing

```hcl
provider "arubacloud" {
  api_key    = var.api_key
  api_secret = var.api_secret
  log_level  = "DEBUG"
}
```

```sh
TF_LOG=DEBUG terraform apply
# or target only SDK output:
TF_LOG_PROVIDER_ARUBACLOUD_SDK=DEBUG terraform apply
```

To route trace output to a file (keeps terminal output readable):

```sh
TF_LOG=DEBUG TF_LOG_PATH=./trace.log terraform plan
```

Filter the captured file to SDK HTTP lines only:

```sh
grep '@module=arubacloud.arubacloud-sdk' trace.log
```

This emits each outbound HTTP request (method, URL, headers, body) and response (status, headers, body) to stderr or the log file. The `Authorization` header is automatically redacted as `Bearer [REDACTED]`.

> **Warning**: request/response bodies may contain sensitive data. Do not commit debug logs to version control.

### Common pitfalls

- **`log_level = "DEBUG"` set but no output appears** — `log_level` is filter #1 (SDK → Terraform). Filter #2 (Terraform → stderr/file) is controlled by `TF_LOG`. Both must be set. Add `TF_LOG=DEBUG` (or `TF_LOG_PROVIDER_ARUBACLOUD_SDK=DEBUG`) to the same command.
- **Env vars set on one line, `terraform` run on the next** — In bash, the `VAR=value command` prefix form only applies to the command on that line. Setting them on a blank line is a no-op. Use a single-line invocation (`TF_LOG=DEBUG terraform plan`) or `export TF_LOG=DEBUG` first.
- **Want SDK silent but still see provider events** — Set `log_level = "OFF"` (the default) or omit it. Provider-level `tflog` calls (resource lifecycle events, wait/retry status) are unaffected by `log_level` and continue to respect `TF_LOG` alone.

### Suppress SDK output while keeping provider tracing

Set `log_level = "OFF"` (the default) or omit it entirely. The provider-level `tflog` calls (resource lifecycle events, wait/retry status) are unaffected by `log_level` and continue to respect `TF_LOG` alone.

