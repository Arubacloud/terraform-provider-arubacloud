terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud" # Uses public registry by default
      version = ">= 0.5"                # Use v0.5+ or override with local build via .terraformrc
    }
  }
}


provider "arubacloud" {
  client_id     = var.arubacloud_client_id
  client_secret = var.arubacloud_client_secret
  
  log_level = "DEBUG" # Accepted: OFF, ERROR, WARN, INFO, DEBUG, TRACE. Default: OFF. Requires TF_LOG=DEBUG to surface.

  # Optional: Configure timeout for resource creation (default: 10m)
  resource_timeout = "15m"
}


