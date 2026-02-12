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

