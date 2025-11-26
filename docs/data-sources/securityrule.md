---
page_title: "arubacloud_securityrule Data Source - ArubaCloud"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud Security Rule.
---

# arubacloud_securityrule (Data Source)

Reads an existing ArubaCloud Security Rule.

```terraform
data "arubacloud_security_rule" "example" {
  name       = "example-security-rule"
  project_id = "example-project"
  direction  = "inbound"
  protocol   = "tcp"
  port       = 80
}
```

## Argument Reference

<!-- tfplugindocs injects arguments -->

## Attribute Reference

<!-- tfplugindocs injects attributes -->

