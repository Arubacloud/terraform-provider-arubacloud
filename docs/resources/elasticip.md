---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_elasticip"
sidebar_current: "docs-resource-elasticip"
description: |-
  Elastic IP is a static, public IPv4 address designed for dynamic cloud computing.
---

# arubacloud_elasticip

Elastic IP addresses can be associated with resources in your ArubaCloud account.

## Usage example

```hcl
resource "arubacloud_elasticip" "example" {
  name           = "example-elasticip"
  location       = "ITBG-Bergamo"
  tags           = ["public", "test"]
  billing_period = "hourly"
  project_id     = arubacloud_project.example.id
}
```


## Argument reference

* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the Elastic IP.
* `address` - (Computed)[string] The Elastic IP address.

## Import

To import an Elastic IP, define an empty resource in your plan:

```
resource "arubacloud_elasticip" "example" {
}
```

Import using the Elastic IP ID:

```
terraform import arubacloud_elasticip.example <elasticip_id>
```
