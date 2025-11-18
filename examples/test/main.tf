terraform {
    required_providers {
      arubacloud = {
        source  = "hashicorp/arubacloud"
      }
    }
    
}

provider "arubacloud" {
  api_key    = "your_api_key"
}