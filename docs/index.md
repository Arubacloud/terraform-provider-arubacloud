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

This emits each outbound HTTP request (method, URL, headers, body) and response (status, headers, body) to stderr. The `Authorization` header is automatically redacted as `Bearer [REDACTED]`.

> **Warning**: request/response bodies may contain sensitive data. Do not commit debug logs to version control.

### Suppress SDK output while keeping provider tracing

Set `log_level = "OFF"` (the default) or omit it entirely. The 154 provider-level `tflog` calls (resource lifecycle events, wait/retry status) are unaffected by `log_level` and continue to respect `TF_LOG` alone.

