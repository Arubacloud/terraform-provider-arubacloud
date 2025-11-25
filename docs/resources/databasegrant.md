---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_databasegrant"
sidebar_current: "docs-resource-databasegrant"
description: |-
  Database Grant provides access control for users on databases.
---

# arubacloud_databasegrant

Database Grants allow you to assign roles to users for specific databases.

## Usage example

```hcl
resource "arubacloud_databasegrant" "example" {
  database = arubacloud_database.example.id
  user_id  = arubacloud_dbaasuser.example.id
  role     = "admin"
}
```

## Argument reference

* `database` - (Required)[string] The ID of the database to grant access to.
* `user_id` - (Required)[string] The ID of the user to grant access.
* `role` - (Required)[string] The role to assign (e.g., "admin", "read", "write").

## Attribute reference

* `id` - (Computed)[string] The ID of the database grant.

## Import

To import a database grant, define an empty resource in your plan:

```
resource "arubacloud_databasegrant" "example" {
}
```

Import using the database grant ID:

```
terraform import arubacloud_databasegrant.example <databasegrant_id>
```
