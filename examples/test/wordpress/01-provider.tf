terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = "~> 0.0.1"
    }
  }
}

provider "arubacloud" {
  api_key    = var.arubacloud_api_key
  api_secret = var.arubacloud_api_secret

  resource_timeout = "20m"
}
