---
page_title: "arubacloud_vpc"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud VPC.
---

# arubacloud_vpc

Reads an existing ArubaCloud VPC.

```terraform
data "arubacloud_vpc" "basic" {
  id         = "vpc-id"
  project_id = "your-project-id"
}

output "vpc_name" {
  value = data.arubacloud_vpc.basic.name
}
output "vpc_location" {
  value = data.arubacloud_vpc.basic.location
}
output "vpc_tags" {
  value = data.arubacloud_vpc.basic.tags
}
```

## Schema

### Arguments

The following arguments are supported:

#### Required

- `id` (String) VPC identifier
- `project_id` (String) ID of the project this VPC belongs to

### Attributes Reference

In addition to all arguments above, the following attributes are exported:

#### Read-Only

- `location` (String) VPC location
- `name` (String) VPC name
- `tags` (List of String) List of tags for the VPC
