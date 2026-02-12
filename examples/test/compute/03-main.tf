terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud" # Uses public registry by default
      version = "~> 0.0.1"              # Use v0.0.1 or override with local build via .terraformrc
    }
  }
}