---
page_title: "arubacloud_kaas Resource - terraform-provider-arubacloud"
subcategory: "Container"
description: |-
  Manages an ArubaCloud Kubernetes as a Service (KaaS) cluster.
---

# arubacloud_kaas (Resource)

Manages an ArubaCloud Kubernetes as a Service (KaaS) cluster.

## Example Usage

```terraform
resource "arubacloud_kaas" "example" {
  name           = "my-kaas-cluster"
  location       = "ITBG-Bergamo"
  project_id     = arubacloud_project.example.id
  tags           = ["k8s", "production"]
  billing_period = "Hour"

  network {
    vpc_uri_ref         = arubacloud_vpc.example.uri
    subnet_uri_ref      = arubacloud_subnet.example.uri
    security_group_name = "kaas-security-group"

    node_cidr = {
      address = "10.0.2.0/24"
      name    = "kaas-node-cidr"
    }

    pod_cidr = "10.0.3.0/24"
  }

  settings {
    kubernetes_version = "1.33.2"

    node_pools = [
      {
        name        = "pool-1"
        nodes       = 2
        instance    = "K2A4"
        zone        = "ITBG-1"
        autoscaling = true
        min_count   = 1
        max_count   = 5
      }
    ]

    ha = true
  }
}
```

## Schema

### Required

- `location` (String) KaaS location/region
- `name` (String) KaaS cluster name
- `network` (Block, Required) Network configuration for the cluster (see [below for nested schema](#nestedblock--network))
- `project_id` (String) ID of the project this KaaS resource belongs to
- `settings` (Block, Required) Kubernetes cluster settings (see [below for nested schema](#nestedblock--settings))

### Optional

- `billing_period` (String) Billing period: Hour, Month, or Year
- `tags` (List of String) List of tags for the KaaS resource

### Read-Only

- `id` (String) KaaS identifier
- `uri` (String) KaaS URI

<a id="nestedblock--network"></a>
### Nested Schema for `network`

Required:

- `node_cidr` (Block, Required) Node CIDR configuration (see [below for nested schema](#nestedblock--network--node_cidr))
- `security_group_name` (String) Security group name
- `subnet_uri_ref` (String) Subnet URI reference (e.g., `arubacloud_subnet.example.uri`)
- `vpc_uri_ref` (String) VPC URI reference (e.g., `arubacloud_vpc.example.uri`)

Optional:

- `pod_cidr` (String) Pod CIDR in CIDR notation (e.g., 10.0.3.0/24)

<a id="nestedblock--network--node_cidr"></a>
### Nested Schema for `network.node_cidr`

Required:

- `address` (String) Node CIDR address in CIDR notation (e.g., 10.0.0.0/24)
- `name` (String) Node CIDR name

<a id="nestedblock--settings"></a>
### Nested Schema for `settings`

Required:

- `kubernetes_version` (String) Kubernetes version. Available versions in [ArubaCloud API documentation](https://api.arubacloud.com/docs/metadata#kubernetes-version)
- `node_pools` (Block List, Required) Node pools configuration (see [below for nested schema](#nestedblock--settings--node_pools))

Optional:

- `ha` (Boolean) Enable high availability

<a id="nestedblock--settings--node_pools"></a>
### Nested Schema for `settings.node_pools`

Required:

- `instance` (String) KaaS flavor name. Available flavors in [ArubaCloud API documentation](https://api.arubacloud.com/docs/metadata#kaas-flavors). Example: `K2A4` (2 CPU, 4GB RAM, 40GB storage)
- `name` (String) Node pool name
- `nodes` (Number) Number of nodes in the node pool
- `zone` (String) Datacenter/zone code for nodes

Optional:

- `autoscaling` (Boolean) Enable autoscaling for node pool
- `max_count` (Number) Maximum number of nodes for autoscaling
- `min_count` (Number) Minimum number of nodes for autoscaling

## Import

Aruba Cloud KaaS can be imported using the `kaas_id`:

```shell
terraform import arubacloud_kaas.example <kaas_id>
```
