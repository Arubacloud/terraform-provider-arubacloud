---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_database"
sidebar_current: "docs-resource-database"
description: |-
  Database resource represents a database instance within a DBaaS cluster.
---

# arubacloud_database

Databases can be created and managed within a DBaaS cluster.

## Usage example

```hcl
resource "arubacloud_database" "example" {
  dbaas_id = arubacloud_dbaas.example.id
  name     = "exampledb"
}
```

## Argument reference

* `dbaas_id` - (Required)[string] The DBaaS instance ID.
* `name` - (Required)[string] The name of the database.
* ...other arguments...

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
