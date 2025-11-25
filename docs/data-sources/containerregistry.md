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

* `id` - (Required)[string] The ID of the container registry to query.

## Attribute reference

* `name` - (Computed)[string] The name of the container registry.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
