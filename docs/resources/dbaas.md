---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_dbaas"
sidebar_current: "docs-resource-dbaas"
description: |-
  DBaaS (Database as a Service) provides managed database instances in ArubaCloud.
---

# arubacloud_dbaas

DBaaS allows you to deploy, scale, and manage cloud databases easily.

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
* `project_id` - (Required)[string] The project ID.
* `engine` - (Required)[string] The database engine.
* ...other arguments...

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
