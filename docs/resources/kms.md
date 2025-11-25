---
subcategory: "Security"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_kms"
sidebar_current: "docs-resource-kms"
description: |-
  Key Management Service (KMS) provides secure key storage and management for ArubaCloud resources.
---

# arubacloud_kms

KMS allows you to create, manage, and use cryptographic keys securely.

## Usage example

```hcl
resource "arubacloud_kms" "example" {
  name       = "example-kms"
  project_id = arubacloud_project.example.id
  location   = "ITBG-Bergamo"
  tags       = ["security", "kms"]
  description = "KMS for sensitive workloads"
}
```

## Argument reference

* `name` - (Required)[string] The name of the KMS instance.
* `project_id` - (Required)[string] The project ID.
* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the KMS instance.

## Import

To import a KMS instance, define an empty resource in your plan:

```
resource "arubacloud_kms" "example" {
}
```

Import using the KMS ID:

```
terraform import arubacloud_kms.example <kms_id>
```
