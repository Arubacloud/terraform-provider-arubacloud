---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_securityrule"
sidebar_current: "docs-datasource-securityrule"
description: |-
  Data source for querying Security Rule resources in ArubaCloud.
---

# arubacloud_securityrule (Data Source)

Use this data source to retrieve information about a Security Rule resource.

## Usage example

```hcl
data "arubacloud_securityrule" "example" {
  id = "securityrule-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the Security Rule to query.

## Attribute reference

* `name` - [string] The name of the Security Rule.
* `location` - [string] The location of the Security Rule.
* `project_id` - [string] The project ID.
* `vpc_id` - [string] The VPC ID.
* `security_group_id` - [string] The Security Group ID.
* `properties` - [object]
  * `direction` - [string] Direction of the rule (Ingress/Egress).
  * `protocol` - [string] Protocol (ANY, TCP, UDP, ICMP).
  * `port` - [string] Port or port range (for TCP/UDP).
  * `target` - [object]
  * `kind` - [string] Type of the target (Ip/SecurityGroup).
  * `value` - [string] Value of the target (CIDR or SecurityGroup URI).
