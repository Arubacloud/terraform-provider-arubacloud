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
* `location` - (Computed)[string] The location of the DBaaS.
* `tags` - (Computed)[list(string)] Tags for the DBaaS.
* `project_id` - (Computed)[string] The project ID.
* `engine` - (Computed)[string] Database engine (mysql-8.0, mssql-2022-web, mssql-2022-standard, mssql-2022-enterprise).
* `zone` - (Computed)[string] Zone (ITBG-1, ITBG-2, ITBG-3).
* `flavor` - (Computed)[string] Flavor type.
* `storage_size` - (Computed)[int] Storage size.
* `billing_period` - (Computed)[string] Billing period.
* `network` - (Computed)[object]
  * `vpc_id` - (Computed)[string] VPC ID.
  * `subnet_id` - (Computed)[string] Subnet ID.
  * `security_group_id` - (Computed)[string] Security Group ID.
  * `elastic_ip_id` - (Computed)[string] Elastic IP ID.
* `autoscaling` - (Computed)[object]
  * `enabled` - (Computed)[bool] Autoscaling enabled.
  * `available_space` - (Computed)[int] Available space for autoscaling.
  * `step_size` - (Computed)[int] Step size for autoscaling.
