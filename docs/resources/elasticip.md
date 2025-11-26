page_title: "arubacloud_elastic_ip Resource - ArubaCloud"
# arubacloud_elastic_ip (Resource)
```terraform
resource "arubacloud_elastic_ip" "basic" {
  name = "basic-elastic-ip"
}
```
  Manages an ArubaCloud Elastic IP.
---

# arubacloud_elasticip (Resource)

Manages an ArubaCloud Elastic IP.

## Example Usage

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
