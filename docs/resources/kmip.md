---
page_title: "arubacloud_kmip"
subcategory: "Security"
description: |-
  Manages an ArubaCloud Kmip.
---

---
page_title: "arubacloud_kmip"
subcategory: "Security"
description: |-
  Manages an ArubaCloud KMIP resource.
---

```terraform
resource "arubacloud_kmip" "basic" {
  name = "basic-kmip"
}
```

<no value>

## Import

Aruba Cloud Kmip can be imported using the `kmip_id`.

```shell
terraform import arubacloud_kmip.example <kmip_id>
```
