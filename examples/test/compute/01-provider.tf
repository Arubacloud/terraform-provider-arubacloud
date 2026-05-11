terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud" # Uses public registry by default
      version = "~> 0.1.0"              # Use v0.1.x or override with local build via .terraformrc
    }
  }
}


provider "arubacloud" {
  api_key    = var.arubacloud_api_key
  api_secret = var.arubacloud_api_secret
  
  log_level = "DEBUG" # Accepted: OFF, ERROR, WARN, INFO, DEBUG, TRACE. Default: OFF. Requires TF_LOG=DEBUG to surface.

  # Optional: Configure timeout for resource creation (default: 10m)
  resource_timeout = "15m"
}


