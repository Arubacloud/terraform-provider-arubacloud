terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = ">= 0.0.1"
    }
  }
}

provider "arubacloud" {
  api_key    = "YOUR_API_KEY"
  api_secret = "YOUR_API_SECRET"
}
