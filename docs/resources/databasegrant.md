---
page_title: "arubacloud_databasegrant Resource - ArubaCloud"
subcategory: "Database"
description: |-
  Manages an ArubaCloud Database Grant.
---

# arubacloud_databasegrant (Resource)

Manages an ArubaCloud Database Grant.

```terraform
resource "arubacloud_databasegrant" "example" {
  name       = "example-database-grant"
  database   = "example-database"
  user       = "example-user"
  privileges = ["SELECT", "INSERT"]
  project_id = "example-project"
}
```



## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_databasegrant.example <databasegrant_id>
```