---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_dbaasuser"
sidebar_current: "docs-resource-dbaasuser"
description: |-
  DBaaS User provides user accounts for managed database instances.
---

# arubacloud_dbaasuser

DBaaS Users allow you to create and manage database users for DBaaS instances.

## Usage example

```hcl
resource "arubacloud_dbaasuser" "example" {
  dbaas_id = arubacloud_dbaas.example.id
  username = "dbuser"
  password = "supersecretpassword"
}
```

## Argument reference

* `dbaas_id` - (Required)[string] The ID of the DBaaS instance this user belongs to.
* `username` - (Required)[string] The username for the DBaaS user.
* `password` - (Required)[string] The password for the DBaaS user.

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
