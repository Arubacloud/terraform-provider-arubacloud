terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = ">= 0.5"
    }
  }
}

provider "arubacloud" {
  client_id     = var.arubacloud_client_id
  client_secret = var.arubacloud_client_secret

  resource_timeout = "20m"
}
