---
subcategory: "Container"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_kaas"
sidebar_current: "docs-resource-kaas"
description: |-
  KaaS (Kubernetes as a Service) provides managed Kubernetes clusters in ArubaCloud.
---

# arubacloud_kaas

KaaS allows you to deploy, scale, and manage Kubernetes clusters easily.

## Usage example

```hcl
resource "arubacloud_kaas" "example" {
  name       = "example-kaas"
  location   = "ITBG-Bergamo"
  tags       = ["k8s", "test"]
  project_id = arubacloud_project.example.id
  preset     = true
  vpc_id     = arubacloud_vpc.example.id
  subnet_id  = arubacloud_subnet.example.id
  node_cidr {
    address     = "10.0.2.0/24"
    subnet_name = "kaas-subnet"
  }
  security_group_name = arubacloud_securitygroup.example.name
  version             = "1.32.2"
  node_pools = [
    {
      node_pool_name = "pool-1"
      replicas       = 2
      type           = "c2.medium"
      zone           = "ITBG-1"
    },
    {
      node_pool_name = "pool-2"
      replicas       = 1
      type           = "c2.large"
      zone           = "ITBG-2"
    }
  ]
  ha             = true
  billing_period = "Hour"
}
```

## Argument reference

* `name` - (Required)[string] The name of the KaaS cluster.
* `project_id` - (Required)[string] The project ID.
* `vpc_id` - (Required)[string] The VPC ID.
* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the KaaS cluster.

## Import

To import a KaaS cluster, define an empty resource in your plan:

```
resource "arubacloud_kaas" "example" {
}
```

Import using the KaaS ID:

```
terraform import arubacloud_kaas.example <kaas_id>
```
