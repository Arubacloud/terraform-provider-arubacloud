
---
page_title: "arubacloud_containerregistry Resource - ArubaCloud"
subcategory: "Container"
description: |-
  Manages an ArubaCloud Container Registry.
---

# arubacloud_containerregistry (Resource)

Manages an ArubaCloud Container Registry.

```terraform
resource "arubacloud_containerregistry" "example" {
  name       = "example-container-registry"
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
terraform import arubacloud_containerregistry.example <containerregistry_id>
```
