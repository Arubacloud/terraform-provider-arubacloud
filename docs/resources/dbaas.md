---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_dbaas"
sidebar_current: "docs-resource-dbaas"
description: |-
  DBaaS provides managed database instances for scalable and reliable data storage.
---

# arubacloud_dbaas

DBaaS allows you to deploy, scale, and manage database instances easily.

## Usage example

```hcl
resource "arubacloud_dbaas" "example" {
  name           = "example-dbaas"
  location       = "ITBG-Bergamo"
  tags           = ["dbaas", "test"]
  project_id     = arubacloud_project.example.id
  engine         = "mysql-8.0"
  zone           = "ITBG-1"
  flavor         = "db.t3.medium"
  storage_size   = 50
  billing_period = "Hour"
  network {
    vpc_id            = arubacloud_vpc.example.id
    subnet_id         = arubacloud_subnet.example.id
    security_group_id = arubacloud_securitygroup.example.id
    elastic_ip_id     = arubacloud_elasticip.example.id
  }
  autoscaling {
    enabled         = true
    available_space = 100
    step_size       = 10
  }
}
```

## Argument reference

* `name` - (Required)[string] The name of the DBaaS instance.
* `location` - (Required)[string] The location for the DBaaS instance.
* `tags` - (Optional)[list(string)] Tags for the DBaaS resource.
* `project_id` - (Required)[string] The project ID.
* `engine` - (Required)[string] Database engine (e.g., "mysql-8.0", "mssql-2022-web").
* `zone` - (Required)[string] Zone (e.g., "ITBG-1").
* `flavor` - (Required)[string] Flavor type.
* `storage_size` - (Required)[int] Storage size in GB.
* `billing_period` - (Required)[string] Billing period.
* `network` - (Required)[object] Network configuration:
  * `vpc_id` - (Required)[string] VPC ID.
  * `subnet_id` - (Required)[string] Subnet ID.
  * `security_group_id` - (Required)[string] Security Group ID.
  * `elastic_ip_id` - (Required)[string] Elastic IP ID.
* `autoscaling` - (Required)[object] Autoscaling configuration:
  * `enabled` - (Required)[bool] Whether autoscaling is enabled.
  * `available_space` - (Required)[int] Available space for autoscaling.
  * `step_size` - (Required)[int] Step size for autoscaling.

## Attribute reference

* `id` - (Computed)[string] The ID of the DBaaS instance.

## Import

To import a DBaaS instance, define an empty resource in your plan:

```
resource "arubacloud_dbaas" "example" {
}
```

Import using the DBaaS ID:

```
terraform import arubacloud_dbaas.example <dbaas_id>
```
