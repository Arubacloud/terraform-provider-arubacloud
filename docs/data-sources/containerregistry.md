---
subcategory: "Container"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_containerregistry"
sidebar_current: "docs-datasource-containerregistry"
description: |-
  Data source for querying Container Registry resources in ArubaCloud.
---

# arubacloud_containerregistry (Data Source)

Use this data source to retrieve information about a Container Registry resource.

## Usage example

```hcl
data "arubacloud_containerregistry" "example" {
  id = "containerregistry-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the Container Registry to query.

## Attribute reference

* `name` - (Computed)[string] The name of the Container Registry.
* `location` - (Computed)[string] The location of the Container Registry.
* `project_id` - (Computed)[string] The project ID.
* `elasticip_id` - (Computed)[string] Elastic IP ID.
* `subnet_id` - (Computed)[string] Subnet ID.
* `security_group_id` - (Computed)[string] Security Group ID.
* `block_storage_id` - (Computed)[string] Block Storage ID.
* `billing_period` - (Computed)[string] Billing period.
* `admin_user` - (Computed)[string] Admin user for the Container Registry.
* `tags` - (Computed)[list(string)] Tags for the Container Registry.
