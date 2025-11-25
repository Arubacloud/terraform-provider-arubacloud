---
subcategory: "Container"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_kaas"
sidebar_current: "docs-datasource-kaas"
description: |-
  Data source for querying KaaS (Kubernetes as a Service) resources in ArubaCloud.
---

# arubacloud_kaas (Data Source)

Use this data source to retrieve information about a KaaS cluster.

## Usage example

```hcl
data "arubacloud_kaas" "example" {
  id = "kaas-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the KaaS cluster to query.

## Attribute reference

* `name` - (Computed)[string] The name of the KaaS cluster.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
