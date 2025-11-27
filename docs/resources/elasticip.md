---
page_title: "arubacloud_elasticip Resource - ArubaCloud"
subcategory: "Network"
description: |-
  Manages an ArubaCloud ElasticIP.
---

# arubacloud_elasticip

Manages an ArubaCloud ElasticIP.

```terraform
resource "arubacloud_elasticip" "example" {
  name       = "example-elastic-ip"
  location   = "example-location"
  project_id = "example-project"
}
```


## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_elasticip.example <elasticip_id>
```
