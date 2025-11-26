---
page_title: "arubacloud_schedulejob Data Source - ArubaCloud"
subcategory: "Job"
description: |-
  Reads an existing ArubaCloud Schedule Job.
---

# arubacloud_schedulejob (Data Source)

Reads an existing ArubaCloud Schedule Job.

```terraform
data "arubacloud_schedule_job" "example" {
  name       = "example-schedule-job"
  project_id = "example-project"
  cron       = "0 0 * * *"
}
```

## Argument Reference

<!-- tfplugindocs injects arguments -->

## Attribute Reference

<!-- tfplugindocs injects attributes -->
