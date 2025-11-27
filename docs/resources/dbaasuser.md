---
page_title: "arubacloud_dbaasuser Resource - ArubaCloud"
subcategory: "Database"
description: |-
  Manages an ArubaCloud DBaaS User.
---

# arubacloud_dbaasuser

Manages an ArubaCloud DBaaS User.

```terraform
resource "arubacloud_dbaasuser" "example" {
  dbaas_id = "example-dbaas-id"
  username = "example-user"
  password = "example-password"
}
```


## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_dbaasuser.example <dbaasuser_id>
```
