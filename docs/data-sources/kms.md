---
subcategory: "Security"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_kms"
sidebar_current: "docs-datasource-kms"
description: |-
  Data source for querying KMS (Key Management Service) resources in ArubaCloud.
---

# arubacloud_kms (Data Source)

Use this data source to retrieve information about a KMS instance.

## Usage example

```hcl
data "arubacloud_kms" "example" {
  id = "kms-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the KMS instance to query.

## Attribute reference

* `id` - (Computed)[string] The ID of the KMS instance.
* `name` - (Computed)[string] The name of the KMS instance.
* `description` - (Computed)[string] The description of the KMS instance.
* `endpoint` - (Computed)[string] The endpoint of the KMS instance.
