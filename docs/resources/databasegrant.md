---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_databasegrant"
sidebar_current: "docs-resource-databasegrant"
description: |-
  Database Grant resource manages user privileges for databases.
---

# arubacloud_databasegrant

Database Grants allow you to assign roles and privileges to users for specific databases.

## Usage example

```hcl
resource "arubacloud_databasegrant" "example" {
  database = arubacloud_database.example.id
  user_id  = arubacloud_dbaasuser.example.id
  role     = "admin"
}
```

## Argument reference

* `database` - (Required)[string] The database ID.
* `user_id` - (Required)[string] The user ID.
* `role` - (Required)[string] The role to assign.
* ...other arguments...

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
