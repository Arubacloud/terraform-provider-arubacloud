---
subcategory: "General"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_project"
sidebar_current: "docs-resource-project"
description: |-
  Project is a logical grouping of resources in ArubaCloud.
---

# arubacloud_project

A project allows you to organize and manage resources as a unit.


## Usage example

```hcl
resource "arubacloud_project" "example" {
  name        = "example-project"
  description = "Example ArubaCloud project"
  tags        = ["dev", "test", "terraform"]
}
```

## Argument reference

* `id` - (Computed)[string] The ID of the Project.
* `name` - (Required)[string] The name of the Project.
* ...other arguments...

## Import

To import a Project, define an empty resource in your plan:

```
resource "arubacloud_project" "example" {
}
```

Import using the Project ID:

```
terraform import arubacloud_project.example <project_id>
```
