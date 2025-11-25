---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_securitygroup"
sidebar_current: "docs-resource-securitygroup"
description: |-
  Security Group is a virtual firewall for your cloud resources.
---

# arubacloud_securitygroup

Security Groups control inbound and outbound traffic for resources in your VPC.

## Usage example

```hcl
resource "arubacloud_securitygroup" "example" {
  name       = "example-security-group"
  location   = "ITBG-Bergamo"
  tags       = ["web", "prod"]
  project_id = arubacloud_project.example.id
  vpc_id     = arubacloud_vpc.example.id
}
```


## Argument reference

* `name` - (Required)[string] The name of the Security Group.
* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the Security Group.

## Import

To import a Security Group, define an empty resource in your plan:

```
resource "arubacloud_securitygroup" "example" {
}
```

Import using the Security Group ID:

```
terraform import arubacloud_securitygroup.example <securitygroup_id>
```
