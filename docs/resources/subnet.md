---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_subnet"
sidebar_current: "docs-resource-subnet"
description: |-
  Subnet is a range of IP addresses in your VPC.
---

# arubacloud_subnet

A subnet divides your VPC into smaller, isolated networks.

## Usage example

```hcl
resource "arubacloud_subnet" "example" {
  name       = "example-subnet"
  location   = "ITBG-Bergamo"
  tags       = ["subnet", "test"]
  project_id = arubacloud_project.example.id
  vpc_id     = arubacloud_vpc.example.id
  type       = "Advanced"
  network = {
    address = "10.0.1.0/24"
  }
  dhcp = {
    enabled = true
    range = {
      start = "10.0.1.10"
      count = 20
    }
  }
  routes = [
    {
      address = "0.0.0.0"
      gateway = "10.0.1.1"
    }
  ]
  dns = ["8.8.8.8", "8.8.4.4"]
}
```


## Argument reference

* `cidr_block` - (Required)[string] The CIDR block for the Subnet.
* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the Subnet.

## Import

To import a Subnet, define an empty resource in your plan:

```
resource "arubacloud_subnet" "example" {
}
```

Import using the Subnet ID:

```
terraform import arubacloud_subnet.example <subnet_id>
```
