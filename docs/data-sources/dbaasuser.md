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

* `id` - (Required)[string] The ID of the DBaaS user to query.

## Attribute reference

* `username` - (Computed)[string] The username of the DBaaS user.
* `dbaas_id` - (Computed)[string] The DBaaS ID.
* ...other attributes...
