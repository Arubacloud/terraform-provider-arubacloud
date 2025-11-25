---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_dbaas"
sidebar_current: "docs-datasource-dbaas"
description: |-
  Data source for querying DBaaS resources in ArubaCloud.
---

# arubacloud_dbaas (Data Source)

Use this data source to retrieve information about a DBaaS resource.

## Usage example

```hcl
data "arubacloud_dbaas" "example" {
  id = "dbaas-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the DBaaS to query.

## Attribute reference

* `name` - (Computed)[string] The name of the DBaaS.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
