---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_securityrule"
sidebar_current: "docs-datasource-securityrule"
description: |-
  Data source for querying Security Rule resources in ArubaCloud.
---

# arubacloud_securityrule (Data Source)

Use this data source to retrieve information about a Security Rule resource.

## Usage example

```hcl
data "arubacloud_securityrule" "example" {
  id = "securityrule-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the security rule to query.

## Attribute reference

* `name` - (Computed)[string] The name of the security rule.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
