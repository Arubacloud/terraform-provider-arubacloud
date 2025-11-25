---
subcategory: "Security"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_kmip"
sidebar_current: "docs-datasource-kmip"
description: |-
  Data source for querying KMIP (Key Management Interoperability Protocol) resources in ArubaCloud.
---

# arubacloud_kmip (Data Source)

Use this data source to retrieve information about a KMIP instance.

## Usage example

```hcl
data "arubacloud_kmip" "example" {
  id = "kmip-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the KMIP instance to query.

## Attribute reference

* `name` - (Computed)[string] The name of the KMIP instance.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
