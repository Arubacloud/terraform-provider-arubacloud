# Copyright (c) HashiCorp, Inc.


## Elastic IP
resource "arubacloud_elasticip" "example" {
  name           = "example-elasticip"
  location       = "ITBG-Bergamo"
  tags           = ["public", "test"]
  billing_period = "hourly"
  project_id     = arubacloud_project.example.id
}


## VPC 
resource "arubacloud_vpc" "example" {
  name     = "example-vpc"
  location = "ITBG-Bergamo"
  tags     = ["network", "test"]
}

resource "arubacloud_vpc" "remote" {
  name     = "remote-vpc"
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

resource "arubacloud_subnet" "example2" {
  name       = "example-subnet2"
  location   = "ITBG-Bergamo"
  tags       = ["subnet", "test"]
  project_id = arubacloud_project.example.id
  vpc_id     = arubacloud_vpc.example.id
  type       = "Advanced"
  network = {
    address = "10.0.2.0/24"
  }
  dhcp = {
    enabled = true
    range = {
      start = "10.0.2.10"
      count = 20
    }
  }
  routes = [
    {
      address = "0.0.0.0"
      gateway = "10.0.2.1"
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

resource "arubacloud_securitygroup" "example2" {
  name       = "example-security-group2"
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

## VPC Peering
resource "arubacloud_vpcpeering" "example" {
  name     = "example-vpc-peering"
  location = "ITBG-Bergamo"
  tags     = ["peering", "prod"]
  peer_vpc = arubacloud_vpc.remote.id
}

## VPC PeeringRoute
resource "arubacloud_vpcpeeringroute" "example" {
  name                   = "example-vpc-peering-route"
  location               = "ITBG-Bergamo"
  tags                   = ["route", "prod"]
  local_network_address  = "10.0.0.0/24"
  remote_network_address = "192.168.1.0/24"
  billing_period         = "Hour"
  vpc_peering_id         = arubacloud_vpcpeering.example.id
}

## VPN Tunnel
resource "arubacloud_vpntunnel" "example" {
  name       = "example-vpn-tunnel"
  location   = "ITBG-Bergamo"
  tags       = ["vpn", "prod"]
  project_id = arubacloud_project.example.id
  properties = {
    vpn_type            = "Site-To-Site"
    vpn_client_protocol = "ikev2"
    ip_configurations = {
      vpc = {
        id = arubacloud_vpc.example.id
      }
      subnet = {
        id = arubacloud_subnet.example.id
      }
      public_ip = {
        id = arubacloud_elasticip.example.id
      }
    }
    vpn_client_settings = {
      ike = {
        lifetime     = 3600
        encryption   = "AES256"
        hash         = "SHA256"
        dh_group     = "group14"
        dpd_action   = "restart"
        dpd_interval = 30
        dpd_timeout  = 120
      }
      esp = {
        lifetime   = 3600
        encryption = "AES256"
        hash       = "SHA256"
        pfs        = "group14"
      }
      psk = {
        cloud_site   = "cloud-site-1"
        on_prem_site = "on-prem-site-1"
        secret       = "supersecretkey"
      }
    }
    peer_client_public_ip = "203.0.113.1"
  }
}

## VPN Route
resource "arubacloud_vpnroute" "example" {
  name          = "example-vpn-route"
  location      = "ITBG-Bergamo"
  tags          = ["route", "vpn"]
  project_id    = arubacloud_project.example.id
  vpn_tunnel_id = arubacloud_vpntunnel.example.id
  properties = {
    cloud_subnet   = "10.0.0.0/24"
    on_prem_subnet = "192.168.1.0/24"
  }
}



