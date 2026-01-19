---
page_title: "arubacloud_databasegrant Data Source - terraform-provider-arubacloud"
subcategory: "Database"
description: |-
  Retrieves information about an ArubaCloud Database Grant.
---

# arubacloud_databasegrant (Data Source)

Retrieves information about an ArubaCloud Database Grant (user permissions for a database).

## Example Usage

```terraform
data "arubacloud_databasegrant" "example" {
  project_id = "project-123"
  dbaas_id   = "dbaas-456"
  database   = "mydb"
  user_id    = "myuser"
}
```

## Schema

### Required

- `project_id` (String) ID of the project
- `dbaas_id` (String) DBaaS instance ID
- `database` (String) Database name
- `user_id` (String) Database user ID (username)

### Read-Only

- `id` (String) Database Grant identifier (composite key)
- `uri` (String) Database Grant URI
- `role` (String) Granted role. One of: `liteadmin`, `readwrite`, `readonly`


