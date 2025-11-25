---
subcategory: "Container"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_kaas"
sidebar_current: "docs-datasource-kaas"
description: |-
  Data source for querying KaaS resources in ArubaCloud.
---

# arubacloud_kaas (Data Source)

Use this data source to retrieve information about a KaaS resource.

## Usage example

```hcl
data "arubacloud_kaas" "example" {
  id = "kaas-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the KaaS to query.

## Attribute reference

* `name` - (Computed)[string] The name of the KaaS cluster.
* `location` - (Computed)[string] The location for the cluster.
* `project_id` - (Computed)[string] The project ID.
* `preset` - (Computed)[bool] Whether to use a preset configuration.
* `vpc_id` - (Computed)[string] VPC ID for the cluster.
* `subnet_id` - (Computed)[string] Subnet ID for the cluster.
* `node_cidr` - (Computed)[object] Node CIDR configuration:
  * `address` - (Computed)[string] Node CIDR address.
  * `subnet_name` - (Computed)[string] Node CIDR subnet name.
* `security_group_name` - (Computed)[string] Security group name.
* `version` - (Computed)[string] Kubernetes version.
* `node_pools` - (Computed)[list(object)] Node pool configuration:
  * `node_pool_name` - (Computed)[string] Name of the node pool.
  * `replicas` - (Computed)[int] Number of nodes in the pool.
  * `type` - (Computed)[string] Instance type for nodes.
  * `zone` - (Computed)[string] Zone for the node pool.
* `ha` - (Computed)[bool] Enable high availability.
* `billing_period` - (Computed)[string] Billing period.
* `tags` - (Computed)[list(string)] Tags for the cluster.
