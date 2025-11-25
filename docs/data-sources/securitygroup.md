---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_securitygroup"
sidebar_current: "docs-datasource-securitygroup"
description: |-
  Data source for querying Security Group resources in ArubaCloud.
---

# arubacloud_securitygroup (Data Source)

Use this data source to retrieve information about a Security Group resource.

## Usage example

```hcl
data "arubacloud_securitygroup" "example" {
  id = "securitygroup-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the Security Group to query.

## Attribute reference

* `name` - [string] The name of the Security Group.
* `location` - [string] The location of the Security Group.
* `tags` - [list(string)] List of tags for the Security Group.
* `project_id` - [string] The project ID.
* `vpc_id` - [string] The VPC ID.
