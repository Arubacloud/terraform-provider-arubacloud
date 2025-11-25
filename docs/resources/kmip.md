---
subcategory: "Security"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_kmip"
sidebar_current: "docs-resource-kmip"
description: |-
  Key Management Interoperability Protocol (KMIP) provides standardized key management for ArubaCloud resources.
---

# arubacloud_kmip

KMIP enables secure and standardized management of cryptographic keys across services.

## Usage example

```hcl
resource "arubacloud_kmip" "example" {
  name       = "example-kmip"
  project_id = arubacloud_project.example.id
  location   = "ITBG-Bergamo"
  tags       = ["security", "kmip"]
  description = "KMIP for compliance"
}
```

## Argument reference

* `name` - (Required)[string] The name of the KMIP instance.
* `project_id` - (Required)[string] The project ID.
* `kms_id` - (Required)[string] The ID of the associated KMS instance.

## Attribute reference

* `id` - (Computed)[string] The ID of the KMIP instance.

## Import

To import a KMIP instance, define an empty resource in your plan:

```
resource "arubacloud_kmip" "example" {
}
```

Import using the KMIP ID:

```
terraform import arubacloud_kmip.example <kmip_id>
```
