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
  id = "vpc-id"
}

output "vpc_name" {
  value = data.arubacloud_vpc.basic.name
}
output "vpc_location" {
  value = data.arubacloud_vpc.basic.location
}
output "vpc_project_id" {
  value = data.arubacloud_vpc.basic.project_id
}
output "vpc_tags" {
  value = data.arubacloud_vpc.basic.tags
}
```

## Argument Reference

<!-- tfplugindocs injects arguments -->

## Attribute Reference

<!-- tfplugindocs injects attributes -->
