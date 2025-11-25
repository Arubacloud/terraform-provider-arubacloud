---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_securityrule"
sidebar_current: "docs-resource-securityrule"
description: |-
  Security Rule defines traffic rules for Security Groups.
---

# arubacloud_securityrule

Security Rules specify allowed or denied traffic for resources in a Security Group.

## Usage example

```hcl
resource "arubacloud_securityrule" "example" {
  name              = "example-security-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.example.id
  vpc_id            = arubacloud_vpc.example.id
  security_group_id = arubacloud_securitygroup.example.id
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "80"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}
```


## Argument reference

* `name` - (Required)[string] The name of the Security Rule.
* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the Security Rule.

## Import

To import a Security Rule, define an empty resource in your plan:

```
resource "arubacloud_securityrule" "example" {
}
```

Import using the Security Rule ID:

```
terraform import arubacloud_securityrule.example <securityrule_id>
```
