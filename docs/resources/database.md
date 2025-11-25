---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_database"
sidebar_current: "docs-resource-database"
description: |-
  Database provides individual databases within a managed DBaaS instance.
---

# arubacloud_database

Databases allow you to create and manage individual databases within a DBaaS instance.

## Usage example

```hcl
resource "arubacloud_database" "example" {
  dbaas_id = arubacloud_dbaas.example.id
  name     = "exampledb"
}
```

## Argument reference

* `dbaas_id` - (Required)[string] The ID of the DBaaS instance this database belongs to.
* `name` - (Required)[string] The name of the database.

## Attribute reference

* `id` - (Computed)[string] The ID of the database.

## Import

To import a database, define an empty resource in your plan:

```
resource "arubacloud_database" "example" {
}
```

Import using the database ID:

```
terraform import arubacloud_database.example <database_id>
```
