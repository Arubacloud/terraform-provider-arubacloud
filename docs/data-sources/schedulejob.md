---
subcategory: "Schedule"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_schedulejob"
sidebar_current: "docs-datasource-schedulejob"
description: |-
  Data source for querying Schedule Job resources in ArubaCloud.
---

# arubacloud_schedulejob (Data Source)

Use this data source to retrieve information about a Schedule Job resource.

## Usage example

```hcl
data "arubacloud_schedulejob" "example" {
  id = "schedulejob-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the schedule job to query.

## Attribute reference

* `name` - (Computed)[string] The name of the schedule job.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
