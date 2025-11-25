
---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_vpc"
sidebar_current: "docs-resource-vpc"
description: |-
  VPC (Virtual Private Cloud) is a logically isolated section of the ArubaCloud cloud where you can launch resources in a virtual network.
---

# arubacloud_vpc

A Virtual Private Cloud (VPC) allows you to define a custom network topology for your cloud resources.

## Usage example

```hcl
resource "arubacloud_vpc" "example" {
  name     = "example-vpc"
  location = "ITBG-Bergamo"
  tags     = ["network", "test"]
}
```


## Argument reference

* `name` - (Required)[string] The name of the VPC.
* `cidr_block` - (Required)[string] The CIDR block for the VPC.
* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the VPC.

## Import

To import a VPC, define an empty VPC resource in your plan:

```
resource "arubacloud_vpc" "example" {
}
```

Import using the VPC ID:

```
terraform import arubacloud_vpc.example <vpc_id>
```
