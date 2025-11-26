
---
page_title: "arubacloud_databasegrant Data Source - ArubaCloud"
subcategory: "Database"
description: |-
  Retrieves an ArubaCloud Database Grant.
---

# arubacloud_databasegrant (Data Source)

Retrieves an ArubaCloud Database Grant.

```terraform
data "arubacloud_database_grant" "example" {
  name       = "example-database-grant"
  project_id = "example-project"
  database   = "example-db"
  privileges = ["SELECT", "INSERT"]
}
```

## Schema

<no value>
