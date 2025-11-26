---
page_title: "arubacloud_schedulejob Resource - ArubaCloud"
subcategory: "Job"
description: |-
  Manages an ArubaCloud Schedule Job.
---

# arubacloud_schedulejob (Resource)

Manages an ArubaCloud Schedule Job.

## Example Usage

```terraform
resource "arubacloud_schedulejob" "example" {
  name       = "example-schedule-job"
  project_id = "example-project"
  location   = "example-location"
  tags       = ["tag1", "tag2"]
  properties = {
    enabled            = true
    schedule_job_type  = "Recurring"
    schedule_at        = "2025-11-26T10:00:00Z"
    execute_until      = "2025-12-01T10:00:00Z"
  }
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_schedulejob.example <schedulejob_id>
```
