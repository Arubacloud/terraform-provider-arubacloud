terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud" # Uses public registry by default
      version = ">= 1.0"                # Use v1.0+ or override with local build via .terraformrc
    }
  }
}

provider "arubacloud" {
  client_id     = var.arubacloud_client_id
  client_secret = var.arubacloud_client_secret
}
