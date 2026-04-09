---
page_title: "arubacloud_schedulejob"
subcategory: "Schedule"
description: |-
  Reads an existing ArubaCloud Schedule Job.
---

# arubacloud_schedulejob

Reads an existing ArubaCloud Schedule Job.

```terraform
data "arubacloud_schedulejob" "example" {
  id         = "your-schedulejob-id"
  project_id = "your-project-id"
}

output "schedulejob_name" {
  value = data.arubacloud_schedulejob.example.name
}
output "schedulejob_cron" {
  value = data.arubacloud_schedulejob.example.cron
}
```

## Schema

### Arguments

The following arguments are supported:

#### Required

- `id` (String) Schedule Job identifier
- `project_id` (String) ID of the project this Schedule Job belongs to

### Attributes Reference

In addition to all arguments above, the following attributes are exported:

#### Read-Only

- `cron` (String) Cron expression for the schedule
- `description` (String) Schedule Job description
- `name` (String) Schedule Job name
