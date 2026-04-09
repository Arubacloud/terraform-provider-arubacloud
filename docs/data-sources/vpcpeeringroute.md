---
page_title: "arubacloud_vpc_peering_route"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud VPC peering route.
---

# arubacloud_vpc_peering_route

Reads an existing ArubaCloud VPC peering route.

```terraform
data "arubacloud_vpcpeeringroute" "basic" {
  id             = "route-name"
  project_id     = "your-project-id"
  vpc_id         = "your-vpc-id"
  vpc_peering_id = "your-vpc-peering-id"
}

output "vpcpeeringroute_name" {
  value = data.arubacloud_vpcpeeringroute.basic.name
}
```

## Argument Reference

<!-- tfplugindocs injects arguments -->

## Attribute Reference

<!-- tfplugindocs injects attributes -->
