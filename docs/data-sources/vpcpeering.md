---
page_title: "arubacloud_vpc_peering"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud VPC peering.
---

# arubacloud_vpc_peering

Reads an existing ArubaCloud VPC peering.

```terraform
data "arubacloud_vpcpeering" "basic" {
  id         = "vpc-peering-id"
  project_id = "your-project-id"
  vpc_id     = "your-vpc-id"
}

output "vpcpeering_name" {
  value = data.arubacloud_vpcpeering.basic.name
}
```

## Argument Reference

<!-- tfplugindocs injects arguments -->

## Attribute Reference

<!-- tfplugindocs injects attributes -->
