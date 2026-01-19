---
page_title: "arubacloud_databasegrant Resource - terraform-provider-arubacloud"
subcategory: "Database"
description: |-
  Manages database user permissions (grants) for an ArubaCloud DBaaS instance.
---

# arubacloud_databasegrant (Resource)

Manages database user permissions (grants) for an ArubaCloud DBaaS instance. This resource associates a database user with a database and assigns a specific role that defines the user's permissions.

## Example Usage

```terraform
resource "arubacloud_databasegrant" "example" {
  project_id = arubacloud_project.example.id
  dbaas_id   = arubacloud_dbaas.example.id
  database   = arubacloud_database.example.id
  user_id    = arubacloud_dbaasuser.example.id
  role       = "readwrite"  # Options: liteadmin, readwrite, readonly
}
```

## Schema

### Required

- `project_id` (String) ID of the project this grant belongs to
- `dbaas_id` (String) DBaaS instance ID
- `database` (String) Database name or ID
- `user_id` (String) Database user ID (username)
- `role` (String) Role to grant. Must be one of: `liteadmin`, `readwrite`, `readonly`

### Read-Only

- `id` (String) Database Grant identifier (composite key)
- `uri` (String) Database Grant URI

## Import

Aruba Cloud Database Grant can be imported using a composite key format: `project_id/dbaas_id/database/user_id`

```shell
terraform import arubacloud_databasegrant.example project-123/dbaas-456/mydb/myuser
```