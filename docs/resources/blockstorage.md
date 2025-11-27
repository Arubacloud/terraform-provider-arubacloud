
---
page_title: "arubacloud_blockstorage Resource - ArubaCloud"
subcategory: "Storage"
description: |-
  Manages an ArubaCloud Block Storage.
---

# arubacloud_blockstorage

Manages an ArubaCloud Block Storage.

```terraform
resource "arubacloud_blockstorage" "example" {
  name       = "example-blockstorage"
  location   = "example-location"
  size       = 100
  project_id = "example-project"
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_blockstorage.example <blockstorage_id>
```
