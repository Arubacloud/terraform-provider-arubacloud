---
subcategory: "Schedule"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_schedulejob"
sidebar_current: "docs-resource-schedulejob"
description: |-
  Schedule Job allows you to automate tasks and operations in ArubaCloud resources.
---

# arubacloud_schedulejob

Schedule Job enables automated execution of operations on resources at defined times.

## Usage example

```hcl
resource "arubacloud_schedulejob" "example" {
  name        = "daily-backup"
  project_id  = arubacloud_project.example.id
  resource_id = arubacloud_blockstorage.example.id
  action      = "backup"
  schedule    = "0 2 * * *"
  enabled     = true
  tags        = ["automation", "backup"]
}
```

## Argument reference

* `name` - (Required)[string] The name of the schedule job.
* `project_id` - (Required)[string] The project ID.
* `resource_id` - (Required)[string] The target resource ID.
* `action` - (Required)[string] The action to perform (e.g., "backup").
* `schedule` - (Required)[string] Cron expression for the schedule.
* `enabled` - (Optional)[bool] Whether the job is enabled.
* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the schedule job.

## Import

To import a schedule job, define an empty resource in your plan:

```
resource "arubacloud_schedulejob" "example" {
}
```

Import using the schedule job ID:

```
terraform import arubacloud_schedulejob.example <schedulejob_id>
```
