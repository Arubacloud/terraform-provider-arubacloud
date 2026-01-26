---
page_title: "arubacloud_schedulejob"
subcategory: "Schedule"
description: |-
  Reads an existing ArubaCloud Schedule Job.
---

# arubacloud_schedulejob

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
