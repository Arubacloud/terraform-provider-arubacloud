## VPC 
resource "arubacloud_vpc" "example" {
  name     = "example-vpc"
  location = "ITBG-Bergamo"
  tags     = ["network", "test"]
}

## Subnet
resource "arubacloud_subnet" "example" {
  name       = "example-subnet"
  location   = "ITBG-Bergamo"
  tags       = ["subnet", "test"]
  project_id = arubacloud_project.example.id
  vpc_id     = arubacloud_vpc.example.id
  type       = "Advanced"
  network = {
    address = "10.0.1.0/24"
  }
  dhcp = {
    enabled = true
    range = {
      start = "10.0.1.10"
      count = 20
    }
  }
  routes = [
    {
      address = "0.0.0.0"
      gateway = "10.0.1.1"
    }
  ]
  dns = ["8.8.8.8", "8.8.4.4"]
}

## Security Group
resource "arubacloud_securitygroup" "example" {
  name       = "example-security-group"
  location   = "ITBG-Bergamo"
  tags       = ["web", "prod"]
  project_id = arubacloud_project.example.id
  vpc_id     = arubacloud_vpc.example.id
}

## Security Rule
resource "arubacloud_securityrule" "example" {
  name              = "example-security-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.example.id
  vpc_id            = arubacloud_vpc.example.id
  security_group_id = arubacloud_securitygroup.example.id
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "80"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

## Elastic IP
resource "arubacloud_elasticip" "example" {
  name           = "example-elasticip"
  location       = "ITBG-Bergamo"
  tags           = ["public", "test"]
  billing_period = "hourly"
  project_id     = arubacloud_project.example.id
}
