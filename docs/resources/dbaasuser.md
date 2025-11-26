page_title: "arubacloud_dbaas_user Resource - ArubaCloud"
# arubacloud_dbaas_user (Resource)
```terraform
resource "arubacloud_dbaas_user" "basic" {
  name = "basic-dbaas-user"
  dbaas_id = "dbaas-id"
}
```
  Manages an ArubaCloud DBaaS User.
---

# arubacloud_dbaasuser (Resource)

Manages an ArubaCloud DBaaS User.

## Example Usage

```terraform
resource "arubacloud_dbaasuser" "example" {
  dbaas_id = "example-dbaas-id"
  username = "example-user"
  password = "example-password"
}
```


## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_dbaasuser.example <dbaasuser_id>
```
