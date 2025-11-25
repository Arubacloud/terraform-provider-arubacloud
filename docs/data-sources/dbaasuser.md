---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_dbaasuser"
sidebar_current: "docs-datasource-dbaasuser"
description: |-
  Data source for querying DBaaS User resources in ArubaCloud.
---

# arubacloud_dbaasuser (Data Source)

Use this data source to retrieve information about a DBaaS User resource.

## Usage example

```hcl
data "arubacloud_dbaasuser" "example" {
  id = "dbaasuser-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the DBaaS User to query.

## Attribute reference

* `dbaas_id` - (Computed)[string] The DBaaS ID this user belongs to.
* `username` - (Computed)[string] The username for the DBaaS user.
* `password` - (Computed, Sensitive)[string] The password for the DBaaS user.
