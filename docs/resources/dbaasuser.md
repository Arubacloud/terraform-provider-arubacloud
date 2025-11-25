---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_dbaasuser"
sidebar_current: "docs-resource-dbaasuser"
description: |-
  DBaaS User resource manages database users for DBaaS clusters.
---

# arubacloud_dbaasuser

DBaaS Users can be created and managed for database access.

## Usage example

```hcl
resource "arubacloud_dbaasuser" "example" {
  dbaas_id = arubacloud_dbaas.example.id
  username = "dbuser"
  password = "supersecretpassword"
}
```

## Argument reference

* `dbaas_id` - (Required)[string] The DBaaS instance ID.
* `username` - (Required)[string] The username for the database user.
* `password` - (Required)[string] The password for the database user.
* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the DBaaS user.

## Import

To import a DBaaS user, define an empty resource in your plan:

```
resource "arubacloud_dbaasuser" "example" {
}
```

Import using the DBaaS user ID:

```
terraform import arubacloud_dbaasuser.example <dbaasuser_id>
```
