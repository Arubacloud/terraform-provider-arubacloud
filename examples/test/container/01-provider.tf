terraform {
  required_providers {
    arubacloud = {
      source  = "hashicorp.com/arubacloud/arubacloud"
#      version = "~> 1.0"
    }
  }
}

provider "arubacloud" {
  api_key    = var.arubacloud_api_key
  api_secret = var.arubacloud_api_secret

  # Optional: Configure timeout for resource creation (default: 10m)
  resource_timeout = "20m"
}
