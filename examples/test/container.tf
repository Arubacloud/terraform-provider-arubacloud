# KaaS example
resource "arubacloud_kaas" "example" {
  name        = "example-kaas"
  location    = "ITBG-Bergamo"
  tags        = ["k8s", "test"]
  project_id  = arubacloud_project.example.id
  preset      = true
  vpc_id      = arubacloud_vpc.example.id
  subnet_id   = arubacloud_subnet.example.id
  node_cidr {
    address      = "10.0.2.0/24"
    subnet_name  = "kaas-subnet"
  }
  security_group_name = arubacloud_securitygroup.example.name
  version      = "1.32.2"
  node_pools = [
    {
      node_pool_name = "pool-1"
      replicas       = 2
      type           = "c2.medium"
      zone           = "ITBG-1"
    },
    {
      node_pool_name = "pool-2"
      replicas       = 1
      type           = "c2.large"
      zone           = "ITBG-2"
    }
  ]
  ha            = true
  billing_period = "Hour"
}

#container registry example
resource "arubacloud_containerregistry" "example" {
  name               = "example-registry"
  location           = "ITBG-Bergamo"
  tags               = ["container", "test"]
  project_id         = arubacloud_project.example.id
  elasticip_id       = arubacloud_elasticip.example.id
  subnet_id          = arubacloud_subnet.example.id
  security_group_id  = arubacloud_securitygroup.example.id
  block_storage_id   = arubacloud_blockstorage.example.id
  billing_period     = "Hour"
  admin_user         = "adminuser"
}
