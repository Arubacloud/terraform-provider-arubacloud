terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = ">= 0.1.0"
    }
  }
}

provider "arubacloud" {
  api_key    = "YOUR_API_KEY"
  api_secret = "YOUR_API_SECRET"
}
