page_title: "arubacloud_elastic_ip Data Source - ArubaCloud"
# arubacloud_elastic_ip (Data Source)
```terraform
data "arubacloud_elastic_ip" "basic" {
  id = "elastic-ip-id"
}
```
  Reads an existing ArubaCloud Elastic IP.
---

# arubacloud_elasticip (Data Source)

Reads an existing ArubaCloud Elastic IP.

```terraform
data "arubacloud_elastic_ip" "example" {
  name       = "example-elastic-ip"
  project_id = "example-project"
  location   = "eu-1"
}
```


## Schema

<no value>
