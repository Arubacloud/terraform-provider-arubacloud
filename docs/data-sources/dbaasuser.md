

# arubacloud_dbaasuser (Data Source)

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
