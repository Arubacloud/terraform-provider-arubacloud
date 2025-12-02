---
page_title: "arubacloud_dbaasuser"
subcategory: "Database"
description: |-
  Retrieves an ArubaCloud DBaaS User.
---

# arubacloud_dbaasuser

Reads an existing ArubaCloud DBaaS user.

```terraform
data "arubacloud_dbaas_user" "example" {
  name       = "example-dbaas-user"
  project_id = "example-project"
  database   = "example-db"
  role       = "admin"
}
```


## Schema

<no value>
