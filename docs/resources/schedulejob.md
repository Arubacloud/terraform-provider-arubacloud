---
subcategory: "Automation"
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
  name       = "example-job"
  project_id = "example-project-id"
  location   = "eu-central-1"
  tags       = ["nightly", "backup"]
  properties = {
    enabled           = true
    schedule_job_type = "OneShot"              # or "Recurring"
    schedule_at       = "2025-12-01T02:00:00Z" # Only for OneShot
    execute_until     = null                   # Only for Recurring
    cron              = null                   # Only for Recurring
    steps = [
      {
        name         = "Backup Step"
        resource_uri = ""
        action_uri   = ""
        http_verb    = "POST"
        body         = "{\"example\":\"example\"}"
      }
    ]
  }
}
```

## Argument reference

* `name` - (Required)[string] The name of the schedule job.
* `project_id` - (Required)[string] The project ID.
* `location` - (Required)[string] The location for the job.
* `tags` - (Optional)[list(string)] Tags for the job.
* `properties` - (Required)[object] Schedule job properties:
  * `enabled` - (Optional)[bool] Whether the job is enabled.
  * `schedule_job_type` - (Required)[string] Type of job ("OneShot", "Recurring").
  * `schedule_at` - (Optional)[string] Date and time when the job should run (for OneShot).
  * `execute_until` - (Optional)[string] End date until which the job can run (for Recurring).
  * `cron` - (Optional)[string] CRON expression for recurrence (for Recurring).
  * `steps` - (Optional)[list(object)] Steps to execute:
    * `name` - (Optional)[string] Descriptive name of the step.
    * `resource_uri` - (Required)[string] URI of the resource.
    * `action_uri` - (Required)[string] URI of the action to execute.
    * `http_verb` - (Required)[string] HTTP verb to use (GET, POST, etc.).
    * `body` - (Optional)[string] Optional HTTP request body.

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
