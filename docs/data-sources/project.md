---
subcategory: "Project"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_project"
sidebar_current: "docs-datasource-project"
description: |-
  Data source for querying Project resources in ArubaCloud.
---

# arubacloud_project (Data Source)

Use this data source to retrieve information about a Project resource.

## Usage example

```hcl
data "arubacloud_project" "example" {
  id = "project-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the project to query.

## Attribute reference

* `name` - (Computed)[string] The name of the project.
* ...other attributes...
