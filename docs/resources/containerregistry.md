---
subcategory: "Container"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_containerregistry"
sidebar_current: "docs-resource-containerregistry"
description: |-
  Container Registry provides a secure, scalable registry for storing and managing container images.
---

# arubacloud_containerregistry

Container Registries allow you to store, manage, and share container images for your deployments.

## Usage example

```hcl
resource "arubacloud_containerregistry" "example" {
  name              = "example-registry"
  location          = "ITBG-Bergamo"
  tags              = ["container", "test"]
  project_id        = arubacloud_project.example.id
  elasticip_id      = arubacloud_elasticip.example.id
  subnet_id         = arubacloud_subnet.example.id
  security_group_id = arubacloud_securitygroup.example.id
  block_storage_id  = arubacloud_blockstorage.example.id
  billing_period    = "Hour"
  admin_user        = "adminuser"
}
```

## Argument reference

* `name` - (Required)[string] The name of the container registry.
* `location` - (Required)[string] The location for the registry.
* `tags` - (Optional)[list(string)] Tags for the registry.
* `project_id` - (Required)[string] The project ID.
* `elasticip_id` - (Required)[string] Elastic IP ID.
* `subnet_id` - (Required)[string] Subnet ID.
* `security_group_id` - (Required)[string] Security Group ID.
* `block_storage_id` - (Required)[string] Block Storage ID.
* `billing_period` - (Required)[string] Billing period.
* `admin_user` - (Required)[string] Admin user for the registry.

## Attribute reference

* `id` - (Computed)[string] The ID of the container registry.

## Import

To import a container registry, define an empty resource in your plan:

```
resource "arubacloud_containerregistry" "example" {
}
```

Import using the container registry ID:

```
terraform import arubacloud_containerregistry.example <containerregistry_id>
```
