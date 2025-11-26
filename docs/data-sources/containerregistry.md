
---
page_title: "arubacloud_containerregistry Data Source - ArubaCloud"
subcategory: "Container"
description: |-
  Retrieves an ArubaCloud Container Registry.
---

# arubacloud_containerregistry (Data Source)

Retrieves an ArubaCloud Container Registry.

```terraform
data "arubacloud_containerregistry" "example" {
  name       = "example-container-registry"
  project_id = "example-project"
  location   = "eu-1"
  type       = "Standard"
}
```

## Schema

<no value>
